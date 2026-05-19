package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("mysql", cfg.DB)
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
