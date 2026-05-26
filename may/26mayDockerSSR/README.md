# Docker SSR Hello World (Go)

Trang HTML render trên server (Go + `html/template`), kèm hướng dẫn 3 bước ngay trên UI.

## Bước 1 — Khởi tạo lần đầu

```bash
cd may/26mayDockerSSR
cp .env.example .env
# Sửa DOCKER_IMAGE=tenban/docker-ssr-hello:latest trong .env

docker compose up --build
```

Mở [http://localhost:8080](http://localhost:8080) — trang hiển thị đủ 3 bước (khởi tạo, chạy lại, Docker Hub).

## Bước 2 — Chạy lần sau

```bash
cd may/26mayDockerSSR
docker compose up -d
docker compose logs -f
```

Dừng: `docker compose down`

Chạy từ image đã push (máy khác):

```bash
docker pull your-dockerhub-user/docker-ssr-hello:latest
docker run -d -p 8080:8080 --name ssr-hello your-dockerhub-user/docker-ssr-hello:latest
```

## Bước 3 — Đẩy lên Docker Hub

```bash
cd may/26mayDockerSSR
docker login
docker build -t your-dockerhub-user/docker-ssr-hello:latest .
docker push your-dockerhub-user/docker-ssr-hello:latest
```

Thay `your-dockerhub-user` bằng username Docker Hub của bạn (cùng giá trị `DOCKER_IMAGE` trong `.env`).

## Offline — `output/docker-ssr-hello.tar`

Export image vào `output/` để mang sang máy không có mạng:

```bash
cd may/26mayDockerSSR
chmod +x scripts/*.sh
./scripts/export-offline.sh    # tạo output/docker-ssr-hello.tar
```

Máy offline (chỉ cần Docker Desktop):

```bash
cd may/26mayDockerSSR
./scripts/run-offline.sh
```

Chi tiết: [output/README.md](output/README.md)

## Case: Docker daemon chưa chạy

Lỗi thường gặp:

```text
Cannot connect to the Docker daemon at unix:///Users/.../.docker/run/docker.sock. Is the docker daemon running?
```

**Cách xử lý:**

```bash
open -a Docker          # mở Docker Desktop, đợi daemon sẵn sàng
docker info             # kiểm tra kết nối OK
cd may/26mayDockerSSR
docker compose up -d
```

Gộp một lệnh:

```bash
open -a Docker && sleep 5 && cd may/26mayDockerSSR && docker compose up -d
```

## Chạy local (không Docker)

```bash
cd may/26mayDockerSSR
export DOCKER_IMAGE=your-dockerhub-user/docker-ssr-hello:latest
go run .
```

## Biến môi trường

| Biến | Mặc định | Mô tả |
|------|----------|--------|
| `PORT` | `8080` | Cổng HTTP |
| `DOCKER_IMAGE` | `your-dockerhub-user/docker-ssr-hello:latest` | Tên image hiển thị trên UI |
| `PROJECT_DIR` | `may/26mayDockerSSR` | Đường dẫn trong lệnh mẫu trên UI |
| `OFFLINE_IMAGE` | `docker-ssr-hello:offline` | Tag image trong file tar |
| `OFFLINE_TAR` | `output/docker-ssr-hello.tar` | Đường dẫn file export offline |
