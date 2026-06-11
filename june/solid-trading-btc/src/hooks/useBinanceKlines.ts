import { createEffect, createSignal, onCleanup } from 'solid-js'
import type { Candle, KlineInterval, WsStatus } from '../types/trading'

export const KLINE_INTERVALS: KlineInterval[] = ['1m', '5m', '15m', '1h', '4h', '1d']

const REST_URL = 'https://fapi.binance.com/fapi/v1/klines'
const LIMIT = 500

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

export function useBinanceKlines(initialInterval: KlineInterval = '15m') {
  const [interval, setInterval] = createSignal<KlineInterval>(initialInterval)
  const [candles, setCandles] = createSignal<Candle[]>([])
  const [loading, setLoading] = createSignal(true)
  const [status, setStatus] = createSignal<WsStatus>('connecting')

  createEffect(() => {
    const iv = interval()
    let cancelled = false
    setLoading(true)

    fetch(`${REST_URL}?symbol=BTCUSDT&interval=${iv}&limit=${LIMIT}`)
      .then((res) => res.json())
      .then((data) => {
        if (cancelled || !Array.isArray(data)) return
        setCandles(parseKlines(data))
        setLoading(false)
      })
      .catch(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  })

  createEffect(() => {
    const iv = interval()
    let ws: WebSocket | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | undefined
    let closed = false

    const connect = () => {
      if (closed) return
      setStatus('connecting')
      ws = new WebSocket(`wss://fstream.binance.com/ws/btcusdt@kline_${iv}`)

      ws.onopen = () => setStatus('connected')

      ws.onmessage = (event) => {
        const data = JSON.parse(event.data as string) as Record<string, unknown>
        if (data.e !== 'kline') return

        const candle = parseKlineEvent(data)
        setCandles((prev) => {
          const idx = prev.findIndex((c) => c.time === candle.time)
          if (idx >= 0) {
            const next = prev.slice()
            next[idx] = candle
            return next
          }
          return [...prev, candle].slice(-LIMIT)
        })
      }

      ws.onerror = () => setStatus('disconnected')

      ws.onclose = () => {
        setStatus('disconnected')
        if (!closed) reconnectTimer = setTimeout(connect, 3000)
      }
    }

    connect()

    onCleanup(() => {
      closed = true
      if (reconnectTimer) clearTimeout(reconnectTimer)
      ws?.close()
    })
  })

  return { interval, setInterval, candles, loading, status, intervals: KLINE_INTERVALS }
}
