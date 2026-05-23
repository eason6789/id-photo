# ID Photo (证件照生成) - 服务器部署指南

## 访问地址

- **证件照工具**: https://tuteng3.site/claw/photo/
- **API 接口**: https://tuteng3.site/claw/api/data/ (结果文件)

## 服务器配置

- **前端路径**: `/root/project/portfolio-home/src/claw/photo/`
- **后端服务**: `photo-service` (Go 语言，监听 8080)
- **服务目录**: `/root/photo-service/`
- **结果存储**: `/var/www/html/claw/api/data/results/`
- **数据库**: SQLite `/var/www/html/claw/api/data/results.db`

## 组件说明

### 前端

将 `frontend/` 目录下的文件部署到服务器的 `/root/project/portfolio-home/src/claw/photo/` 路径。

### 后端服务

后端二进制文件位于服务器 `/root/photo-service/photo-service`，通过 systemd 管理：

```bash
# 查看服务状态
systemctl status photo-service

# 重启服务
systemctl restart photo-service
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

1. 上传前端文件到服务器目标路径
2. 在服务器上验证文件已就位

### 后端部署

1. 本地编译 Go 项目：`go build -o photo-service main.go`
2. 上传编译好的二进制文件到服务器
3. 重启服务：`systemctl restart photo-service`

### 完整更新流程

1. 更新前端文件并上传
2. 如后端代码有变更，重新编译并上传
3. 重启服务：`systemctl restart photo-service`
4. 重载 Nginx：`nginx -s reload`

## 维护命令

### 查看服务状态

```bash
ps aux | grep photo-service | grep -v grep
```

### 查看服务日志

```bash
journalctl -u photo-service -f
```

### 重启服务

```bash
systemctl restart photo-service && systemctl status photo-service
```

### 查看端口

```bash
netstat -tlnp | grep 8080
```

## 注意事项

1. **前端文件大小限制**: Nginx 配置了 10MB 的请求体大小限制
2. **结果文件缓存**: 结果文件设置了 30 天缓存
3. **超时设置**: API 代理配置了较长的超时时间（180s）以支持大文件处理
4. **数据库**: SQLite 数据库存储在 `/var/www/html/claw/api/data/results.db`

---

*最后更新: 2026-04-25*
