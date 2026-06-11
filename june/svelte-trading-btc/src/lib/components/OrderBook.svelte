<script lang="ts">
  import { formatPrice, formatQty } from '../utils/format'
  import type { OrderBookData } from '../types/trading'

  interface Props {
    book: OrderBookData
    mark: number
  }

  let { book, mark }: Props = $props()

  const maxQty = $derived(
    Math.max(...[...book.bids, ...book.asks].map((l) => l.quantity), 0.0001),
  )

  const asks = $derived([...book.asks].reverse().slice(0, 12))
</script>

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
      {#each asks as level (level.price)}
        <div class="orderbook__row orderbook__row--ask">
          <div
            class="orderbook__bar orderbook__bar--ask"
            style:width="{(level.quantity / maxQty) * 100}%"
          ></div>
          <span>{formatPrice(level.price)}</span>
          <span>{formatQty(level.quantity)}</span>
        </div>
      {/each}
    </div>

    {#if mark > 0}
      <div class="orderbook__mid">
        <strong>{formatPrice(mark)}</strong>
        <span>Mark</span>
      </div>
    {/if}

    <div class="orderbook__bids">
      {#each book.bids.slice(0, 12) as level (level.price)}
        <div class="orderbook__row orderbook__row--bid">
          <div
            class="orderbook__bar orderbook__bar--bid"
            style:width="{(level.quantity / maxQty) * 100}%"
          ></div>
          <span>{formatPrice(level.price)}</span>
          <span>{formatQty(level.quantity)}</span>
        </div>
      {/each}
    </div>
  </div>
</section>
