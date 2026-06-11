import { createSignal, onCleanup, onMount } from 'solid-js'
import type {
  AggTrade,
  MarkPriceData,
  MarketSnapshot,
  OrderBookData,
  TickerData,
  WsStatus,
} from '../types/trading'

const STREAMS = [
  'btcusdt@markPrice@1s',
  'btcusdt@ticker',
  'btcusdt@depth20@100ms',
  'btcusdt@aggTrade',
].join('/')

const WS_URL = `wss://fstream.binance.com/stream?streams=${STREAMS}`
const MAX_TRADES = 40

function parseLevels(levels: [string, string][]): OrderBookData['bids'] {
  return levels.map(([price, qty]) => ({
    price: Number(price),
    quantity: Number(qty),
  }))
}

export function useBinanceFuturesWs() {
  const [status, setStatus] = createSignal<WsStatus>('connecting')
  const [ticker, setTicker] = createSignal<TickerData | null>(null)
  const [markPrice, setMarkPrice] = createSignal<MarkPriceData | null>(null)
  const [orderBook, setOrderBook] = createSignal<OrderBookData>({ bids: [], asks: [] })
  const [recentTrades, setRecentTrades] = createSignal<AggTrade[]>([])

  onMount(() => {
    let ws: WebSocket | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | undefined
    let closed = false

    const connect = () => {
      if (closed) return
      setStatus('connecting')
      ws = new WebSocket(WS_URL)

      ws.onopen = () => setStatus('connected')

      ws.onmessage = (event) => {
        const payload = JSON.parse(event.data as string) as { stream: string; data: Record<string, unknown> }
        const { stream, data } = payload

        if (stream === 'btcusdt@markPrice@1s') {
          setMarkPrice({
            symbol: String(data.s),
            markPrice: Number(data.p),
            indexPrice: Number(data.i),
            fundingRate: Number(data.r),
            nextFundingTime: Number(data.T),
          })
          return
        }

        if (stream === 'btcusdt@ticker') {
          setTicker({
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
          setOrderBook({
            bids: parseLevels(data.b as [string, string][]),
            asks: parseLevels(data.a as [string, string][]),
          })
          return
        }

        if (stream === 'btcusdt@aggTrade') {
          const trade: AggTrade = {
            id: Number(data.a),
            price: Number(data.p),
            quantity: Number(data.q),
            time: Number(data.T),
            isBuyerMaker: Boolean(data.m),
          }
          setRecentTrades((prev) => [trade, ...prev].slice(0, MAX_TRADES))
        }
      }

      ws.onerror = () => setStatus('disconnected')

      ws.onclose = () => {
        setStatus('disconnected')
        if (!closed) {
          reconnectTimer = setTimeout(connect, 3000)
        }
      }
    }

    connect()

    onCleanup(() => {
      closed = true
      if (reconnectTimer) clearTimeout(reconnectTimer)
      ws?.close()
    })
  })

  const snapshot = (): MarketSnapshot => ({
    status: status(),
    ticker: ticker(),
    markPrice: markPrice(),
    orderBook: orderBook(),
    recentTrades: recentTrades(),
  })

  return {
    status,
    ticker,
    markPrice,
    orderBook,
    recentTrades,
    snapshot,
  }
}
