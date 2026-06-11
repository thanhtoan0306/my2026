<script lang="ts">
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
  import { onMount } from 'svelte'
  import { KLINE_INTERVALS, klineInterval, klineStatus, klinesLoading } from '../binance/klines'
  import type { Candle, KlineInterval, WsStatus } from '../types/trading'

  interface Props {
    data: Candle[]
    loading: boolean
    status: WsStatus
  }

  let { data, loading, status }: Props = $props()

  let container = $state<HTMLDivElement | null>(null)
  let chart: IChartApi | undefined
  let candleSeries: ISeriesApi<'Candlestick'> | undefined
  let volumeSeries: ISeriesApi<'Histogram'> | undefined
  let seededInterval: KlineInterval | undefined

  onMount(() => {
    if (!container) return

    chart = createChart(container, {
      layout: {
        background: { type: ColorType.Solid, color: '#15102a' },
        textColor: '#9d8ec7',
      },
      grid: {
        vertLines: { color: 'rgba(45, 33, 80, 0.85)' },
        horzLines: { color: 'rgba(45, 33, 80, 0.85)' },
      },
      crosshair: { mode: CrosshairMode.Normal },
      rightPriceScale: { borderColor: '#2d2150' },
      timeScale: {
        borderColor: '#2d2150',
        timeVisible: true,
        secondsVisible: false,
      },
      width: container.clientWidth,
      height: 420,
    })

    candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#34d399',
      downColor: '#fb7185',
      borderVisible: false,
      wickUpColor: '#2dd4bf',
      wickDownColor: '#f472b6',
    })

    volumeSeries = chart.addSeries(HistogramSeries, {
      priceFormat: { type: 'volume' },
      priceScaleId: 'volume',
    })

    chart.priceScale('volume').applyOptions({
      scaleMargins: { top: 0.82, bottom: 0 },
    })

    const resizeObserver = new ResizeObserver(() => {
      if (container && chart) {
        chart.applyOptions({ width: container.clientWidth })
      }
    })
    resizeObserver.observe(container)

    return () => {
      resizeObserver.disconnect()
      chart?.remove()
    }
  })

  $effect(() => {
    $klineInterval
    seededInterval = undefined
  })

  $effect(() => {
    if (!candleSeries || !volumeSeries || data.length === 0) return

    const iv = $klineInterval
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
      color: c.close >= c.open ? 'rgba(52, 211, 153, 0.45)' : 'rgba(251, 113, 133, 0.45)',
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
      color: last.close >= last.open ? 'rgba(52, 211, 153, 0.45)' : 'rgba(251, 113, 133, 0.45)',
    })
  })

  $effect(() => {
    chart?.timeScale().applyOptions({
      secondsVisible: $klineInterval === '1m',
    })
  })
</script>

<section class="panel chart-panel chart-panel--svelte">
  <div class="panel__head">
    <div class="chart-panel__title">
      <h2>BTCUSDT Perp</h2>
      <span class="chart-svelte-tag">Svelte</span>
      <span class="status-dot status-dot--{status}"></span>
    </div>
    <div class="interval-toggle">
      {#each KLINE_INTERVALS as iv}
        <button
          type="button"
          class:active={$klineInterval === iv}
          onclick={() => klineInterval.set(iv)}
        >
          {iv}
        </button>
      {/each}
    </div>
  </div>

  <div class="chart-wrap">
    <div class="chart-container" bind:this={container}></div>
    {#if loading}
      <div class="chart-loading">Loading candles…</div>
    {/if}
  </div>
</section>

<style>
  .chart-panel--svelte {
    border-color: rgba(255, 62, 0, 0.22);
    background:
      linear-gradient(180deg, rgba(124, 58, 237, 0.06) 0%, transparent 40%),
      var(--panel);
  }

  .chart-svelte-tag {
    font-size: 0.62rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 0.15rem 0.4rem;
    border-radius: 0.3rem;
    color: #ff3e00;
    border: 1px solid rgba(255, 62, 0, 0.45);
    background: rgba(255, 62, 0, 0.08);
  }
</style>
