import { For, Show } from 'solid-js'
import type { OrderBookData } from '../types/trading'
import { formatPrice, formatQty } from '../utils/format'

interface OrderBookProps {
  orderBook: () => OrderBookData
  markPrice: () => number
}

export function OrderBook(props: OrderBookProps) {
  const maxQty = () => {
    const book = props.orderBook()
    const all = [...book.bids, ...book.asks].map((l) => l.quantity)
    return Math.max(...all, 0.0001)
  }

  return (
    <section class="panel">
      <div class="panel__head">
        <h2>Order Book</h2>
        <span class="panel__hint">depth 20 · 100ms</span>
      </div>

      <div class="orderbook">
        <div class="orderbook__header">
          <span>Price (USDT)</span>
          <span>Size (BTC)</span>
        </div>

        <div class="orderbook__asks">
          <For each={[...props.orderBook().asks].reverse().slice(0, 12)}>
            {(level) => (
              <div class="orderbook__row orderbook__row--ask">
                <div
                  class="orderbook__bar orderbook__bar--ask"
                  style={{ width: `${(level.quantity / maxQty()) * 100}%` }}
                />
                <span>{formatPrice(level.price)}</span>
                <span>{formatQty(level.quantity)}</span>
              </div>
            )}
          </For>
        </div>

        <Show when={props.markPrice() > 0}>
          <div class="orderbook__mid">
            <strong>{formatPrice(props.markPrice())}</strong>
            <span>Mark</span>
          </div>
        </Show>

        <div class="orderbook__bids">
          <For each={props.orderBook().bids.slice(0, 12)}>
            {(level) => (
              <div class="orderbook__row orderbook__row--bid">
                <div
                  class="orderbook__bar orderbook__bar--bid"
                  style={{ width: `${(level.quantity / maxQty()) * 100}%` }}
                />
                <span>{formatPrice(level.price)}</span>
                <span>{formatQty(level.quantity)}</span>
              </div>
            )}
          </For>
        </div>
      </div>
    </section>
  )
}
