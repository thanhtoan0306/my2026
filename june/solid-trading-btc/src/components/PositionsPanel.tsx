import { For, Show } from 'solid-js'
import { tradingStore } from '../stores/tradingStore'
import { formatUsd, shortTime } from '../utils/format'

interface PositionsPanelProps {
  markPrice: () => number
}

export function PositionsPanel(props: PositionsPanelProps) {
  return (
    <section class="panel positions-panel">
      <div class="panel__head">
        <h2>Positions</h2>
        <button type="button" class="link-btn" onClick={() => tradingStore.resetAccount()}>
          Reset account
        </button>
      </div>

      <Show
        when={tradingStore.state.positions.length > 0}
        fallback={<p class="empty-state">No open positions</p>}
      >
        <div class="positions-list">
          <For each={tradingStore.state.positions}>
            {(pos) => {
              const pnl = () => tradingStore.unrealizedPnl(pos, props.markPrice())
              const pnlPct = () =>
                props.markPrice() > 0
                  ? (pnl() / (pos.size / pos.leverage)) * 100
                  : 0

              return (
                <article class={`position-card position-card--${pos.side.toLowerCase()}`}>
                  <div class="position-card__head">
                    <strong>{pos.side}</strong>
                    <span>{pos.leverage}x</span>
                  </div>
                  <div class="position-card__grid">
                    <div><span>Size</span><strong>{formatUsd(pos.size)}</strong></div>
                    <div><span>Entry</span><strong>{formatUsd(pos.entryPrice)}</strong></div>
                    <div><span>Mark</span><strong>{formatUsd(props.markPrice())}</strong></div>
                    <div>
                      <span>uPnL</span>
                      <strong class={pnl() >= 0 ? 'text-up' : 'text-down'}>
                        {formatUsd(pnl())} ({pnlPct().toFixed(2)}%)
                      </strong>
                    </div>
                  </div>
                  <div class="position-card__foot">
                    <span>{shortTime(pos.openedAt)}</span>
                    <button
                      type="button"
                      class="close-btn"
                      onClick={() => tradingStore.closePosition(pos.id, props.markPrice())}
                      disabled={props.markPrice() <= 0}
                    >
                      Close
                    </button>
                  </div>
                </article>
              )
            }}
          </For>
        </div>
      </Show>

      <Show when={tradingStore.state.history.length > 0}>
        <div class="history-block">
          <h3>History</h3>
          <For each={tradingStore.state.history.slice(0, 8)}>
            {(trade) => (
              <div class="history-row">
                <span class={trade.side === 'LONG' ? 'text-up' : 'text-down'}>{trade.side}</span>
                <span>{formatUsd(trade.entryPrice)} → {formatUsd(trade.exitPrice)}</span>
                <span class={trade.pnl >= 0 ? 'text-up' : 'text-down'}>{formatUsd(trade.pnl)}</span>
              </div>
            )}
          </For>
        </div>
      </Show>
    </section>
  )
}
