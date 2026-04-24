# 快速启动指南

## 一键测试（开发环境）

### 1. 启动AI生图服务

**终端1**:
```bash
cd ~/photo-service
source venv/bin/activate
python3 ai_photo.py
```

看到以下输出表示成功：
```
============================================================
AI证件照生成服务启动
端口: 8091
基于阿里百炼通义万相wanx2.1模型
图生图模式：保持用户面部特征，改造为证件照
============================================================
INFO:     Started server process [12345]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://0.0.0.0:8091
```

### 2. 测试AI服务（新终端）

**终端2**:
```bash
# 测试健康检查
curl http://127.0.0.1:8091/health

# 查看支持的证件类型
curl http://127.0.0.1:8091/specs

# 生成证件照（需要一张测试照片）
curl -X POST http://127.0.0.1:8091/ai-idphoto \
  -F "input_image=@test.jpg" \
  -F "spec=cn_one_inch" \
  -F "gender=male" \
  | python3 -m json.tool
```

或使用测试脚本：
```bash
cd ~/photo-service
./test_ai_service.sh test.jpg cn_one_inch male
```

### 3. 启动完整服务（需要Hivision）

**终端3**:
```bash
cd ~/photo-service
go run main.go
```

**测试完整流程**:
```bash
curl -X POST http://127.0.0.1:8090/generate \
  -F "photo=@test.jpg" \
  -F "spec=cn_id" \
  -F "gender=male"
```

## 生产部署（Linux服务器）

### 方法1: systemd服务

```bash
# 1. 复制项目到/root/photo-service
sudo mkdir -p /root/photo-service
sudo cp -r ~/photo-service/* /root/photo-service/

# 2. 安装systemd服务
sudo cp /root/photo-service/ai-photo.service /etc/systemd/system/
sudo systemctl daemon-reload

# 3. 启动服务
sudo systemctl start ai-photo
sudo systemctl enable ai-photo

# 4. 查看状态
sudo systemctl status ai-photo

# 5. 查看日志
sudo journalctl -u ai-photo -f
```

### 方法2: screen后台运行

```bash
# 启动AI服务
screen -S ai-photo
cd ~/photo-service
source venv/bin/activate
python3 ai_photo.py
# 按 Ctrl+A, D 分离

# 启动Go服务
screen -S go-backend
cd ~/photo-service
go run main.go
# 按 Ctrl+A, D 分离

# 重新连接
screen -r ai-photo
screen -r go-backend

# 查看所有screen
screen -ls
```

### 方法3: nohup后台运行

```bash
cd ~/photo-service
source venv/bin/activate
nohup python3 ai_photo.py > ai_photo.log 2>&1 &
echo $! > ai_photo.pid

nohup go run main.go > go_backend.log 2>&1 &
echo $! > go_backend.pid

# 停止服务
kill $(cat ai_photo.pid)
kill $(cat go_backend.pid)
```

## 验证部署

```bash
# 1. 检查端口占用
netstat -tlnp | grep 8091  # AI服务
netstat -tlnp | grep 8090  # Go服务
netstat -tlnp | grep 8080  # Hivision服务

# 2. 健康检查
curl http://127.0.0.1:8091/health
curl http://127.0.0.1:8090/health

# 3. 完整测试
curl -X POST http://127.0.0.1:8090/generate \
  -F "photo=@test.jpg" \
  -F "spec=cn_one_inch" \
  -F "gender=male" \
  -o result.json

# 查看结果
cat result.json | python3 -m json.tool

# 4. 提取生成的图片
cat result.json | python3 -c "
import sys, json, base64
data = json.load(open('result.json'))
img = base64.b64decode(data['image_base64_hd'])
open('generated.jpg', 'wb').write(img)
print('图片已保存到 generated.jpg')
"
```

## 前端集成示例

### HTML + JavaScript

```html
<!DOCTYPE html>
<html>
<head>
    <title>证件照生成</title>
</head>
<body>
    <h1>AI证件照生成</h1>
    <form id="photoForm">
        <input type="file" id="photo" accept="image/*" required>
        <select id="spec" required>
            <option value="cn_id">中国身份证</option>
            <option value="cn_passport">中国护照</option>
            <option value="cn_one_inch">一寸照片</option>
            <option value="us_passport">美国护照</option>
        </select>
        <select id="gender" required>
            <option value="male">男</option>
            <option value="female">女</option>
        </select>
        <button type="submit">生成证件照</button>
    </form>
    <div id="result"></div>

    <script>
        document.getElementById('photoForm').onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData();
            formData.append('photo', document.getElementById('photo').files[0]);
            formData.append('spec', document.getElementById('spec').value);
            formData.append('gender', document.getElementById('gender').value);

            const response = await fetch('http://127.0.0.1:8090/generate', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();
            if (data.success) {
                document.getElementById('result').innerHTML =
                    `<img src="data:image/jpeg;base64,${data.image_base64_hd}" />
                     <p>规格: ${data.spec.name}</p>
                     <p>AI生成: ${data.ai_used ? '是' : '否'}</p>`;
            }
        };
    </script>
</body>
</html>
```

## 常见问题

### Q: 生成速度慢？
A: AI生图需要8-15秒，这是正常的。如需加速可以：
- 使用更快的模型（如有）
- 启用并发处理
- 增加多个API Key轮询

### Q: 生成的照片人脸不像？
A: 调整 `ai_photo.py` 中的 `ref_strength` 参数：
- 当前0.75，范围0.0-1.0
- 越高越接近原图（但可能正装效果差）
- 越低正装效果好（但可能人脸变化大）

### Q: AI服务调用失败？
A: 检查：
1. API Key是否有效
2. 网络是否能访问阿里百炼API
3. 查看错误日志
4. 尝试使用备用API Key

### Q: 如何添加新的证件类型？
A: 修改两个文件的 `PHOTO_SPECS` / `photoSpecs` 配置：
1. `ai_photo.py` - 添加规格和prompt
2. `main.go` - 添加规格定义

## 下一步优化

- [ ] API Key从环境变量读取
- [ ] 添加缓存机制（相同照片+规格不重复生成）
- [ ] 批量处理接口
- [ ] WebSocket实时进度推送
- [ ] 前端UI界面
- [ ] 图片水印/防盗链
- [ ] 数据库存储生成记录
- [ ] 价格计算/配额管理
