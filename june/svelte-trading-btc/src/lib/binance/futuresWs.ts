import { writable } from 'svelte/store'
import type { AggTrade, MarkPriceData, OrderBookData, TickerData, WsStatus } from '../types/trading'

const STREAMS = [
  'btcusdt@markPrice@1s',
  'btcusdt@ticker',
  'btcusdt@depth20@100ms',
  'btcusdt@aggTrade',
].join('/')

const WS_URL = `wss://fstream.binance.com/stream?streams=${STREAMS}`
const MAX_TRADES = 40

function parseLevels(levels: [string, string][]) {
  return levels.map(([price, qty]) => ({
    price: Number(price),
    quantity: Number(qty),
  }))
}

export const wsStatus = writable<WsStatus>('connecting')
export const ticker = writable<TickerData | null>(null)
export const markPrice = writable<MarkPriceData | null>(null)
export const orderBook = writable<OrderBookData>({ bids: [], asks: [] })
export const recentTrades = writable<AggTrade[]>([])

export function startFuturesWs() {
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | undefined
  let closed = false

  const connect = () => {
    if (closed) return
    wsStatus.set('connecting')
    ws = new WebSocket(WS_URL)

    ws.onopen = () => wsStatus.set('connected')

    ws.onmessage = (event) => {
      const payload = JSON.parse(event.data as string) as {
        stream: string
        data: Record<string, unknown>
      }
      const { stream, data } = payload

      if (stream === 'btcusdt@markPrice@1s') {
        markPrice.set({
          symbol: String(data.s),
          markPrice: Number(data.p),
          indexPrice: Number(data.i),
          fundingRate: Number(data.r),
          nextFundingTime: Number(data.T),
        })
        return
      }

      if (stream === 'btcusdt@ticker') {
        ticker.set({
          symbol: String(data.s),
          lastPrice: Number(data.c),
          priceChange: Number(data.p),
          priceChangePercent: Number(data.P),
          high: Number(data.h),
          low: Number(data.l),
          volume: Number(data.v),
          quoteVolume: Number(data.q),
        })
        return
      }

      if (stream === 'btcusdt@depth20@100ms') {
        orderBook.set({
          bids: parseLevels(data.b as [string, string][]),
          asks: parseLevels(data.a as [string, string][]),
        })
        return
      }

      if (stream === 'btcusdt@aggTrade') {
        recentTrades.update((prev) => [
          {
            id: Number(data.a),
            price: Number(data.p),
            quantity: Number(data.q),
            time: Number(data.T),
            isBuyerMaker: Boolean(data.m),
          },
          ...prev,
        ].slice(0, MAX_TRADES))
      }
    }

    ws.onerror = () => wsStatus.set('disconnected')

    ws.onclose = () => {
      wsStatus.set('disconnected')
      if (!closed) reconnectTimer = setTimeout(connect, 3000)
    }
  }

  connect()

  return () => {
    closed = true
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
  }
}
