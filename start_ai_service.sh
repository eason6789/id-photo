#!/bin/bash
# 启动AI证件照生成服务

cd "$(dirname "$0")"
source venv/bin/activate
python3 ai_photo.py
