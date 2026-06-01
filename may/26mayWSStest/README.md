# WebSocket Test (Go SSR)

Trang HTML render trên server (Go + `html/template`) với ô nhập **WebSocket URL**, nút Connect/Disconnect/Send và log realtime.

Server có sẵn endpoint echo tại `/ws` — gửi gì nhận lại `echo: …`.

## Chạy local

```bash
cd may/26mayWSStest
go mod tidy
go run .
```

Mở [http://localhost:8080](http://localhost:8080):

1. URL mặc định: `ws://localhost:8080/ws`
2. **Connect** → **Send** (vd. `ping`) → log hiển thị `echo: ping`
3. Dán URL khác (public echo, OKX, …) để test endpoint ngoài

**Card Binance:** Connect tới `wss://stream.binance.com:9443/ws/btcusdt@ticker` — parse JSON `24hrTicker` (giá, %, high/low, volume, bid/ask).

## Docker (tùy chọn)

```bash
cd may/26mayWSStest
docker compose up --build
```

## Biến môi trường

| Biến | Mặc định | Mô tả |
|------|----------|--------|
| `PORT` | `8080` | Cổng HTTP |
| `DEFAULT_WS_URL` | `ws(s)://<host>/ws` | URL hiển thị sẵn trong input (SSR) |

## Lưu ý

- Trình duyệt chỉ kết nối WS từ **client** (JavaScript); Go server chỉ render HTML + cung cấp `/ws` echo.
- Trang `https` không kết nối được `ws://` (mixed content) — dùng `wss://`.
- Một số server ngoài chặn origin; lỗi hiện trong log và tab Network.
