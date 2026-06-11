<script lang="ts">
  import { openPosition } from '../stores/tradingStore'
  import { formatUsd } from '../utils/format'
  import type { OrderType, Side } from '../types/trading'

  interface Props {
    mark: number
  }

  let { mark }: Props = $props()

  let side = $state<Side>('LONG')
  let orderType = $state<OrderType>('MARKET')
  let size = $state(100)
  let leverage = $state(10)
  let limitPrice = $state('')
  let message = $state<string | null>(null)

  const leverageOptions = [1, 2, 5, 10, 20, 50, 100]
  const margin = $derived(size / leverage)
  const notional = $derived(size * leverage)

  function submit() {
    if (mark <= 0) {
      message = 'Waiting for live price…'
      return
    }

    const entry = orderType === 'LIMIT' && limitPrice ? Number(limitPrice) : mark

    if (!Number.isFinite(entry) || entry <= 0) {
      message = 'Invalid limit price'
      return
    }

    if (orderType === 'LIMIT' && side === 'LONG' && entry > mark) {
      message = 'Limit buy above mark — use market or lower limit'
      return
    }

    if (orderType === 'LIMIT' && side === 'SHORT' && entry < mark) {
      message = 'Limit sell below mark — use market or higher limit'
      return
    }

    const ok = openPosition({ side, entryPrice: entry, size, leverage })
    message = ok ? `${side} opened @ ${formatUsd(entry)}` : 'Insufficient margin'
  }
</script>

<section class="panel trade-form">
  <div class="panel__head">
    <h2>Open Futures</h2>
    <span class="panel__hint">Paper · no real orders</span>
  </div>

  <div class="side-toggle">
    <button
      type="button"
      class="side-toggle__btn side-toggle__btn--long"
      class:active={side === 'LONG'}
      onclick={() => (side = 'LONG')}
    >
      Long
    </button>
    <button
      type="button"
      class="side-toggle__btn side-toggle__btn--short"
      class:active={side === 'SHORT'}
      onclick={() => (side = 'SHORT')}
    >
      Short
    </button>
  </div>

  <div class="type-toggle">
    <button type="button" class:active={orderType === 'MARKET'} onclick={() => (orderType = 'MARKET')}>
      Market
    </button>
    <button type="button" class:active={orderType === 'LIMIT'} onclick={() => (orderType = 'LIMIT')}>
      Limit
    </button>
  </div>

  {#if orderType === 'LIMIT'}
    <label class="field">
      <span>Limit Price (USDT)</span>
      <input
        type="number"
        min="0"
        step="0.01"
        placeholder={mark > 0 ? String(mark) : '0'}
        bind:value={limitPrice}
      />
    </label>
  {/if}

  <label class="field">
    <span>Position Size (USDT)</span>
    <input type="range" min="10" max="5000" step="10" bind:value={size} />
    <div class="field__value">{formatUsd(size)}</div>
  </label>

  <label class="field">
    <span>Leverage</span>
    <div class="leverage-row">
      {#each leverageOptions as x}
        <button
          type="button"
          class="leverage-chip"
          class:active={leverage === x}
          onclick={() => (leverage = x)}
        >
          {x}x
        </button>
      {/each}
    </div>
  </label>

  <div class="trade-summary">
    <div><span>Margin</span><strong>{formatUsd(margin)}</strong></div>
    <div><span>Notional</span><strong>{formatUsd(notional)}</strong></div>
    <div><span>Entry</span><strong>{formatUsd(mark)}</strong></div>
  </div>

  <button
    type="button"
    class="submit-btn submit-btn--{side.toLowerCase()}"
    onclick={submit}
    disabled={mark <= 0}
  >
    {side === 'LONG' ? 'Open Long' : 'Open Short'}
  </button>

  {#if message}
    <p class="form-message">{message}</p>
  {/if}
</section>
