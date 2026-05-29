#!/usr/bin/env python3
"""AI证件照生成 - 阿里百炼wanx-v1图生图"""
import os, time, base64, asyncio
import httpx
from fastapi import FastAPI, File, UploadFile, Form, HTTPException
import uvicorn

app = FastAPI(title="AI证件照", version="6.0")

DASHSCOPE_API_KEY = os.environ["DASHSCOPE_API_KEY"]
WANX_API_URL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis"

# 提示词（中英双语，英文结构化版本用于图生图模型）
SPEC_PROMPTS = {
    # ========== 图生图标准版（wanx-v1 ref_img模式） ==========
    "cn_id": {
        "zh": "请基于参考图片生成一张专业标准证件照。保留人物真实五官比例、脸型基础和肤色特征，可轻度优化面部轮廓使其更清晰自然，但不要过度美颜。人物穿着深色西装外套、白色衬衫，系领带，发型干净利落。背景为纯深蓝色背景，专业影棚质感，光线柔和均匀，正面免冠，表情自然。头部居中，头部占比约70%，头顶预留适当空间，3:4竖版构图。最终成片达到专业照相馆拍摄水准。",
        "en": "Please generate a professional ID photo based on the reference image. Preserve the subject's authentic facial proportions, face shape, and skin tone. Light facial refinement is acceptable for clarity, but avoid excessive retouching. The subject wears a dark suit jacket, white shirt, and tie, with a clean professional hairstyle. Background: solid deep blue, professional studio quality. Lighting: soft and even, no harsh shadows. Front-facing headshot with natural expression, straight-on to camera. Composition: head centered with 70% head ratio, appropriate space above the head, 3:4 vertical format. Final output: professional passport-photo quality."
    },
    "cn_passport": {
        "zh": "请基于参考图片生成中国护照证件照。保留人物五官特征、脸型基础和自然肤色，可适度优化皮肤使其更细腻均匀，但不要磨皮过度。人物穿深色西装、白色衬衫，正式免冠，正面直视镜头。背景为纯白色背景，专业影棚打光，光线柔和均匀无硬阴影。正面构图，头部居中，头顶留白，3:4竖版。最终成片符合护照照片标准。",
        "en": "Please generate a Chinese passport ID photo based on the reference image. Preserve the subject's authentic facial features, face shape, and natural skin tone. Acceptable light skin smoothing for uniformity, but avoid aggressive retouching. Subject wears a dark suit with white shirt, formal no-headwear headshot, front-facing straight to camera. Background: pure white, professional studio lighting with soft even illumination and no harsh shadows. Composition: centered head with space above, 3:4 vertical ratio. Final output meets passport photo standards."
    },
    "cn_one_inch": {
        "zh": "请基于参考图片生成一寸证件照。保留人物真实五官特征、脸型比例和肤色，可轻度美颜优化但不要过度。人物穿深色西装外套、白色衬衫，可搭配领带，发型整洁专业。背景为纯白色背景，专业影棚质感，柔和均匀光线。头部居中，正面免冠，自然微笑，头部占比约70%，头顶预留空间，3:4竖版构图。成片达到专业一寸证件照标准。",
        "en": "Please generate a one-inch ID photo based on the reference image. Preserve the subject's authentic facial features, face shape proportions, and skin tone. Light enhancement is acceptable; avoid over-retouching. Subject wears a dark suit jacket, white shirt, optional tie, with neat professional hairstyle. Background: pure white, professional studio quality with soft even lighting. Head centered, front-facing headshot with natural smile, 70% head ratio with space above the head, 3:4 vertical composition. Output meets professional one-inch ID photo standards."
    },
    "cn_two_inch": {
        "zh": "请基于参考图片生成二寸证件照。保留人物真实五官特征、脸型基础和自然肤色，可适度优化面部轮廓但保持真实质感。人物穿深色西装外套、白色衬衫，正面免冠，自然表情。背景为纯红色背景，专业影棚质感，柔和均匀光线，无硬阴影。头部居中，头部占比约70%，头顶留白，3:4竖版构图。成片符合二寸证件照标准。",
        "en": "Please generate a two-inch ID photo based on the reference image. Preserve authentic facial features, face shape, and natural skin tone. Acceptable light facial contour refinement while maintaining realistic texture. Subject wears a dark suit jacket and white shirt, front-facing headshot with natural expression. Background: solid red, professional studio quality with soft even lighting and no harsh shadows. Head centered with 70% head ratio, space above the head, 3:4 vertical ratio. Output meets two-inch ID photo standards."
    },
    "cn_one_inch_blue": {
        "zh": "请基于参考图片生成一寸蓝色背景证件照。保留人物五官比例、脸型特征和肤色自然度，可轻度优化但不过度美颜。人物穿深色西装、白色衬衫，正面免冠，自然微笑。蓝色纯色背景，专业影棚柔和均匀光线，无硬阴影。头部居中，正面构图，头部占比约70%，头顶预留适当空间，3:4竖版。成片达到专业一寸证件照标准。",
        "en": "Please generate a one-inch ID photo with blue background based on the reference image. Preserve the subject's authentic facial proportions, face shape, and natural skin tone. Light refinement is acceptable; avoid excessive retouching. Subject wears a dark suit and white shirt, front-facing headshot with natural smile. Background: solid blue, professional studio with soft even lighting, no harsh shadows. Head centered in front-facing composition with 70% head ratio, appropriate space above, 3:4 vertical ratio. Output meets professional one-inch blue-background ID photo standards."
    },
    "cn_one_inch_red": {
        "zh": "请基于参考图片生成一寸红色背景证件照。保留人物真实五官特征、脸型比例和肤色，可适度优化但保持真实感。人物穿深色西装外套、白色衬衫，正面免冠，表情自然。红色纯色背景影棚质感，柔和均匀光线无硬阴影。头部居中，正面构图，头部占比约70%，头顶留白，3:4竖版。成片符合专业一寸证件照标准。",
        "en": "Please generate a one-inch ID photo with red background based on the reference image. Preserve the subject's authentic facial features, face shape, and natural skin tone. Moderate enhancement is acceptable while maintaining authenticity. Subject wears a dark suit jacket and white shirt, front-facing headshot with natural expression. Background: solid red with professional studio quality, soft even lighting, no harsh shadows. Head centered with 70% head ratio and space above, 3:4 vertical composition. Output meets professional one-inch red-background ID photo standards."
    },
    "cn_two_inch_blue": {
        "zh": "请基于参考图片生成二寸蓝色背景证件照。保留人物五官特征、脸型基础和肤色自然度，可轻度面部轮廓优化但不要磨皮过度。人物穿深色西装外套、白色衬衫，正面免冠自然表情。蓝色纯色背景，专业影棚质感，柔和均匀光线无硬阴影。头部居中，正面构图，头部占比约70%，头顶预留空间，3:4竖版。成片符合二寸证件照标准。",
        "en": "Please generate a two-inch ID photo with blue background based on the reference image. Preserve authentic facial features, face shape, and natural skin tone. Light facial contour refinement is acceptable; avoid heavy skin smoothing. Subject wears a dark suit jacket and white shirt, front-facing headshot with natural expression. Background: solid blue, professional studio with soft even lighting, no harsh shadows. Head centered, front-facing composition, 70% head ratio with space above, 3:4 vertical ratio. Output meets professional two-inch blue-background ID photo standards."
    },
    "cn_two_inch_red": {
        "zh": "请基于参考图片生成二寸红色背景证件照。保留人物真实五官特征、脸型比例和自然肤色，可适度优化面部轮廓但保持真实质感。人物穿深色西装、白色衬衫，正面免冠，表情自信自然。红色纯色影棚背景，柔和均匀光线无硬阴影。头部居中，正面构图，头部占比约70%，头顶留白，3:4竖版。成片达到二寸证件照专业标准。",
        "en": "Please generate a two-inch ID photo with red background based on the reference image. Preserve the subject's authentic facial features, face shape, and natural skin tone. Moderate facial refinement is acceptable while keeping a realistic texture. Subject wears a dark suit with white shirt, front-facing headshot with natural confident expression. Background: solid red with professional studio quality, soft even illumination, no harsh shadows. Head centered, front-facing composition, 70% head ratio with space above, 3:4 vertical ratio. Output meets professional two-inch red-background ID photo standards."
    },
    "cn_driver_license": {
        "zh": "请基于参考图片生成驾驶证证件照。保留人物真实五官特征、脸型基础和肤色，可适度优化但不要过度修图。人物穿深色有领上衣、白色衬衫，衬衫领口整洁，正面免冠，表情自然自信。白色纯色背景影棚质感，柔和均匀光线无硬阴影。头部居中，正面构图，头部占比约70%，头顶预留适当空间，3:4竖版。成片符合驾驶证照片标准。",
        "en": "Please generate a driver's license ID photo based on the reference image. Preserve authentic facial features, face shape, and natural skin tone. Light enhancement is acceptable; avoid over-editing. Subject wears a dark collared top with a white shirt with a neat collar, front-facing headshot with natural confident expression. Background: pure white, professional studio quality with soft even lighting and no harsh shadows. Head centered, front-facing composition, 70% head ratio, appropriate space above the head, 3:4 vertical ratio. Final output meets driver's license photo standards."
    },
}

