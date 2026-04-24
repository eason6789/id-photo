#!/bin/bash
# 测试AI证件照服务

# 检查参数
if [ $# -lt 3 ]; then
    echo "用法: $0 <图片路径> <证件类型> <性别>"
    echo "示例: $0 test.jpg cn_one_inch male"
    echo ""
    echo "支持的证件类型:"
    echo "  cn_id - 中国身份证"
    echo "  cn_passport - 中国护照"
    echo "  cn_one_inch - 一寸照片"
    echo "  cn_two_inch - 二寸照片"
    echo "  us_passport - 美国护照"
    echo "  schengen_visa - 申根签证"
    echo ""
    echo "性别: male 或 female"
    exit 1
fi

IMAGE_PATH="$1"
SPEC="$2"
GENDER="$3"

if [ ! -f "$IMAGE_PATH" ]; then
    echo "错误: 图片文件不存在: $IMAGE_PATH"
    exit 1
fi

echo "=========================================="
echo "测试AI证件照服务"
echo "图片: $IMAGE_PATH"
echo "证件类型: $SPEC"
echo "性别: $GENDER"
echo "=========================================="

# 调用API
RESPONSE=$(curl -s -X POST http://127.0.0.1:8091/ai-idphoto \
    -F "input_image=@$IMAGE_PATH" \
    -F "spec=$SPEC" \
    -F "gender=$GENDER")

# 检查是否成功
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo "✅ 生成成功!"

    # 提取base64图片并保存
    echo "$RESPONSE" | python3 -c "
import sys, json, base64
data = json.load(sys.stdin)
if data.get('success'):
    img_data = base64.b64decode(data['image_base64'])
    with open('output_ai.jpg', 'wb') as f:
        f.write(img_data)
    print('图片已保存到: output_ai.jpg')
    print('规格信息:', json.dumps(data['spec_info'], ensure_ascii=False))
"
else
    echo "❌ 生成失败:"
    echo "$RESPONSE" | python3 -m json.tool
fi
