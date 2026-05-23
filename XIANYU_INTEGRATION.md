# 咸鱼 (Xianyu) Skills 接入指引文档

本文档面向 AI 模型（如 Claude、GPT-4、DeepSeek 等），指导如何将**AI证件照生成系统**集成到咸鱼平台作为一个 Skill。

---

## 📋 前置条件检查

在开始之前，请确认以下信息：

### 1. 服务器信息
- **服务器 IP**: <YOUR_SERVER_IP>
- **SSH 账号**: root
- **SSH 密码**: <REDACTED>
- **SSH 连接命令**: 
  ```bash
  sshpass -p '<REDACTED>' ssh root@<YOUR_SERVER_IP>
  ```

### 2. 服务状态验证
```bash
# 检查 photo-service 服务状态
systemctl status photo-service

# 检查 Nginx 状态
systemctl status nginx

# 检查 MySQL 数据库连接
mysql -u root -p'<REDACTED>' -e "USE wish_plan; SHOW TABLES;"
```

### 3. API 端点测试
```bash
# 测试 Webhook 端点
curl -X POST http://127.0.0.1:8080/api/webhook/order \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"SKILL_TEST_001","max_uses":5}'

# 预期返回包含 token 和 photo_url
```

---

## 🎯 集成目标

将证件照生成功能封装为一个咸鱼 Skill，用户在咸鱼平台下单后：
1. 咸鱼平台调用 Webhook 创建 Token
2. 返回给用户一个访问链接
3. 用户点击链接 → 上传照片 → 生成证件照 → 下载

---

## 🔧 Skill 配置文件模板

### skill.json (咸鱼 Skill 描述文件)

```json
{
  "skill_name": "ai_photo_id_generator",
  "display_name": "AI证件照生成器",
  "version": "1.0.0",
  "description": "基于阿里云AI的智能证件照生成服务，支持11种国际规格（身份证、护照、签证等），自动美化，15-60秒出图",
  "category": "photo_service",
  "tags": ["证件照", "AI美化", "护照照片", "签证照片"],
  "author": "TuTeng",
  "webhook": {
    "order_created": {
      "url": "https://tuteng3.site/api/webhook/order",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      },
      "timeout": 10000
    }
  },
  "pricing": {
    "base_price": 9.90,
    "currency": "CNY",
    "unit": "次",
    "quota": {
      "default": 10,
      "description": "10次生成次数，72小时内有效"
    }
  },
  "features": [
    "支持11种证件照规格（中国护照、身份证、一寸、二寸、日本/美国/申根签证等）",
    "AI智能美化，自动调整光线、肤色、背景",
    "支持男女不同风格",
    "高清输出（300dpi，符合各国证件标准）",
    "蓝/红/白多种背景色可选",
    "15-60秒快速生成"
  ],
  "instructions": {
    "user_guide": "下单后您将收到一个专属链接，点击后上传您的照片，选择需要的证件照规格（如护照、身份证、签证等），点击生成即可。生成完成后可直接下载高清证件照。",
    "photo_requirements": [
      "正面免冠照片",
      "五官清晰可见",
      "建议白色或浅色背景（AI会自动替换）",
      "JPG或PNG格式，不超过10MB"
    ]
  }
}
```

---

## 📡 Webhook 集成步骤

### 步骤 1: 订单创建时调用 Webhook

当用户在咸鱼平台下单购买证件照生成服务时，咸鱼后台应调用：

**端点**: `POST https://tuteng3.site/api/webhook/order`

**请求头**:
```
Content-Type: application/json
```

**请求体**:
```json
{
  "order_id": "XY20260318123456",
  "product_type": "证件照生成服务",
  "customer_name": "张三",
  "customer_phone": "13800138000",
  "customer_email": "user@example.com",
  "amount": 9.90,
  "max_uses": 10,
  "expires_hours": 72
}
```

**字段说明**:
- `order_id` (必填): 咸鱼订单唯一标识，用于幂等控制
- `product_type` (可选): 商品类型描述
- `customer_name` (可选): 客户姓名
- `customer_phone` (可选): 客户手机号
- `customer_email` (可选): 客户邮箱
- `amount` (可选): 订单金额
- `max_uses` (必填): 允许生成的次数，建议 10 次
- `expires_hours` (可选): Token 有效期（小时），默认 72 小时

**响应示例**:
```json
{
  "success": true,
  "message": "ok",
  "data": {
    "order_id": "XY20260318123456",
    "token": "PHOTO_1WI4FP2lNa_1wcMVq9XdfA",
    "expires_at": "2026-03-21T12:00:00+08:00",
    "photo_url": "https://tuteng3.site/claw/photo/?token=PHOTO_1WI4FP2lNa_1wcMVq9XdfA"
  }
}
```

