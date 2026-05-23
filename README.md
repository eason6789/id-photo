<div align="center">
  <a href="#english">English</a> | <a href="#chinese">中文</a>
</div>

---

<h2 id="english">AI ID Photo Generator v2.0</h2>

An AI-powered ID photo generation system based on Alibaba Bailian (Tongyi Wanxiang) image generation.

### Core Features

✅ **Photo-based Generation** — Uses user-uploaded photos as reference to preserve facial features
✅ **AI Formal Attire** — Automatically adds standard formal wear (suit/blazer)
✅ **Background Replacement** — Standard ID photo backgrounds (white, gray, etc.)
✅ **Precision Cropping** — Hivision service provides exact pixel dimensions and DPI control
✅ **Multi-Country Support** — Chinese ID/passport, US passport/visa, Schengen visa, and more

### System Architecture

```
User uploads photo
    ↓
AI Image Service (port 8091)
    ├─ Alibaba Bailian wanx2.1-t2i-turbo model
    ├─ ref_img: user's photo (preserves facial features)
    ├─ prompt: ID photo specs (background, attire, pose)
    └─ Output: AI-generated formal ID photo
    ↓
Hivision Service (port 8080)
    ├─ Precision crop to standard dimensions
    ├─ DPI adjustment
    └─ Output: Final ID photo
    ↓
Go Backend (port 8090)
    └─ Returns result to frontend
```

### Supported Photo Types

| Code | Name | Size (px) | DPI | Background |
|------|------|-----------|-----|------------|
| `cn_id` | Chinese ID Card | 358×441 | 350 | White |
| `cn_passport` | Chinese Passport | 390×567 | 300 | White |
| `cn_one_inch` | One-Inch Photo | 295×413 | 300 | White |
| `cn_two_inch` | Two-Inch Photo | 413×579 | 300 | White |
| `cn_small_one_inch` | Small One-Inch | 260×378 | 300 | White |
| `cn_driver_license` | Driver's License | 260×378 | 300 | White |
| `us_passport` | US Passport | 600×600 | 300 | White |
| `us_visa` | US Visa | 600×600 | 300 | White |
| `schengen_visa` | Schengen Visa | 413×531 | 300 | Light Gray |
| `uk_passport` | UK Passport | 413×531 | 300 | Light Gray |
| `jp_passport` | Japan Passport | 413×531 | 300 | White |

### Key Design: Image-to-Image Mode

This system uses Alibaba Bailian's image-to-image feature (`ref_img` parameter):

1. **User's photo as reference** (`ref_img`)
2. **AI preserves facial features** (via `ref_strength=0.75`)
3. **Only changes background, attire, and pose** (controlled via prompt)
4. **Not generating from scratch** — transforming the user's photo

### Setup & Deployment

#### 1. Install Dependencies

Python (virtual environment):
```bash
cd ~/photo-service
source venv/bin/activate
```

Go:
```bash
cd ~/photo-service
go mod download
```

#### 2. Start Services

**AI Image Service** (port 8091):
```bash
./start_ai_service.sh
# Or: source venv/bin/activate && python3 ai_photo.py
```

**Go Backend** (port 8090):
```bash
go run main.go
```

#### 3. Ensure HivisionIDPhotos is Running (port 8080)

HivisionIDPhotos must be deployed separately.

### API Reference

#### AI Image Service (port 8091)

**POST /ai-idphoto**

Parameters:
- `input_image` (file) — User's photo
- `spec` (string) — Photo type code
- `gender` (string) — `male` / `female`

Example:
```bash
curl -X POST http://127.0.0.1:8091/ai-idphoto \
  -F "input_image=@my_photo.jpg" \
  -F "spec=cn_one_inch" \
  -F "gender=male"
```

#### Full ID Photo Service (port 8090)

**POST /generate**

Parameters:
- `photo` (file) — User's photo
- `spec` (string) — Photo type code
- `gender` (string) — `male` / `female`

Example:
```bash
curl -X POST http://127.0.0.1:8090/generate \
  -F "photo=@my_photo.jpg" \
  -F "spec=cn_id" \
  -F "gender=female"
```

### Health Checks

```bash
# AI service
curl http://127.0.0.1:8091/health

# Full service
curl http://127.0.0.1:8090/health

# Supported photo types
curl http://127.0.0.1:8091/specs
curl http://127.0.0.1:8090/specs
```

### Notes

1. **API Keys** — Configurable via `config.json`. See `config.go` for all fields.
2. **Generation Time** — AI generation ~8-15s, Hivision cropping ~2-5s, total ~15-20s per photo.
3. **Fallback Strategy** — If AI service fails, falls back to direct cropping of the original photo (`ai_used=false`).
4. **Image Quality** — Upload photos with clearly recognizable faces, resolution ≥ 500px, JPG/PNG format.
5. **systemd Service** (Linux) — Create `/etc/systemd/system/ai-photo.service`:

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

### Troubleshooting

**AI service connection failed** — Check if service is running: `curl http://127.0.0.1:8091/health`

