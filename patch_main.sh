#!/bin/bash
# 修改 main.go 以使用 MySQL

# 1. 在 main 函数开头添加 initDB()
sed -i '/func main() {/a\t// 初始化数据库\n\tif err := initDB(); err != nil {\n\t\tlog.Fatalf("数据库初始化失败: %v", err)\n\t}' main.go

# 2. 添加 webhook 路由
sed -i '/http.HandleFunc("/api\/generate-photo", generatePhoto)/a\thttp.HandleFunc("/api/webhook/order", webhookOrderHandler)' main.go

# 3. 修改 loadTokens 调用为 loadTokensFromDB
sed -i 's/loadTokens()/loadTokensFromDB()/g' main.go

# 4. 修改 saveTokens 调用为 saveTokensToDB
sed -i 's/saveTokens(/saveTokensToDB(/g' main.go

# 5. 在 verifyToken 函数中添加 MySQL 验证逻辑(这个比较复杂,手动处理)

echo "Patch applied!"