**关键字段**:
- `token`: 生成的访问令牌
- `photo_url`: 用户访问链接，**直接发送给用户**

### 步骤 2: 将 photo_url 发送给用户

咸鱼平台收到 Webhook 响应后，应通过以下方式之一通知用户：

1. **站内消息**: "您的证件照生成服务已开通，请点击链接使用：{photo_url}"
2. **短信通知**: "您购买的证件照服务已激活，访问链接：{photo_url}"
3. **订单详情页**: 显示 "立即使用" 按钮，链接到 photo_url

### 步骤 3: 用户使用流程

用户点击链接后：
1. 自动验证 Token（前端自动读取 URL 参数）
2. 显示剩余生成次数（如 10/10）
3. 上传照片
4. 选择证件照规格（身份证、护照、一寸蓝底等）
5. 选择性别（影响 AI 美化风格）
6. 点击"开始生成"
7. 等待 15-60 秒（显示 Loading 动画）
8. 生成完成，显示结果并提供下载按钮
9. 可继续生成（次数 9/10, 8/10...）

---

## 🔐 幂等性保证

**重要**: 相同 `order_id` 多次调用 Webhook，系统会返回**相同的 token**，不会创建新的 token。

**验证方法**:
```bash
# 第一次调用
curl -X POST https://tuteng3.site/api/webhook/order \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"TEST_IDEMPOTENT","max_uses":5}'
# 返回: {"token": "PHOTO_abc123"}

# 第二次调用（相同 order_id）
curl -X POST https://tuteng3.site/api/webhook/order \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"TEST_IDEMPOTENT","max_uses":5}'
# 返回: {"token": "PHOTO_abc123"}  ← 相同 token
```

**用途**: 防止订单重复提交、网络重试导致生成多个 token。

---

## 🧪 测试清单

### 1. 基础功能测试

```bash
# 1.1 创建订单并获取 Token
curl -s -X POST https://tuteng3.site/api/webhook/order \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"SKILL_TEST_001","max_uses":5}' | python3 -m json.tool

# 预期输出包含: token, photo_url

# 1.2 验证 Token
TOKEN="<从上一步获取的 token>"
curl -s "https://tuteng3.site/api/verify-token?token=$TOKEN" | python3 -m json.tool

# 预期输出: {"success": true, "data": {"remaining": 5, "max_uses": 5}}

# 1.3 访问前端页面
curl -I "https://tuteng3.site/claw/photo/?token=$TOKEN"
# 预期返回: HTTP/2 200
```

### 2. 完整生成流程测试

**前置条件**: 准备一张测试照片 `test.jpg`

```bash
TOKEN="<你的token>"

# 2.1 上传并生成证件照
curl -X POST "https://tuteng3.site/api/generate-photo" \
  -F "token=$TOKEN" \
  -F "image=@test.jpg" \
  -F "spec=cn_one_inch_blue" \
  -F "gender=female"

# 2.2 等待 15-60 秒，查看返回结果
# 预期包含: {"success": true, "data": {"image": "/claw/api/data/results/xxx_result.jpg"}}

# 2.3 下载生成的证件照
RESULT_PATH="<从上一步获取的 image 路径>"
curl -o result.jpg "https://tuteng3.site$RESULT_PATH"

# 2.4 验证图片尺寸
file result.jpg
# 预期: JPEG image data, ... 295x413 (一寸照片像素尺寸)
```

### 3. 边界情况测试

```bash
# 3.1 Token 耗尽测试
# 连续生成 5 次（max_uses=5），第 6 次应失败
for i in {1..6}; do
  curl -X POST "https://tuteng3.site/api/generate-photo" \
    -F "token=$TOKEN" -F "image=@test.jpg" \
    -F "spec=cn_one_inch" -F "gender=female"
  echo "第 $i 次生成"
done
# 第 6 次预期: {"success": false, "message": "token次数已用完"}

# 3.2 过期 Token 测试
# 创建一个 1 小时过期的 token
curl -X POST https://tuteng3.site/api/webhook/order \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"EXPIRE_TEST","max_uses":10,"expires_hours":1}'
# 等待 1 小时后验证，应返回 "token已过期"

# 3.3 无效 Token 测试
curl "https://tuteng3.site/api/verify-token?token=INVALID_TOKEN"
# 预期: {"success": false, "message": "token不存在"}
```

---

## 🐛 故障排查指南

### 问题 1: Webhook 返回 404
**原因**: Nginx 代理配置错误或服务未启动