**Generated photo doesn't meet expectations** — Adjust `ref_strength` (currently 0.75), modify prompt template, or add negative_prompt content.

**Hivision service failed** — Confirm port 8080 is running, check if photo contains a face, review Hivision logs.

### Tech Stack

- AI Image Service: FastAPI + Alibaba Bailian SDK
- Backend: Go + Gin
- Cropping Service: HivisionIDPhotos (Python)

### Version

v2.0 — Integrated Alibaba Bailian AI image generation for formal ID photos from user photos.

---

<h2 id="chinese">AI 证件照生成系统 v2.0</h2>

基于阿里百炼通义万相 AI 生图的证件照生成系统。

### 核心特性

✅ **基于用户照片生成** — 使用用户上传的照片作为参考，保持人物面部特征
✅ **AI 正装改造** — 自动添加标准证件照着装（西装/正装）
✅ **背景替换** — 标准证件照背景色（白色/灰色等）
✅ **精确尺寸裁剪** — Hivision 服务提供精确的像素尺寸和 DPI 控制
✅ **多国证件支持** — 中国身份证/护照、美国护照/签证、申根签证等

### 系统架构

```
用户上传照片
    ↓
AI 生图服务 (8091端口)
    ├─ 阿里百炼 wanx2.1-t2i-turbo 模型
    ├─ ref_img: 用户上传的照片（保持面部特征）
    ├─ prompt: 证件照规范（背景、着装、姿态）
    └─ 输出: AI 生成的正装证件照
    ↓
Hivision 服务 (8080端口)
    ├─ 精确裁剪到标准尺寸
    ├─ 调整 DPI
    └─ 输出: 最终证件照
    ↓
Go 后端服务 (8090端口)
    └─ 返回给前端
```

### 支持的证件类型

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

### 关键说明：图生图模式

本系统使用阿里百炼的图生图功能（`ref_img` 参数）：

1. **用户照片作为参考图** (`ref_img`)
2. **AI 保持人物面部特征**（通过 `ref_strength=0.75`）
3. **仅改变背景、着装、姿态**（通过 prompt 控制）
4. **不是凭空生成**，而是基于用户照片改造

### 安装部署

#### 1. 安装依赖

Python 虚拟环境:
```bash
cd ~/photo-service
source venv/bin/activate
```

Go:
```bash
cd ~/photo-service
go mod download
```

#### 2. 启动服务

**AI 生图服务**（8091 端口）:
```bash
./start_ai_service.sh
# 或: source venv/bin/activate && python3 ai_photo.py
```

**Go 后端服务**（8090 端口）:
```bash
go run main.go
```

#### 3. 确保 HivisionIDPhotos 服务运行（8080 端口）

HivisionIDPhotos 需要单独部署。

### API 参考

#### AI 生图服务 (8091 端口)

**POST /ai-idphoto**

参数:
- `input_image` (file) — 用户照片
- `spec` (string) — 证件类型代码
- `gender` (string) — `male` / `female`

示例:
```bash
curl -X POST http://127.0.0.1:8091/ai-idphoto \
  -F "input_image=@my_photo.jpg" \
  -F "spec=cn_one_inch" \
  -F "gender=male"
```

#### 完整证件照服务 (8090 端口)

**POST /generate**

参数:
- `photo` (file) — 用户照片
- `spec` (string) — 证件类型代码
- `gender` (string) — `male` / `female`

示例:
```bash
curl -X POST http://127.0.0.1:8090/generate \
  -F "photo=@my_photo.jpg" \
  -F "spec=cn_id" \
  -F "gender=female"
```

### 健康检查

```bash
# AI 服务
curl http://127.0.0.1:8091/health

# 完整服务
curl http://127.0.0.1:8090/health

# 查看支持的证件类型
curl http://127.0.0.1:8091/specs
curl http://127.0.0.1:8090/specs
```

### 注意事项

1. **API Key 管理** — 通过 `config.json` 配置，参见 `config.go` 中所有字段。
2. **生成时间** — AI 生图约 8-15 秒，Hivision 裁剪约 2-5 秒，总计约 15-20 秒/张。
3. **降级策略** — AI 服务失败时，自动降级为直接使用原图裁剪（`ai_used=false`）。
4. **图片质量要求** — 上传图片需包含清晰可识别的人脸，分辨率 ≥ 500px，支持 JPG/PNG。
5. **systemd 服务**（Linux） — 创建 `/etc/systemd/system/ai-photo.service`:

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

### 故障排查

**AI 服务连接失败** — 检查服务是否启动：`curl http://127.0.0.1:8091/health`

**生成图片不符合预期** — 调整 `ref_strength` 参数（当前 0.75），修改 prompt 模板，或增加 negative_prompt 内容。

**Hivision 服务失败** — 确认 8080 端口服务运行，检查图片中是否包含人脸，查看 Hivision 日志。

### 技术栈

- AI 生图服务: FastAPI + 阿里百炼 SDK
- 后端服务: Go + Gin
- 裁剪服务: HivisionIDPhotos (Python)

### 版本

v2.0 — 集成阿里百炼 AI 生图，基于用户照片生成正装证件照
