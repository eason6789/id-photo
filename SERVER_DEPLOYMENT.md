# ID Photo (证件照生成) - 服务器部署指南

## 访问地址

- **证件照工具**: https://tuteng3.site/claw/photo/
- **API 接口**: https://tuteng3.site/claw/api/data/ (结果文件)

## 服务器配置

- **前端路径**: `/var/www/id-photo/`
- **后端服务**: `photo-service` (Go 语言，监听 8080)
- **服务目录**: `/root/photo-service/`
- **结果存储**: `/var/www/html/claw/api/data/results/`
- **数据库**: SQLite `/var/www/html/claw/api/data/results.db`

## 组件说明

### 前端

```bash
# 上传前端文件
cd /Users/easonlv/Code/backend/id-photo
scp -r frontend/. root@119.29.178.222:/var/www/id-photo/
```

### 后端服务

```bash
# 后端已编译的二进制文件在服务器上
# 位置: /root/photo-service/photo-service

# 查看服务状态
ssh root@119.29.178.222 "ps aux | grep photo-service | grep -v grep"

# 重启服务
ssh root@119.29.178.222 "systemctl restart photo-service"
```

### API 代理

Nginx 已配置：

```nginx
location /claw/api/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 180s;
    proxy_connect_timeout 10s;
    proxy_send_timeout 60s;
    client_max_body_size 10m;
}

location /claw/api/data/results/ {
    alias /var/www/html/claw/api/data/results/;
    expires 30d;
    add_header Cache-Control "public, immutable";
}
```

## 部署步骤

### 前端部署

```bash
# 1. 本地上传前端文件
cd /Users/easonlv/Code/backend/id-photo
scp -r frontend/. root@119.29.178.222:/var/www/id-photo/

# 2. SSH 到服务器验证
ssh root@119.29.178.222 "ls -la /var/www/id-photo/"
```

### 后端部署

```bash
# ⚠️ 禁止从本地上传二进制！Mac ARM → Linux x86_64 不兼容，会导致 Exec format error
# 必须在服务器上编译：

# 1. 先在本地推送代码
cd /Users/easonlv/Code/backend/id-photo
git push origin main

# 2. SSH 到服务器编译部署
ssh root@119.29.178.222
cd /root/photo-service
git pull origin main

# 3. 编译前先备份当前运行中的二进制
cp photo-service photo-service.prev

# 4. 在服务器上编译
go build -o photo-service .

# 5. 重启服务
systemctl restart photo-service
```

### 完整更新流程

```bash
# === 前端更新 ===
cd /Users/easonlv/Code/backend/id-photo
scp -r frontend/. root@119.29.178.222:/var/www/id-photo/

# === 后端更新 ===
# 1. 推送代码
git push origin main

# 2. 服务器上拉取、编译、重启
ssh root@119.29.178.222 "cd /root/photo-service && \
  git pull origin main && \
  cp photo-service photo-service.prev && \
  go build -o photo-service . && \
  systemctl restart photo-service"

# 3. 重载 Nginx（可选）
ssh root@119.29.178.222 "nginx -s reload"
```

## 维护命令

### 查看服务状态

```bash
ssh root@119.29.178.222 "ps aux | grep photo-service | grep -v grep"
```

### 查看服务日志

```bash
ssh root@119.29.178.222 "journalctl -u photo-service -f"
```

### 重启服务

```bash
ssh root@119.29.178.222 "systemctl restart photo-service && systemctl status photo-service"
```

### 查看端口

```bash
ssh root@119.29.178.222 "netstat -tlnp | grep 8080"
```

## 注意事项

1. **前端文件大小限制**: Nginx 配置了 10MB 的请求体大小限制
2. **结果文件缓存**: 结果文件设置了 30 天缓存
3. **超时设置**: API 代理配置了较长的超时时间（180s）以支持大文件处理
4. **数据库**: SQLite 数据库存储在 `/var/www/html/claw/api/data/results.db`

---

## 环境变量 & 配置文件

### Go 后端 (photo-service)

Go 服务通过 systemd 管理，环境变量在 `/etc/systemd/system/photo-service.service`：

```ini
Environment=DASHSCOPE_API_KEY=sk-xxx
Environment=COS_SECRET_ID=AKIDxxx
Environment=COS_SECRET_KEY=xxx
Environment=ENV=prod
```

配置文件 `/root/photo-service/config/config.prod.json` **必须写真实值**，不能用 `${VAR}` 占位符（Go 代码直接 `json.Unmarshal`，不会展开环境变量）。

### Python AI 服务 (ai_photo.py)

环境变量在 `/etc/systemd/system/ai-photo.service`：

```ini
Environment=DASHSCOPE_API_KEY=sk-xxx
Environment=COS_SECRET_ID=AKIDxxx
Environment=COS_SECRET_KEY=xxx
Environment=COS_BUCKET=single-az-1251416377
Environment=COS_REGION=ap-guangzhou
```

### 更换密钥时的操作

```bash
# 1. 更新 systemd 环境变量
vim /etc/systemd/system/photo-service.service   # 修改 Environment= 行
vim /etc/systemd/system/ai-photo.service         # 修改 Environment= 行

# 2. 更新 Go 服务配置 JSON（真实值）
vim /root/photo-service/config/config.prod.json

# 3. 重载并重启
systemctl daemon-reload
systemctl restart photo-service
systemctl restart ai-photo
```

### 从崩溃恢复

如果 `photo-service` 无法启动（`Exec format error`），说明二进制文件被错误覆盖。服务器上有备份：

```bash
# 检查文件格式
file /root/photo-service/photo-service
# 正确输出: ELF 64-bit LSB executable, x86-64
# 错误输出: Mach-O 64-bit arm64 executable (macOS 二进制！)

# 恢复备份
cp /root/photo-service/photo-service.prev /root/photo-service/photo-service
systemctl restart photo-service
```

关键规则：**永远在服务器上编译 Go 代码，禁止从本地上传二进制。**

---

*最后更新: 2026-05-28*