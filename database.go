package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	
	"net/http"
	_ "github.com/go-sql-driver/mysql"
)

type Tokens map[string]*Token
var tokensLock sync.RWMutex
var db *sql.DB

func initDB() error {
	dsn := "root:OpenClaw2026!@tcp(127.0.0.1:3306)/wish_plan?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping数据库失败: %w", err)
	}
	
	logger.Println("MySQL数据库连接成功")
	return nil
}

// 从MySQL加载所有tokens到内存map(兼容现有代码)
func loadTokensFromDB() Tokens {
	if db == nil {
		return make(Tokens)
	}
	
	rows, err := db.Query(`
		SELECT token, max_uses, used_count, 
		       UNIX_TIMESTAMP(expires_at) as expires_at,
		       UNIX_TIMESTAMP(created_at) as created_at,
		       order_id, status
		FROM tokens 
		WHERE status = 1
	`)
	if err != nil {
		logger.Printf("查询tokens失败: %v", err)
		return make(Tokens)
	}
	defer rows.Close()
	
	tokens := make(Tokens)
	for rows.Next() {
		var token string
		var maxUses, usedCount int
		var expiresAt, createdAt sql.NullInt64
		var orderID sql.NullString
		var status int
		
		if err := rows.Scan(&token, &maxUses, &usedCount, &expiresAt, &createdAt, &orderID, &status); err != nil {
			logger.Printf("扫描token行失败: %v", err)
			continue
		}
		
		t := &Token{
			MaxUses:   maxUses,
			UsedCount: usedCount,
			Remaining: maxUses - usedCount,
			Disabled:  false,
		}
		
		if createdAt.Valid {
			t.CreatedAt = createdAt.Int64
		}
		if expiresAt.Valid {
			t.ExpiresAt = expiresAt.Int64
		}
		if orderID.Valid {
			t.OrderID = orderID.String
		}
		
		tokens[token] = t
	}
	
	return tokens
}

// 保存tokens到MySQL(替换JSON文件)
func saveTokensToDB(tokens Tokens) {
	if db == nil {
		return
	}
	
	tokensLock.Lock()
	defer tokensLock.Unlock()
	
	for token, t := range tokens {
		var expiresAt interface{}
		if t.ExpiresAt > 0 {
			expiresAt = time.Unix(t.ExpiresAt, 0)
		}
		
		status := 1
		if t.Disabled || t.Remaining <= 0 {
			status = 2
		}
		
		_, err := db.Exec(`
			INSERT INTO tokens (token, order_id, max_uses, used_count, expires_at, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, FROM_UNIXTIME(?))
			ON DUPLICATE KEY UPDATE
				used_count = VALUES(used_count),
				status = VALUES(status)
		`, token, nullString(t.OrderID), t.MaxUses, t.UsedCount, expiresAt, status, t.CreatedAt)
		
		if err != nil {
			logger.Printf("保存token %s失败: %v", token, err)
		}
	}
}

// 验证token(从MySQL读取最新状态)
func verifyTokenFromDB(tokenStr string) (*Token, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	
	var maxUses, usedCount int
	var expiresAt sql.NullTime
	var createdAt time.Time
	var orderID sql.NullString
	var status int
	
	err := db.QueryRow(`
		SELECT max_uses, used_count, expires_at, created_at, order_id, status
		FROM tokens
		WHERE token = ?
	`, tokenStr).Scan(&maxUses, &usedCount, &expiresAt, &createdAt, &orderID, &status)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询token失败: %w", err)
	}
	
	// 检查状态
	if status != 1 {
		return nil, fmt.Errorf("token已禁用或耗尽")
	}
	
	// 检查过期
	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return nil, fmt.Errorf("token已过期")
	}
	
	// 检查次数
	if usedCount >= maxUses {
		return nil, fmt.Errorf("token次数已用完")
	}
	
	t := &Token{
		MaxUses:   maxUses,
		UsedCount: usedCount,
		Remaining: maxUses - usedCount,
		CreatedAt: createdAt.Unix(),
		Disabled:  false,
	}
	
	if expiresAt.Valid {
		t.ExpiresAt = expiresAt.Time.Unix()
	}
	if orderID.Valid {
		t.OrderID = orderID.String
	}
	
	return t, nil
}

