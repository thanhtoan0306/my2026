import { get, writable } from 'svelte/store'
import type { Candle, KlineInterval, WsStatus } from '../types/trading'

export const KLINE_INTERVALS: KlineInterval[] = ['1m', '5m', '15m', '1h', '4h', '1d']

const REST_URL = 'https://fapi.binance.com/fapi/v1/klines'
const LIMIT = 500

export const klineInterval = writable<KlineInterval>('15m')
export const candles = writable<Candle[]>([])
export const klinesLoading = writable(true)
export const klineStatus = writable<WsStatus>('connecting')

function parseKlines(raw: unknown[]): Candle[] {
  return raw.map((row) => {
    const r = row as [number, string, string, string, string, string]
    return {
      time: Math.floor(r[0] / 1000),
      open: Number(r[1]),
      high: Number(r[2]),
      low: Number(r[3]),
      close: Number(r[4]),
      volume: Number(r[5]),
    }
  })
}

function parseKlineEvent(data: Record<string, unknown>): Candle {
  const k = data.k as Record<string, unknown>
  return {
    time: Math.floor(Number(k.t) / 1000),
    open: Number(k.o),
    high: Number(k.h),
    low: Number(k.l),
    close: Number(k.c),
    volume: Number(k.v),
  }
}

function fetchHistory(interval: KlineInterval) {
  klinesLoading.set(true)
  fetch(`${REST_URL}?symbol=BTCUSDT&interval=${interval}&limit=${LIMIT}`)
    .then((res) => res.json())
    .then((data) => {
      if (Array.isArray(data)) candles.set(parseKlines(data))
    })
    .finally(() => klinesLoading.set(false))
}

export function startKlinesWs() {
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | undefined
  let closed = false
  let currentInterval: KlineInterval = get(klineInterval)

  const connectWs = (interval: KlineInterval) => {
    if (closed) return
    klineStatus.set('connecting')
    ws?.close()
    ws = new WebSocket(`wss://fstream.binance.com/ws/btcusdt@kline_${interval}`)

    ws.onopen = () => klineStatus.set('connected')

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data as string) as Record<string, unknown>
      if (data.e !== 'kline') return

      const candle = parseKlineEvent(data)
      candles.update((prev) => {
        const idx = prev.findIndex((c) => c.time === candle.time)
        if (idx >= 0) {
          const next = prev.slice()
          next[idx] = candle
          return next
        }
        return [...prev, candle].slice(-LIMIT)
      })
    }

    ws.onerror = () => klineStatus.set('disconnected')

    ws.onclose = () => {
      klineStatus.set('disconnected')
      if (!closed && get(klineInterval) === interval) {
        reconnectTimer = setTimeout(() => connectWs(interval), 3000)
      }
    }
  }

  fetchHistory(currentInterval)
  connectWs(currentInterval)

  const unsub = klineInterval.subscribe((interval) => {
    if (interval === currentInterval) return
    currentInterval = interval
    if (reconnectTimer) clearTimeout(reconnectTimer)
    fetchHistory(interval)
    connectWs(interval)
  })

  return () => {
    closed = true
    unsub()
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
  }
}
