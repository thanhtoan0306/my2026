# BTC Futures Trading (Svelte)

Phiên bản **Svelte** của app paper trading BTCUSDT Perpetual Futures — tương đương `solid-trading-btc`.

## Chạy app

```bash
npm install
npm run dev
```

Mở http://localhost:5173

## Stack

- Svelte 5 + Vite + TypeScript
- Binance Futures WebSocket (mark, ticker, depth, trades, kline)
- TradingView lightweight-charts (candlestick + volume)

## Tính năng

- Chart nến realtime (1m · 5m · 15m · 1h · 4h · 1d)
- Order book, recent trades, mark price / funding
- Paper Long/Short với leverage, uPnL, lịch sử lệnh
- Balance demo $10,000

## Cấu trúc

| Path | Mô tả |
|------|--------|
| `src/lib/binance/futuresWs.ts` | WebSocket market data |
| `src/lib/binance/klines.ts` | REST + kline WebSocket |
| `src/lib/stores/tradingStore.ts` | Paper trading state |
| `src/lib/components/` | UI components |