NEGATIVE_PROMPT = "卡通，动漫，绘画，模糊，低质量，水印，过度美颜，休闲装，复杂背景，侧脸，闭眼"

def get_prompt(spec: str, lang: str = "en") -> str:
    """返回指定语言的提示词，lang: en=结构化英文版，zh=中文版"""
    base_spec = spec.replace("_male", "").replace("_female", "")
    entry = SPEC_PROMPTS.get(base_spec, SPEC_PROMPTS["cn_two_inch"])
    return entry.get(lang, entry["en"])

async def call_wanx_api(image_base64: str, prompt: str) -> str:
    headers = {
        "Authorization": f"Bearer {DASHSCOPE_API_KEY}",
        "Content-Type": "application/json",
        "X-DashScope-Async": "enable"
    }
    
    # 关键：用data:image/jpeg;base64,格式
    payload = {
        "model": "wanx-v1",
        "input": {
            "prompt": prompt,
            "negative_prompt": NEGATIVE_PROMPT,
            "ref_img": f"data:image/jpeg;base64,{image_base64}"
        },
        "parameters": {"size": "768*1152", "n": 1}
    }
    
    async with httpx.AsyncClient(timeout=180.0) as client:
        # 提交任务
        response = await client.post(WANX_API_URL, json=payload, headers=headers)
        response.raise_for_status()
        result = response.json()
        
        task_id = result["output"]["task_id"]
        print(f"Task ID: {task_id}")
        
        # 轮询等待
        for i in range(90):
            await asyncio.sleep(2)
            query_url = f"https://dashscope.aliyuncs.com/api/v1/tasks/{task_id}"
            query_response = await client.get(query_url, headers={"Authorization": f"Bearer {DASHSCOPE_API_KEY}"})
            query_result = query_response.json()
            task_status = query_result["output"]["task_status"]
            
            if task_status == "SUCCEEDED":
                return query_result["output"]["results"][0]["url"]
            elif task_status == "FAILED":
                raise Exception(f"生成失败: {query_result['output'].get('message')}")
        
        raise Exception("生成超时")

async def download_image(url: str) -> bytes:
    async with httpx.AsyncClient(timeout=30.0) as client:
        response = await client.get(url)
        response.raise_for_status()
        return response.content

@app.post("/ai-idphoto")
async def generate_id_photo(input_image: UploadFile = File(...), spec: str = Form(...), gender: str = Form(...)):
    try:
        image_bytes = await input_image.read()
        image_base64 = base64.b64encode(image_bytes).decode('utf-8')
        print(f"收到请求: spec={spec}, size={len(image_bytes)}")
        
        prompt = get_prompt(spec, lang="en")
        print(f"提示词: {prompt}")
        
        # 调用API
        image_url = await call_wanx_api(image_base64, prompt)
        print(f"生成成功: {image_url}")
        
        # 下载图片
        image_data = await download_image(image_url)
        
        return {
            "success": True,
            "image": f"data:image/jpeg;base64,{base64.b64encode(image_data).decode('utf-8')}"
        }
    except Exception as e:
        print(f"生成失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8091)
