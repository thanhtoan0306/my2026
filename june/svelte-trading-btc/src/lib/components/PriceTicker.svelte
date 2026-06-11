<script lang="ts">
  import { formatFundingRate, formatPercent, formatPrice, formatUsd, shortTime } from '../utils/format'
  import type { MarkPriceData, TickerData } from '../types/trading'

  interface Props {
    ticker: TickerData | null
    mark: MarkPriceData | null
  }

  let { ticker, mark }: Props = $props()

  const price = $derived(mark?.markPrice ?? ticker?.lastPrice ?? 0)
  const change = $derived(ticker?.priceChangePercent ?? 0)
  const isUp = $derived(change >= 0)
</script>

<section class="panel price-panel">
  <div class="price-panel__main">
    <div>
      <p class="label">Mark Price</p>
      {#if price > 0}
        <p class="price {isUp ? 'price--up' : 'price--down'}">{formatUsd(price)}</p>
      {:else}
        <p class="price price--loading">—</p>
      {/if}
    </div>

    {#if ticker}
      <div class="change-badge {isUp ? 'change-badge--up' : 'change-badge--down'}">
        {formatPercent(ticker.priceChangePercent)}
        <span>{formatUsd(ticker.priceChange, 2)}</span>
      </div>
    {/if}
  </div>

  <div class="stats-grid">
    {#if ticker}
      <div class="stat">
        <span>24h High</span>
        <strong>{formatUsd(ticker.high)}</strong>
      </div>
      <div class="stat">
        <span>24h Low</span>
        <strong>{formatUsd(ticker.low)}</strong>
      </div>
      <div class="stat">
        <span>24h Volume</span>
        <strong>{formatPrice(ticker.volume, 0)} BTC</strong>
      </div>
      <div class="stat">
        <span>Quote Vol</span>
        <strong>${formatPrice(ticker.quoteVolume / 1_000_000, 2)}M</strong>
      </div>
    {/if}

    {#if mark}
      <div class="stat">
        <span>Index</span>
        <strong>{formatUsd(mark.indexPrice)}</strong>
      </div>
      <div class="stat">
        <span>Funding</span>
        <strong class={mark.fundingRate >= 0 ? 'text-up' : 'text-down'}>
          {formatFundingRate(mark.fundingRate)}
        </strong>
      </div>
      <div class="stat">
        <span>Next Funding</span>
        <strong>{shortTime(mark.nextFundingTime)}</strong>
      </div>
    {/if}
  </div>
</section>
