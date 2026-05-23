# Deploy: AI 证件照生成系统 (ID Photo)

## Architecture

- **Frontend**: Static HTML served by Nginx at `portfolio-home/src/claw/photo/`
- **Go Backend** (port 8080): Main API server, handles token management, photo upload, and orchestration
- **AI 生图服务** (port 8091): Python FastAPI service, calls Aliyun Bailian wanx2.1-t2i-turbo model for AI photo generation
- **HivisionIDPhotos** (port 8080): Python service for precise cropping, DPI control, and face matting

```
User --> Nginx --> /claw/photo/           --> frontend/index.html (static)
                --> /claw/api/*           --> localhost:8080 (Go backend)
                       |
                       +-- AI生图: localhost:8091 (Python + 阿里百炼)
                       +-- 裁剪:   localhost:8080 (HivisionIDPhotos)
```

## Deploy Frontend

The frontend is a single static HTML file. Copy it to the server:

```bash
scp frontend/index.html root@<SERVER_IP>:/root/project/portfolio-home/src/claw/photo/
```

## Deploy Go Backend

1. Build the binary locally:

   ```bash
   cd /Users/easonlv/progs/id-photo
   go build -o photo-service main.go
   ```

2. Upload to the server:

   ```bash
   scp photo-service root@<SERVER_IP>:/root/photo-service/
   ```

3. Restart the systemd service:

   ```bash
   ssh root@<SERVER_IP>
   systemctl restart photo-service
   ```

## Deploy AI 生图 Service (Python)

1. Copy the Python service file and install dependencies:

   ```bash
   scp ai_photo.py requirements.txt root@<SERVER_IP>:/root/photo-service/
   ssh root@<SERVER_IP>
   cd /root/photo-service
   source venv/bin/activate
   pip install -r requirements.txt
   ```

2. Configure environment variables (copy `.env.example` to `.env`):

   ```bash
   DASHSCOPE_API_KEY=your-api-key
   DASHSCOPE_API_KEY_BACKUP=your-backup-api-key
   ```

3. Start the AI service:

   ```bash
   systemctl restart ai-photo
   ```

   Or manually:

   ```bash
   source venv/bin/activate && python3 ai_photo.py
   ```

## Service Management

### Go Backend (photo-service)

```bash
systemctl status photo-service         # Check status
systemctl restart photo-service        # Restart
journalctl -u photo-service -f         # View logs
```

### AI 生图 Service (ai-photo)

```bash
systemctl status ai-photo              # Check status
systemctl restart ai-photo             # Restart
journalctl -u ai-photo -f              # View logs
```

### HivisionIDPhotos

HivisionIDPhotos runs on port 8080 and must be deployed separately as a prerequisite. Verify it is running:

```bash
curl http://127.0.0.1:8080/idphoto
```

### Port verification

```bash
netstat -tlnp | grep -E '8080|8081|8090|8091'
```

Expected:
- `8080` - HivisionIDPhotos
- `8081` - Nginx or alternate entry
- `8090` - (optional, legacy)
- `8091` - AI 生图 Python service

## Nginx Reference

```nginx
location /claw/api/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 180s;
    proxy_connect_timeout 10s;
    proxy_send_timeout 60s;
    client_max_body_size 10m;
}

location /claw/api/data/results/ {
    alias /var/www/html/claw/api/data/results/;
    expires 30d;
    add_header Cache-Control "public, immutable";
}
```

Reload Nginx:

```bash
nginx -t && nginx -s reload
```

## Data Directories

| Path | Purpose |
|------|---------|
| `/var/www/html/claw/api/data/uploads/` | Uploaded user photos |
| `/var/www/html/claw/api/data/results/` | Generated ID photos |
| `/var/www/html/claw/api/data/results.db` | SQLite token database |

## Environment Variables

See `.env.example`:

| Variable | Description |
|----------|-------------|
| `DASHSCOPE_API_KEY` | Aliyun Bailian API key (primary) |
| `DASHSCOPE_API_KEY_BACKUP` | Aliyun Bailian API key (fallback) |
| `ADMIN_PASSWORD` | Admin password for token creation API |

## Quick Health Check

```bash
# Check Go backend
curl http://127.0.0.1:8080/api/verify-token?token=test

# Check AI service
curl http://127.0.0.1:8091/health

# Check Hivision
curl http://127.0.0.1:8080/idphoto
```
