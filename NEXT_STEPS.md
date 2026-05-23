# 下一步改进计划

## 📋 文档已创建

1. **README.md** - 完整的项目文档
   - 系统架构、技术栈、核心流程
   - API 接口、数据库设计
   - 部署配置、故障排查

2. **XIANYU_INTEGRATION.md** - 咸鱼 Skills 接入指引
   - Webhook 集成步骤
   - 测试清单、故障排查
   - 面向 AI 模型的详细说明

3. **PHOTO_SPECS_STANDARDS.md** - 证件照规格与标准规范
   - 通用规范要求（头部占比、位置、表情等）
   - 25+ 种证件照详细规格（中国、美国、日本、欧洲等）
   - 尺寸对照表
   - 实现建议（OpenCV、HivisionIDPhotos）

4. **ISSUES_AND_FIXES.md** - 问题汇总
   - 已修复：尺寸、前端上传框
   - 待解决：人脸相似度、背景色

---

## 🔴 核心问题总结

### 问题 1: AI 生成的人脸不符合证件照规范

**具体表现**:
1. **人脸相似度低**: 生成的脸与原照片完全不同
2. **头部占比错误**: 没有按 60-70% 的规范要求
3. **头部位置不对**: 眼睛位置不在 55-60% 高度
4. **背景色不匹配**: 选择蓝底但结果不是蓝色
5. **性别风格不准**: 选择 male/female 但效果不明显

**根本原因**:
- HivisionIDPhotos API 的 `f_idcard_male`/`f_idcard_female` 是风格化模板
- 不支持精细控制（头部占比、眼睛位置、背景色）
- 只做风格化，不做规范化

---

## 💡 推荐的解决方案

### 方案 A: 完全不用 AI，纯图像处理（最可控）

**流程**:
```
用户上传照片
    ↓
1. 人脸检测（OpenCV / dlib）
    ↓
2. 头部占比调整（缩放到 65%）
    ↓
3. 头部位置调整（眼睛移到 57% 高度）
    ↓
4. 抠图（RemoveBG / U2-Net）
    ↓
5. 背景替换（填充指定颜色）
    ↓
6. 尺寸裁剪（已实现）
    ↓
输出符合规范的证件照
```

**优点**:
- ✅ 100% 保留人脸相似度
- ✅ 100% 符合规范（头部占比、位置）
- ✅ 100% 背景色匹配
- ✅ 可控性强，无意外

**缺点**:
- ❌ 无 AI 美化效果
- ❌ 需要部署抠图模型（或使用 RemoveBG API）

**技术栈**:
- **人脸检测**: OpenCV Haar Cascade / dlib
- **抠图**: RemoveBG API（收费）或 U2-Net（免费，本地部署）
- **图像处理**: PHP GD / Python PIL / Go imaging

**实现难度**: ⭐⭐⭐ 中等

---

### 方案 B: HivisionIDPhotos 开源方案（推荐）

**GitHub**: https://github.com/Zeyi-Lin/HivisionIDPhotos

**功能**:
- 自动人脸检测
- **自动调整头部占比**到规范要求
- **自动调整头部位置**（眼睛位置）
- 背景抠图 + 替换
- 支持 25+ 种证件照规格
- 轻度美颜（可选）

**部署方式**:
```bash
# 1. 克隆项目
git clone https://github.com/Zeyi-Lin/HivisionIDPhotos.git
cd HivisionIDPhotos

# 2. 安装依赖
pip install -r requirements.txt

# 3. 下载模型
python download_models.py

# 4. 启动 API 服务
python app.py --port 8090

# 5. Go 后端调用
POST http://localhost:8090/idphoto
{
  "image": "base64图片",
  "height": 413,
  "width": 295,
  "bg_color": "1E90FF"
}
```

**优点**:
- ✅ 开源免费
- ✅ 自动符合规范
- ✅ 保真度高（不改变人脸）
- ✅ 支持背景色指定

**缺点**:
- ❌ 需要部署 Python 环境
- ❌ GPU 推荐（CPU 也可以但慢）

**实现难度**: ⭐⭐ 简单（已有完整项目）

---

### 方案 C: 切换到保真度更高的 AI（次选）

**选项 1: 百度智能云 - 人像特效**
```bash
# API 文档: https://ai.baidu.com/ai-doc/IMAGEPROCESS/

POST https://aip.baidubce.com/rest/2.0/image-process/v1/portrait_seg
{
  "image": "base64图片",
  "type": "id_photo",
  "bg_color": "1E90FF"
}
```

**优点**:
- ✅ 支持背景色参数
- ✅ 有保真模式
- ✅ 云端服务，无需部署

**缺点**:
- ❌ 收费
- ❌ 仍然有 AI 美化，可能改变人脸

**选项 2: 腾讯云 - 人像变换**
类似百度，也是商业 API

**实现难度**: ⭐⭐ 简单（改 API 调用）

---

## 🎯 推荐实施步骤

### 第一阶段: 快速修复（1-2天）

#### 步骤 1: 部署 HivisionIDPhotos
```bash
# 在服务器上部署
ssh <user>@<server_ip>

# 安装 Python 环境（如果没有）
apt-get update && apt-get install python3-pip python3-venv -y

# 创建虚拟环境
cd /root
python3 -m venv idphoto-env
source idphoto-env/bin/activate

# 克隆并安装
git clone https://github.com/Zeyi-Lin/HivisionIDPhotos.git
cd HivisionIDPhotos
pip install -r requirements.txt
python download_models.py

# 启动服务
nohup python app.py --port 8090 > idphoto.log 2>&1 &
```

