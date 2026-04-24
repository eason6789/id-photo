# 证件照生成系统 v2.0

基于阿里百炼通义万相AI生图的证件照生成系统。

## 核心特性

✅ **基于用户照片生成** - 使用用户上传的照片作为参考，保持人物面部特征
✅ **AI正装改造** - 自动为照片添加标准证件照着装（西装/正装）
✅ **标准背景替换** - 替换为证件照标准背景色（白色/灰色等）
✅ **精确尺寸裁剪** - Hivision服务提供精确的像素尺寸和DPI控制
✅ **多国证件支持** - 中国身份证/护照、美国护照/签证、申根签证等

## 系统架构

```
用户上传照片
    ↓
AI生图服务 (8091端口)
    ├─ 阿里百炼wanx2.1-t2i-turbo模型
    ├─ ref_img: 用户上传的照片（保持面部特征）
    ├─ prompt: 证件照规范（背景、着装、姿态）
    └─ 输出: AI生成的正装证件照
    ↓
Hivision服务 (8080端口)
    ├─ 精确裁剪到标准尺寸
    ├─ 调整DPI
    └─ 输出: 最终证件照
    ↓
Go后端服务 (8090端口)
    └─ 返回给前端
```

## 关键说明：图生图模式

**本系统使用阿里百炼的图生图功能（ref_img参数）**：

1. **用户上传的照片作为参考图** (`ref_img`)
2. **AI模型保持人物面部特征**（通过 `ref_strength=0.75`）
3. **仅改变背景、着装、姿态**（通过 prompt 控制）
4. **不是凭空生成**，而是基于用户照片改造

示例prompt:
```
保持照片中人物的面部特征、五官、发型、肤色完全不变。
将照片改造成标准中国身份证证件照格式。
纯白色背景。
人物改穿深藏蓝色正式西装外套，内搭白色衬衫，系深色领带。
调整为正面免冠标准证件照姿态...
```

## 支持的证件类型

| 代码 | 名称 | 尺寸(px) | DPI | 背景色 |
|------|------|----------|-----|--------|
| `cn_id` | 中国身份证 | 358×441 | 350 | 白色 |
| `cn_passport` | 中国护照 | 390×567 | 300 | 白色 |
| `cn_one_inch` | 一寸照片 | 295×413 | 300 | 白色 |
| `cn_two_inch` | 二寸照片 | 413×579 | 300 | 白色 |
| `cn_small_one_inch` | 小一寸 | 260×378 | 300 | 白色 |
| `cn_driver_license` | 驾驶证 | 260×378 | 300 | 白色 |
| `us_passport` | 美国护照 | 600×600 | 300 | 白色 |
| `us_visa` | 美国签证 | 600×600 | 300 | 白色 |
| `schengen_visa` | 申根签证 | 413×531 | 300 | 浅灰色 |
| `uk_passport` | 英国护照 | 413×531 | 300 | 浅灰色 |
| `jp_passport` | 日本护照 | 413×531 | 300 | 白色 |

## 安装部署

### 1. 安装依赖

Python依赖已通过虚拟环境安装：
```bash
cd ~/photo-service
source venv/bin/activate
```

Go依赖：
```bash
cd ~/photo-service
go mod download
```

### 2. 启动服务

**启动AI生图服务**（8091端口）：
```bash
./start_ai_service.sh
# 或
source venv/bin/activate && python3 ai_photo.py
```

**启动Go后端服务**（8090端口）：
```bash
go run main.go
```

### 3. 确保HivisionIDPhotos服务运行（8080端口）

HivisionIDPhotos需要单独部署。

## API使用

### AI生图服务 (8091端口)

**端点**: `POST /ai-idphoto`

**参数**:
- `input_image` (file) - 用户上传的照片
- `spec` (string) - 证件类型代码
- `gender` (string) - 性别 (`male` / `female`)

**示例**:
```bash
curl -X POST http://127.0.0.1:8091/ai-idphoto \
  -F "input_image=@my_photo.jpg" \
  -F "spec=cn_one_inch" \
  -F "gender=male"
```

**返回**:
```json
{
  "success": true,
  "image_base64": "base64编码的图片...",
  "spec_info": {
    "name": "一寸照片",
    "size_px": [295, 413],
    "dpi": 300,
    "bg_color": "#FFFFFF"
  },
  "prompt_used": "保持照片中人物的面部特征..."
}
```

### 完整证件照服务 (8090端口)

**端点**: `POST /generate`

**参数**:
- `photo` (file) - 用户上传的照片
- `spec` (string) - 证件类型代码
- `gender` (string) - 性别

**示例**:
```bash
curl -X POST http://127.0.0.1:8090/generate \
  -F "photo=@my_photo.jpg" \
  -F "spec=cn_id" \
  -F "gender=female"
```

**返回**:
```json
{
  "success": true,
  "image_url": "/download/final_xxx.jpg",
  "image_base64_hd": "base64编码的高清图片...",
  "spec": {
    "name": "中国身份证",
    "size_px": [358, 441],
    "size_mm": [26, 32],
    "dpi": 350,
    "bg_color": "#FFFFFF"
  },
  "ai_used": true
}
```

## 测试

### 测试AI服务

```bash
./test_ai_service.sh your_photo.jpg cn_one_inch male
```

### 健康检查

```bash
# AI服务
curl http://127.0.0.1:8091/health

# 完整服务
curl http://127.0.0.1:8090/health
```

### 查看支持的证件类型

```bash
curl http://127.0.0.1:8091/specs
curl http://127.0.0.1:8090/specs
```

## 注意事项

1. **API Key管理**
   - 主Key: `YOUR_DASHSCOPE_API_KEY`
   - 备用Key: `YOUR_DASHSCOPE_API_KEY_BACKUP`
   - 写在代码中，生产环境建议改为环境变量

2. **生成时间**
   - AI生图约 8-15秒（阿里百炼异步API）
   - Hivision裁剪约 2-5秒
   - 总计约 15-20秒/张

3. **降级策略**
   - AI服务失败时，自动降级为直接使用原图裁剪
   - `ai_used` 字段表示是否使用了AI

4. **图片质量要求**
   - 上传图片需包含清晰可识别的人脸
   - 建议分辨率 ≥ 500px
   - 支持 JPG/PNG 格式

5. **systemd服务**（Linux环境）

   创建 `/etc/systemd/system/ai-photo.service`:
   ```ini
   [Unit]
   Description=AI ID Photo Service
   After=network.target

   [Service]
   Type=simple
   User=your_user
   WorkingDirectory=/path/to/photo-service
   ExecStart=/path/to/photo-service/venv/bin/python3 ai_photo.py
   Restart=always

   [Install]
   WantedBy=multi-user.target
   ```

   启动：
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl start ai-photo
   sudo systemctl enable ai-photo
   ```

## 故障排查

### AI服务连接失败
- 检查服务是否启动: `curl http://127.0.0.1:8091/health`
- 查看日志输出
- 检查API Key是否有效

### 生成图片不符合预期
- 调整 `ref_strength` 参数（当前0.75）
- 修改prompt模板
- 增加negative_prompt内容

### Hivision服务失败
- 确认8080端口服务运行
- 检查图片中是否包含人脸
- 查看Hivision服务日志

## 开发者

- AI生图服务: FastAPI + 阿里百炼SDK
- 后端服务: Go + Gin
- 裁剪服务: HivisionIDPhotos (Python)

## 版本

v2.0 - 集成阿里百炼AI生图，基于用户照片生成正装证件照