// 消费token(原子操作)
func consumeTokenInDB(tokenStr string) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	
	result, err := db.Exec(`
		UPDATE tokens 
		SET used_count = used_count + 1,
		    status = CASE WHEN used_count + 1 >= max_uses THEN 2 ELSE status END
		WHERE token = ? 
		  AND status = 1 
		  AND used_count < max_uses
		  AND (expires_at IS NULL OR expires_at > NOW())
	`, tokenStr)
	
	if err != nil {
		return fmt.Errorf("更新token失败: %w", err)
	}
	
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("token无效、已过期或次数已用完")
	}
	
	return nil
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Webhook接口: 创建订单和token
func webhookOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonResp(w, false, "只支持POST", nil)
		return
	}
	
	var req struct {
		OrderID      string  `json:"order_id"`
		ProductType  string  `json:"product_type"`
		CustomerName string  `json:"customer_name"`
		CustomerPhone string `json:"customer_phone"`
		CustomerEmail string `json:"customer_email"`
		Amount       float64 `json:"amount"`
		MaxUses      int     `json:"max_uses"`
		ExpiresHours int     `json:"expires_hours"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, false, "参数错误", nil)
		return
	}
	
	if req.OrderID == "" {
		jsonResp(w, false, "order_id不能为空", nil)
		return
	}
	
	if req.MaxUses <= 0 {
		req.MaxUses = 1
	}
	if req.ExpiresHours <= 0 {
		req.ExpiresHours = 72
	}
	
	// 幂等检查
	var existingToken string
	err := db.QueryRow("SELECT token FROM orders WHERE order_id = ?", req.OrderID).Scan(&existingToken)
	if err == nil {
		// 订单已存在,返回现有token
		photoURL := fmt.Sprintf("https://tuteng3.site/claw/photo/?token=%s", existingToken)
		jsonResp(w, true, "订单已存在", map[string]interface{}{
			"order_id":  req.OrderID,
			"token":     existingToken,
			"photo_url": photoURL,
		})
		return
	}
	
	// 创建新token
	tokenStr := generateTokenID()
	expiresAt := time.Now().Add(time.Duration(req.ExpiresHours) * time.Hour)
	
	tx, err := db.Begin()
	if err != nil {
		jsonResp(w, false, "数据库错误", nil)
		return
	}
	defer tx.Rollback()
	
	// 插入token
	result, err := tx.Exec(`
		INSERT INTO tokens (token, order_id, max_uses, used_count, expires_at, status, note)
		VALUES (?, ?, ?, 0, ?, 1, '通过webhook创建')
	`, tokenStr, req.OrderID, req.MaxUses, expiresAt)
	if err != nil {
		jsonResp(w, false, "创建token失败", nil)
		return
	}
	
	tokenID, _ := result.LastInsertId()
	
	// 插入订单
	rawPayload, _ := json.Marshal(req)
	_, err = tx.Exec(`
		INSERT INTO orders (order_id, product_type, customer_name, customer_phone, customer_email, amount, token_id, status, raw_payload, webhook_ip)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`, req.OrderID, req.ProductType, req.CustomerName, req.CustomerPhone, req.CustomerEmail, req.Amount, tokenID, string(rawPayload), r.RemoteAddr)
	if err != nil {
		jsonResp(w, false, "创建订单失败", nil)
		return
	}
	
	if err := tx.Commit(); err != nil {
		jsonResp(w, false, "提交失败", nil)
		return
	}
	
	photoURL := fmt.Sprintf("https://tuteng3.site/claw/photo/?token=%s", tokenStr)
	
	jsonResp(w, true, "ok", map[string]interface{}{
		"order_id":   req.OrderID,
		"token":      tokenStr,
		"expires_at": expiresAt.Format(time.RFC3339),
		"photo_url":  photoURL,
	})
	
	logger.Printf("Webhook创建订单成功: %s, token: %s", req.OrderID, tokenStr)
}
