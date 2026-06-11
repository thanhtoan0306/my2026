# BTC Futures Trading (SolidJS)

Paper trading app cho **BTCUSDT Perpetual Futures** với dữ liệu realtime qua Binance WebSocket.

## Chạy app

```bash
npm install
npm run dev
```

Mở http://localhost:5173

## WebSocket streams

Kết nối `wss://fstream.binance.com` với các stream:

- `btcusdt@markPrice@1s` — mark price, index, funding rate
- `btcusdt@ticker` — 24h ticker
- `btcusdt@depth20@100ms` — order book
- `btcusdt@aggTrade` — recent trades

## Tính năng

- Giá mark / index / funding cập nhật realtime
- Order book + recent trades
- **Candlestick chart** (TradingView lightweight-charts) với volume, interval 1m–1d, cập nhật qua kline WebSocket
- Mở Long/Short futures (paper) với leverage 1x–100x
- Theo dõi uPnL và lịch sử lệnh
- Balance demo $10,000 (không gửi lệnh thật lên sàn)
