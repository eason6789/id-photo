#!/usr/bin/env python3
"""AI证件照生成 - 阿里百炼wanx-v1图生图"""
import os, time, base64, asyncio
import httpx
from fastapi import FastAPI, File, UploadFile, Form, HTTPException
import uvicorn

app = FastAPI(title="AI证件照", version="6.0")

DASHSCOPE_API_KEY = "YOUR_KEY"
WANX_API_URL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis"

# 提示词
SPEC_PROMPTS = {
    "cn_id": "标准身份证证件照，深蓝色背景，人物穿深色西装，白色衬衫，系领带，正式免冠",
    "cn_passport": "中国护照证件照，白色背景，人物穿深色西装，白色衬衫，正式免冠",
    "cn_one_inch": "一寸证件照，白色背景，人物穿深色西装，白色衬衫，领带，正面免冠",
    "cn_two_inch": "二寸证件照，红色背景，人物穿深色西装，白色衬衫，正面免冠",
    "cn_one_inch_blue": "一寸证件照，蓝色背景，人物穿深色西装，白色衬衫，正面免冠",
    "cn_one_inch_red": "一寸证件照，红色背景，人物穿深色西装，白色衬衫，正面免冠",
    "cn_two_inch_blue": "二寸证件照，蓝色背景，人物穿深色西装，白色衬衫，正面免冠",
    "cn_two_inch_red": "二寸证件照，红色背景，人物穿深色西装，白色衬衫，正面免冠",
    "cn_driver_license": "驾驶证证件照，白色背景，人物穿深色有领上衣，白色衬衫，正面免冠",
}

NEGATIVE_PROMPT = "卡通，动漫，绘画，模糊，低质量，水印，过度美颜，休闲装，复杂背景，侧脸，闭眼"

def get_prompt(spec: str) -> str:
    base_spec = spec.replace("_male", "").replace("_female", "")
    return SPEC_PROMPTS.get(base_spec, SPEC_PROMPTS["cn_two_inch"])

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
        
        prompt = get_prompt(spec)
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
