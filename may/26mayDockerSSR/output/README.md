# Offline Docker bundle

Thư mục chứa image Docker đã export — dùng khi **không có mạng** (không cần `docker pull` / Hub).

| File | Mô tả |
|------|--------|
| `docker-ssr-hello.tar` | Image `docker-ssr-hello:offline` (tạo bằng script export) |

## Tạo / cập nhật bản offline (máy có mạng + Docker)

```bash
cd may/26mayDockerSSR
./scripts/export-offline.sh
```

## Chạy offline (máy chỉ cần Docker)

```bash
cd may/26mayDockerSSR
./scripts/run-offline.sh
```

Mở http://localhost:8080
