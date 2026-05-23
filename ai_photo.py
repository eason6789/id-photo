#!/usr/bin/env python3
"""AI证件照生成 - 阿里百炼 FaceChain + 腾讯云 COS"""
import os, time, base64, uuid, asyncio
import httpx
from fastapi import FastAPI, File, UploadFile, Form, HTTPException
from qcloud_cos import CosConfig, CosS3Client
import uvicorn

app = FastAPI(title="AI证件照", version="7.0")

# ---- Config ----
DASHSCOPE_API_KEY = os.getenv("DASHSCOPE_API_KEY", "YOUR_KEY")
COS_SECRET_ID = os.getenv("COS_SECRET_ID", "YOUR_COS_SECRET_ID")
COS_SECRET_KEY = os.getenv("COS_SECRET_KEY", "YOUR_COS_SECRET_KEY")
COS_BUCKET = os.getenv("COS_BUCKET", "top-comm-1251416377")
COS_REGION = os.getenv("COS_REGION", "ap-guangzhou")

FACECHAIN_URL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/album/gen_potrait"
TASK_URL = "https://dashscope.aliyuncs.com/api/v1/tasks/"

# FaceChain portrait template (pre-uploaded to COS)
TEMPLATE_URL = f"https://{COS_BUCKET}.cos.{COS_REGION}.myqcloud.com/id-photo/templates/portrait_template.png"

# ---- Init COS ----
cos = CosS3Client(CosConfig(
    Region=COS_REGION,
    SecretId=COS_SECRET_ID,
    SecretKey=COS_SECRET_KEY
))

def upload_cos(data: bytes, folder: str, ext: str = ".jpg") -> str:
    """Upload to COS, return public URL."""
    key = f"id-photo/{folder}/{uuid.uuid4().hex}{ext}"
    cos.put_object(Bucket=COS_BUCKET, Body=data, Key=key, EnableMD5=False)
    return f"https://{COS_BUCKET}.cos.{COS_REGION}.myqcloud.com/{key}"

async def call_facechain(user_urls: list[str]) -> str:
    """Call FaceChain API, return result image URL."""
    headers = {
        "Authorization": f"Bearer {DASHSCOPE_API_KEY}",
        "Content-Type": "application/json",
        "X-DashScope-Async": "enable"
    }
    payload = {
        "model": "facechain-generation",
        "parameters": {
            "style": "train_free_portrait_url_template",
            "n": 1,
            "skin_retouch": True
        },
        "input": {
            "template_url": TEMPLATE_URL,
            "user_urls": user_urls
        }
    }

    async with httpx.AsyncClient(timeout=30.0) as client:
        r = await client.post(FACECHAIN_URL, json=payload, headers=headers)
        r.raise_for_status()
        data = r.json()

        if "output" not in data or "task_id" not in data["output"]:
            raise Exception(f"FaceChain提交失败: {data}")

        task_id = data["output"]["task_id"]
        print(f"FaceChain task: {task_id}")

        # Poll for result (max 3 minutes)
        for i in range(90):
            await asyncio.sleep(2)
            qr = await client.get(
                f"{TASK_URL}{task_id}",
                headers={"Authorization": f"Bearer {DASHSCOPE_API_KEY}"}
            )
            qdata = qr.json()
            status = qdata["output"]["task_status"]

            if status == "SUCCEEDED":
                return qdata["output"]["results"][0]["url"]
            elif status == "FAILED":
                msg = qdata["output"].get("message", "未知错误")
                raise Exception(f"FaceChain生成失败: {msg}")

        raise Exception("FaceChain生成超时")

@app.post("/ai-idphoto")
async def generate_id_photo(
    input_image: UploadFile = File(...),
    spec: str = Form(...),
    gender: str = Form(...)
):
    try:
        image_bytes = await input_image.read()
        print(f"收到请求: spec={spec}, gender={gender}, size={len(image_bytes)}")

        # 1. Upload original to COS
        cos_url = upload_cos(image_bytes, "uploads")
        print(f"原始图COS: {cos_url}")

        # 2. Call FaceChain (user's face → professional portrait)
        result_url = await call_facechain([cos_url])
        print(f"FaceChain结果: {result_url}")

        # 3. Download FaceChain result
        async with httpx.AsyncClient(timeout=60.0) as client:
            rr = await client.get(result_url)
            rr.raise_for_status()
            result_bytes = rr.content

        # 4. Upload result to COS
        result_cos_url = upload_cos(result_bytes, "results")
        print(f"结果COS: {result_cos_url}")

        # 5. Return (base64 for backward compat, cos_url for new flow)
        result_b64 = base64.b64encode(result_bytes).decode()

        return {
            "success": True,
            "image": f"data:image/jpeg;base64,{result_b64}",
            "cos_url": result_cos_url
        }

    except Exception as e:
        print(f"生成失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/cos/upload")
async def cos_upload(file: UploadFile = File(...)):
    """Go backend uploads files to COS via this endpoint."""
    try:
        data = await file.read()
        folder = "uploads"
        ext = os.path.splitext(file.filename or ".jpg")[1] or ".jpg"
        url = upload_cos(data, folder, ext)
        return {"success": True, "url": url}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health():
    return {"status": "ok", "engine": "facechain-generation", "storage": "cos"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8091)