#### 步骤 2: 修改 Go 代码调用 HivisionIDPhotos
```go
// 替换 callHivisionIDPhotos 函数
func callHivisionIDPhoto(imagePath, bgColor string, width, height int) (string, error) {
    imageData := base64.StdEncoding.EncodeToString(readFile(imagePath))
    
    reqBody := map[string]interface{}{
        "image": imageData,
        "height": height,
        "width": width,
        "bg_color": bgColor[1:], // 去掉 # 号
    }
    
    jsonData, _ := json.Marshal(reqBody)
    
    resp, err := http.Post("http://127.0.0.1:8090/idphoto", 
                          "application/json", 
                          bytes.NewReader(jsonData))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    // result["image"] 是处理后的 base64
    // 保存到文件
    processedImage, _ := base64.StdEncoding.DecodeString(result["image"].(string))
    outputPath := strings.Replace(imagePath, ".jpg", "_processed.jpg", 1)
    os.WriteFile(outputPath, processedImage, 0644)
    
    return outputPath, nil
}

// 修改 generatePhoto 函数
// 删除: callHivisionIDPhotos, pollResult, downloadFile
// 改为: callHivisionIDPhoto
```

#### 步骤 3: 测试验证
```bash
# 上传测试照片
# 选择"一寸蓝底"
# 验证结果:
#   - 背景是蓝色 ✓
#   - 头部占比 65% ✓
#   - 眼睛在 57% 高度 ✓
#   - 人脸与原图相同 ✓
```

**预计时间**: 4-6 小时

---

### 第二阶段: 优化体验（3-5天）

#### 1. 添加质量检查
```go
func validatePhoto(imagePath string) error {
    // 使用 OpenCV 检测:
    // - 是否有人脸
    // - 是否只有一个人脸
    // - 人脸是否正面
    // - 是否有眼镜反光
    // - 清晰度是否足够
}
```

#### 2. 前端优化
- 添加上传前预检查（本地 JS 人脸检测）
- 显示头部占比预览框
- 给出姿势调整建议（"请将头部居中"）

#### 3. 添加更多规格
- 更新 main.go 中的 photoSpecs（使用 /tmp/new_photo_specs.go）
- 前端下拉框同步更新（35+ 种规格）

#### 4. 性能优化
- HivisionIDPhotos 使用 GPU 加速
- 添加结果缓存（相同照片 + 相同规格）

---

### 第三阶段: 高级功能（可选）

#### 1. 智能建议
- AI 检测照片质量并给出建议
- "光线不足，建议重新拍摄"
- "头部偏左，请调整"

#### 2. 批量处理
- 一次上传多张照片
- 自动生成多种规格

#### 3. 在线调整
- 前端允许用户手动调整头部位置
- 实时预览符合规范的效果

---

## 📊 方案对比

| 方案 | 人脸相似度 | 符合规范 | 背景色 | 美化效果 | 实施难度 | 成本 |
|------|-----------|---------|--------|---------|---------|-----|
| **方案A: 纯图像处理** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐ | ⭐⭐⭐ | 低 |
| **方案B: HivisionIDPhotos** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐ | 免费 |
| **方案C: 商业AI** | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | 中 |
| **当前: HivisionIDPhotos** | ⭐ | ⭐ | ⭐ | ⭐⭐⭐⭐ | ⭐ | 低 |

**综合推荐**: **方案B（HivisionIDPhotos）**
- 开源免费
- 完全符合规范
- 保留人脸特征
- 实施简单

---

## ✅ 立即可做的临时措施

在部署新方案前，可以先做这些：

### 1. 前端添加说明
```html
<div class="warning">
  <strong>⚠️ 重要提示：</strong>
  <ul>
    <li>AI 会进行风格化处理，生成结果可能与原照片有较大差异</li>
    <li>如需100%保留原貌，建议联系客服获取专业版</li>
    <li>建议上传时头部居中、正面直视、光线充足</li>
  </ul>
</div>
```

### 2. 增加规格选项
将 /tmp/new_photo_specs.go 中的配置替换到 main.go
前端增加 35+ 种规格下拉选项

### 3. 记录用户反馈
在数据库添加 feedback 表，收集用户对结果的满意度

---

## 📞 需要决策的问题

**请您告诉我：**

1. **采用哪个方案？**
   - [ ] 方案 A: 纯图像处理（无美化）
   - [ ] 方案 B: HivisionIDPhotos（推荐）
   - [ ] 方案 C: 商业 AI（百度/腾讯云）
   - [ ] 继续研究 HivisionIDPhotos 参数调优

2. **时间安排？**
   - [ ] 立即开始（今天部署 HivisionIDPhotos）
   - [ ] 本周内完成
   - [ ] 先做临时措施，慢慢规划

3. **功能优先级？**
   - [ ] 优先修复人脸相似度（最重要）
   - [ ] 优先修复背景色匹配
   - [ ] 优先增加更多规格
   - [ ] 全部都要

**我的建议**: 
立即部署 HivisionIDPhotos（2小时内完成），先解决核心问题，后续再优化体验。

---

**创建时间**: 2026-03-18
**待您反馈后开始实施**