**排查步骤**:
```bash
# 检查服务状态
sshpass -p '<REDACTED>' ssh root@<YOUR_SERVER_IP> "systemctl status photo-service"

# 检查 Nginx 配置
sshpass -p '<REDACTED>' ssh root@<YOUR_SERVER_IP> "grep -A 5 'location /api/' /etc/nginx/conf.d/tuteng3.site.conf"

# 应该是: proxy_pass http://127.0.0.1:8080;
```

**解决方法**:
```bash
# 如果 proxy_pass 端口错误
sshpass -p '<REDACTED>' ssh root@<YOUR_SERVER_IP> "sed -i 's|proxy_pass http://127.0.0.1:5000|proxy_pass http://127.0.0.1:8080|g' /etc/nginx/conf.d/tuteng3.site.conf && nginx -t && systemctl reload nginx"
```

### 问题 2: 生成超时（180秒无响应）
**原因**: AI 服务调用失败或网络问题

**排查步骤**:
```bash
# 查看服务日志
sshpass -p '<REDACTED>' ssh root@<YOUR_SERVER_IP> "journalctl -u photo-service -n 50 | grep -i 'facechain\|error\|panic'"
```

**可能原因**:
1. DashScope API 密钥失效
2. 阿里云账号欠费
3. 上传的图片不符合要求（非人脸照片）

### 问题 3: 图片尺寸不正确
**症状**: 生成的图片只有几 KB，像素很小（如 25x35 px）

**解决方法**: 已在最新版本修复，确认服务版本
```bash
# 检查代码中是否包含 mm 到 px 的转换
sshpass -p '<REDACTED>' ssh root@<YOUR_SERVER_IP> "grep -A 3 '将毫米转换为像素' /root/photo-service/main.go"

# 应该有: widthPx := int(float64(width) / 25.4 * float64(dpi))
```

### 问题 4: 数据库连接失败
**症状**: 日志显示 "数据库连接失败"

**排查步骤**:
```bash
# 测试数据库连接
sshpass -p '<REDACTED>' ssh root@<YOUR_SERVER_IP> "mysql -u root -p'<REDACTED>' -e 'SELECT 1' 2>&1"

# 检查 MySQL 服务状态
sshpass -p '<REDACTED>' ssh root@<YOUR_SERVER_IP> "systemctl status mysqld"
```

---

## 📊 数据统计与监控

### 查询订单统计

```sql
-- 连接到数据库
mysql -u root -p'<REDACTED>' wish_plan

-- 查询今日订单数
SELECT COUNT(*) as today_orders 
FROM orders 
WHERE DATE(created_at) = CURDATE();

-- 查询今日生成数
SELECT COUNT(*) as today_generations 
FROM generation_logs 
WHERE DATE(created_at) = CURDATE();

-- 查询热门规格
SELECT spec, COUNT(*) as count 
FROM generation_logs 
GROUP BY spec 
ORDER BY count DESC;

-- 查询 Token 使用情况
SELECT 
    COUNT(*) as total_tokens,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as active_tokens,
    SUM(used_count) as total_generations,
    AVG(used_count) as avg_uses_per_token
FROM tokens;
```

### 性能监控

```bash
# CPU 和内存占用
top -b -n 1 | grep photo-service

# 请求响应时间（从 Nginx 日志分析）
tail -100 /var/log/nginx/access.log | grep "/api/generate-photo" | awk '{print $NF}'

# 磁盘空间（结果图片占用）
du -sh /var/www/html/claw/api/data/results/
```

---

## 🔄 更新与维护

### 更新代码

```bash
# 1. SSH 登录服务器
sshpass -p '<REDACTED>' ssh root@<YOUR_SERVER_IP>

# 2. 备份当前版本
cp /root/photo-service/main.go /root/photo-service/main.go.backup.$(date +%Y%m%d%H%M%S)

# 3. 修改代码
vim /root/photo-service/main.go

# 4. 重新编译
cd /root/photo-service
go build -o photo-service .

# 5. 重启服务
systemctl restart photo-service

# 6. 验证服务状态
systemctl status photo-service
journalctl -u photo-service -f
```

### 数据库维护

```sql
-- 清理过期 token（保留 30 天）
DELETE FROM tokens 
WHERE expires_at < DATE_SUB(NOW(), INTERVAL 30 DAY);

-- 清理旧的生成日志（保留 90 天）
DELETE FROM generation_logs 
WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY);

-- 清理旧的订单记录（保留 180 天）
DELETE FROM orders 
WHERE created_at < DATE_SUB(NOW(), INTERVAL 180 DAY);
```

### 清理临时文件

