import { For } from 'solid-js'
import type { AggTrade } from '../types/trading'
import { formatPrice, formatQty, formatTime } from '../utils/format'

interface RecentTradesProps {
  trades: () => AggTrade[]
}

export function RecentTrades(props: RecentTradesProps) {
  return (
    <section class="panel">
      <div class="panel__head">
        <h2>Recent Trades</h2>
        <span class="panel__hint">aggTrade stream</span>
      </div>

      <div class="trades-table">
        <div class="trades-table__header">
          <span>Price</span>
          <span>Qty</span>
          <span>Time</span>
        </div>
        <For each={props.trades()}>
          {(trade) => (
            <div class={`trades-table__row ${trade.isBuyerMaker ? 'trades-table__row--sell' : 'trades-table__row--buy'}`}>
              <span>{formatPrice(trade.price)}</span>
              <span>{formatQty(trade.quantity)}</span>
              <span>{formatTime(trade.time)}</span>
            </div>
          )}
        </For>
      </div>
    </section>
  )
}
