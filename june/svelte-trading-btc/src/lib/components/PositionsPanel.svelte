<script lang="ts">
  import { closePosition, history, positions, resetAccount, unrealizedPnl } from '../stores/tradingStore'
  import { formatUsd, shortTime } from '../utils/format'

  interface Props {
    mark: number
  }

  let { mark }: Props = $props()
</script>

<section class="panel positions-panel">
  <div class="panel__head">
    <h2>Positions</h2>
    <button type="button" class="link-btn" onclick={resetAccount}>Reset account</button>
  </div>

  {#if $positions.length === 0}
    <p class="empty-state">No open positions</p>
  {:else}
    <div class="positions-list">
      {#each $positions as pos (pos.id)}
        {@const pnl = unrealizedPnl(pos, mark)}
        {@const pnlPct = mark > 0 ? (pnl / (pos.size / pos.leverage)) * 100 : 0}
        <article class="position-card position-card--{pos.side.toLowerCase()}">
          <div class="position-card__head">
            <strong>{pos.side}</strong>
            <span>{pos.leverage}x</span>
          </div>
          <div class="position-card__grid">
            <div><span>Size</span><strong>{formatUsd(pos.size)}</strong></div>
            <div><span>Entry</span><strong>{formatUsd(pos.entryPrice)}</strong></div>
            <div><span>Mark</span><strong>{formatUsd(mark)}</strong></div>
            <div>
              <span>uPnL</span>
              <strong class={pnl >= 0 ? 'text-up' : 'text-down'}>
                {formatUsd(pnl)} ({pnlPct.toFixed(2)}%)
              </strong>
            </div>
          </div>
          <div class="position-card__foot">
            <span>{shortTime(pos.openedAt)}</span>
            <button
              type="button"
              class="close-btn"
              onclick={() => closePosition(pos.id, mark)}
              disabled={mark <= 0}
            >
              Close
            </button>
          </div>
        </article>
      {/each}
    </div>
  {/if}

  {#if $history.length > 0}
    <div class="history-block">
      <h3>History</h3>
      {#each $history.slice(0, 8) as trade (trade.id + String(trade.closedAt))}
        <div class="history-row">
          <span class={trade.side === 'LONG' ? 'text-up' : 'text-down'}>{trade.side}</span>
          <span>{formatUsd(trade.entryPrice)} → {formatUsd(trade.exitPrice)}</span>
          <span class={trade.pnl >= 0 ? 'text-up' : 'text-down'}>{formatUsd(trade.pnl)}</span>
        </div>
      {/each}
    </div>
  {/if}
</section>