```bash
# 清理上传的临时文件（7天前）
find /var/www/html/claw/api/data/uploads/ -name "*.jpg" -mtime +7 -delete

# 清理结果文件（30天前）
find /var/www/html/claw/api/data/results/ -name "*.jpg" -mtime +30 -delete
```

---

## 🚀 生产环境建议

### 1. 负载均衡
如果访问量大，建议部署多个实例：
```
用户 → Nginx LB → photo-service-1 (8080)
                  → photo-service-2 (8081)
                  → photo-service-3 (8082)
```

### 2. Redis 缓存
Token 验证频繁，可引入 Redis：
```go
// 伪代码
func verifyTokenFromDB(token string) (*Token, error) {
    // 先查 Redis
    if cached := redis.Get("token:" + token); cached != "" {
        return parseCachedToken(cached), nil
    }
    
    // 再查 MySQL
    tok := queryMySQLToken(token)
    
    // 写入 Redis（TTL = token过期时间）
    redis.Set("token:" + token, serialize(tok), ttl)
    
    return tok, nil
}
```

### 3. 异步队列
AI 生成改为异步：
```
用户提交 → 立即返回 task_id → 后台队列处理 → Webhook 通知完成
```

### 4. OSS 存储
结果图片存储到阿里云 OSS，节省服务器空间。

### 5. 监控告警
集成 Prometheus + Grafana：
- API 响应时间
- Token 消费速率
- AI 生成成功率
- 服务器资源使用

---

## 📞 技术支持

### 日志位置
- **应用日志**: `journalctl -u photo-service -f`
- **Nginx 访问日志**: `/var/log/nginx/access.log`
- **Nginx 错误日志**: `/var/log/nginx/error.log`

### 常用命令速查

```bash
# 查看服务状态
systemctl status photo-service

# 重启服务
systemctl restart photo-service

# 实时查看日志
journalctl -u photo-service -f

# 查看最近 100 行日志
journalctl -u photo-service -n 100

# 测试 Webhook
curl -X POST https://tuteng3.site/api/webhook/order \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"TEST_'1773799446\","max_uses":5}'

# 查询数据库
mysql -u root -p'<REDACTED>' wish_plan -e "SELECT * FROM tokens ORDER BY created_at DESC LIMIT 10;"
```

---

## ✅ 集成检查清单

在正式上线前，请确认以下项目：

- [ ] Webhook 端点可访问（返回 200 OK）
- [ ] 幂等性测试通过（相同 order_id 返回相同 token）
- [ ] Token 验证功能正常
- [ ] 完整生成流程测试通过（上传→生成→下载）
- [ ] 边界情况测试（token 耗尽、过期、无效）
- [ ] 前端页面在移动端和桌面端正常显示
- [ ] 生成的图片尺寸正确（295x413px 等）
- [ ] 所有 11 种证件照规格均可正常生成
- [ ] 服务日志正常，无报错
- [ ] 数据库连接正常，数据正确写入
- [ ] systemd 服务已启用自动启动
- [ ] Nginx 超时配置正确（180s）
- [ ] HTTPS 证书有效

---

## 📚 参考资料

- **项目 README**: `/root/photo-service/README.md`
- **数据库 Schema**: `/root/photo-service/schema.sql`
- **前端页面**: `/root/project/portfolio-home/src/claw/photo/index.html`
- **服务配置**: `/etc/systemd/system/photo-service.service`
- **Nginx 配置**: `/etc/nginx/conf.d/tuteng3.site.conf`

---

## 🎓 AI 模型执行建议

如果你是一个 AI 模型（如 Claude、GPT-4、DeepSeek）尝试执行此集成，建议按以下步骤：

1. **阅读理解**: 先通读本文档，理解整体架构
2. **环境验证**: 执行"前置条件检查"中的所有命令
3. **基础测试**: 按照"测试清单"逐项执行
4. **问题排查**: 遇到错误时参考"故障排查指南"
5. **记录结果**: 将测试结果、遇到的问题和解决方法记录下来
6. **反馈改进**: 如果发现文档不清晰或遗漏，提出改进建议

**关键提示**:
- 本系统已部署并运行，**不需要重新安装**
- 主要任务是**测试集成**和**验证功能**
- 使用 SSH 命令时注意密码中的特殊字符需要转义
- Webhook 调用需要使用 HTTPS（https://tuteng3.site）
- 生成过程需要等待 15-60 秒，不要超时中断

---

**最后更新**: 2026-03-18  
**文档版本**: 1.0  
**维护者**: TuTeng  
**联系方式**: 服务器 <YOUR_SERVER_IP>
