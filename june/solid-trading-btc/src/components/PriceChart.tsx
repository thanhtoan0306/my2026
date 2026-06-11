import {
  CandlestickSeries,
  ColorType,
  CrosshairMode,
  createChart,
  HistogramSeries,
  type IChartApi,
  type ISeriesApi,
  type UTCTimestamp,
} from 'lightweight-charts'
import { createEffect, For, onCleanup, onMount, Show } from 'solid-js'
import type { Candle, KlineInterval, WsStatus } from '../types/trading'

interface PriceChartProps {
  candles: () => Candle[]
  loading: () => boolean
  status: () => WsStatus
  interval: () => KlineInterval
  setInterval: (value: KlineInterval) => void
  intervals: KlineInterval[]
}

export function PriceChart(props: PriceChartProps) {
  let containerRef: HTMLDivElement | undefined
  let chart: IChartApi | undefined
  let candleSeries: ISeriesApi<'Candlestick'> | undefined
  let volumeSeries: ISeriesApi<'Histogram'> | undefined
  let seededInterval: KlineInterval | undefined

  onMount(() => {
    if (!containerRef) return

    chart = createChart(containerRef, {
      layout: {
        background: { type: ColorType.Solid, color: '#121821' },
        textColor: '#8b98a9',
      },
      grid: {
        vertLines: { color: 'rgba(30, 39, 51, 0.8)' },
        horzLines: { color: 'rgba(30, 39, 51, 0.8)' },
      },
      crosshair: { mode: CrosshairMode.Normal },
      rightPriceScale: { borderColor: '#1e2733' },
      timeScale: {
        borderColor: '#1e2733',
        timeVisible: true,
        secondsVisible: props.interval() === '1m',
      },
      width: containerRef.clientWidth,
      height: 420,
    })

    candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#00c087',
      downColor: '#ff5b5b',
      borderVisible: false,
      wickUpColor: '#00c087',
      wickDownColor: '#ff5b5b',
    })

    volumeSeries = chart.addSeries(HistogramSeries, {
      priceFormat: { type: 'volume' },
      priceScaleId: 'volume',
    })

    chart.priceScale('volume').applyOptions({
      scaleMargins: { top: 0.82, bottom: 0 },
    })

    const resizeObserver = new ResizeObserver(() => {
      if (containerRef && chart) {
        chart.applyOptions({ width: containerRef.clientWidth })
      }
    })
    resizeObserver.observe(containerRef)

    onCleanup(() => {
      resizeObserver.disconnect()
      chart?.remove()
    })
  })

  createEffect(() => {
    props.interval()
    seededInterval = undefined
  })

  createEffect(() => {
    const data = props.candles()
    const iv = props.interval()
    if (!candleSeries || !volumeSeries || data.length === 0) return

    const candleData = data.map((c) => ({
      time: c.time as UTCTimestamp,
      open: c.open,
      high: c.high,
      low: c.low,
      close: c.close,
    }))

    const volumeData = data.map((c) => ({
      time: c.time as UTCTimestamp,
      value: c.volume,
      color: c.close >= c.open ? 'rgba(0, 192, 135, 0.45)' : 'rgba(255, 91, 91, 0.45)',
    }))

    if (seededInterval !== iv) {
      candleSeries.setData(candleData)
      volumeSeries.setData(volumeData)
      chart?.timeScale().fitContent()
      seededInterval = iv
      return
    }

    const last = data[data.length - 1]
    candleSeries.update({
      time: last.time as UTCTimestamp,
      open: last.open,
      high: last.high,
      low: last.low,
      close: last.close,
    })
    volumeSeries.update({
      time: last.time as UTCTimestamp,
      value: last.volume,
      color: last.close >= last.open ? 'rgba(0, 192, 135, 0.45)' : 'rgba(255, 91, 91, 0.45)',
    })
  })

  createEffect(() => {
    chart?.timeScale().applyOptions({
      secondsVisible: props.interval() === '1m',
    })
  })

  return (
    <section class="panel chart-panel">
      <div class="panel__head">
        <div class="chart-panel__title">
          <h2>BTCUSDT Perp</h2>
          <span class={`status-dot status-dot--${props.status()}`} />
        </div>
        <div class="interval-toggle">
          <For each={props.intervals}>
            {(iv) => (
              <button
                type="button"
                class={props.interval() === iv ? 'active' : ''}
                onClick={() => props.setInterval(iv)}
              >
                {iv}
              </button>
            )}
          </For>
        </div>
      </div>

      <div class="chart-wrap">
        <div class="chart-container" ref={containerRef} />
        <Show when={props.loading()}>
          <div class="chart-loading">Loading candles…</div>
        </Show>
      </div>
    </section>
  )
}
