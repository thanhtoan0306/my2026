import { Show } from 'solid-js'
import type { MarkPriceData, TickerData } from '../types/trading'
import { formatFundingRate, formatPercent, formatPrice, formatUsd, shortTime } from '../utils/format'

interface PriceTickerProps {
  ticker: () => TickerData | null
  markPrice: () => MarkPriceData | null
}

export function PriceTicker(props: PriceTickerProps) {
  const price = () => props.markPrice()?.markPrice ?? props.ticker()?.lastPrice ?? 0
  const change = () => props.ticker()?.priceChangePercent ?? 0
  const isUp = () => change() >= 0

  return (
    <section class="panel price-panel">
      <div class="price-panel__main">
        <div>
          <p class="label">Mark Price</p>
          <Show when={price() > 0} fallback={<p class="price price--loading">—</p>}>
            <p class={`price ${isUp() ? 'price--up' : 'price--down'}`}>
              {formatUsd(price())}
            </p>
          </Show>
        </div>

        <Show when={props.ticker()}>
          {(t) => (
            <div class={`change-badge ${isUp() ? 'change-badge--up' : 'change-badge--down'}`}>
              {formatPercent(t().priceChangePercent)}
              <span>{formatUsd(t().priceChange, 2)}</span>
            </div>
          )}
        </Show>
      </div>

      <div class="stats-grid">
        <Show when={props.ticker()}>
          {(t) => (
            <>
              <div class="stat">
                <span>24h High</span>
                <strong>{formatUsd(t().high)}</strong>
              </div>
              <div class="stat">
                <span>24h Low</span>
                <strong>{formatUsd(t().low)}</strong>
              </div>
              <div class="stat">
                <span>24h Volume</span>
                <strong>{formatPrice(t().volume, 0)} BTC</strong>
              </div>
              <div class="stat">
                <span>Quote Vol</span>
                <strong>${formatPrice(t().quoteVolume / 1_000_000, 2)}M</strong>
              </div>
            </>
          )}
        </Show>

        <Show when={props.markPrice()}>
          {(m) => (
            <>
              <div class="stat">
                <span>Index</span>
                <strong>{formatUsd(m().indexPrice)}</strong>
              </div>
              <div class="stat">
                <span>Funding</span>
                <strong class={m().fundingRate >= 0 ? 'text-up' : 'text-down'}>
                  {formatFundingRate(m().fundingRate)}
                </strong>
              </div>
              <div class="stat">
                <span>Next Funding</span>
                <strong>{shortTime(m().nextFundingTime)}</strong>
              </div>
            </>
          )}
        </Show>
      </div>
    </section>
  )
}
