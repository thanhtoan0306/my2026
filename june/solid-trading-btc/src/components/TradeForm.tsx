import { createSignal, For, Show } from 'solid-js'
import type { OrderType, Side } from '../types/trading'
import { tradingStore } from '../stores/tradingStore'
import { formatUsd } from '../utils/format'

interface TradeFormProps {
  markPrice: () => number
}

export function TradeForm(props: TradeFormProps) {
  const [side, setSide] = createSignal<Side>('LONG')
  const [orderType, setOrderType] = createSignal<OrderType>('MARKET')
  const [size, setSize] = createSignal(100)
  const [leverage, setLeverage] = createSignal(10)
  const [limitPrice, setLimitPrice] = createSignal('')
  const [message, setMessage] = createSignal<string | null>(null)

  const margin = () => size() / leverage()
  const notional = () => size() * leverage()

  const submit = () => {
    const price = props.markPrice()
    if (price <= 0) {
      setMessage('Waiting for live price…')
      return
    }

    const entry =
      orderType() === 'LIMIT' && limitPrice()
        ? Number(limitPrice())
        : price

    if (!Number.isFinite(entry) || entry <= 0) {
      setMessage('Invalid limit price')
      return
    }

    if (orderType() === 'LIMIT' && side() === 'LONG' && entry > price) {
      setMessage('Limit buy above mark — use market or lower limit')
      return
    }

    if (orderType() === 'LIMIT' && side() === 'SHORT' && entry < price) {
      setMessage('Limit sell below mark — use market or higher limit')
      return
    }

    const ok = tradingStore.openPosition({
      side: side(),
      entryPrice: entry,
      size: size(),
      leverage: leverage(),
    })

    setMessage(ok ? `${side()} opened @ ${formatUsd(entry)}` : 'Insufficient margin')
  }

  return (
    <section class="panel trade-form">
      <div class="panel__head">
        <h2>Open Futures</h2>
        <span class="panel__hint">Paper · no real orders</span>
      </div>

      <div class="side-toggle">
        <button
          type="button"
          class={`side-toggle__btn ${side() === 'LONG' ? 'side-toggle__btn--long active' : ''}`}
          onClick={() => setSide('LONG')}
        >
          Long
        </button>
        <button
          type="button"
          class={`side-toggle__btn ${side() === 'SHORT' ? 'side-toggle__btn--short active' : ''}`}
          onClick={() => setSide('SHORT')}
        >
          Short
        </button>
      </div>

      <div class="type-toggle">
        <button
          type="button"
          class={orderType() === 'MARKET' ? 'active' : ''}
          onClick={() => setOrderType('MARKET')}
        >
          Market
        </button>
        <button
          type="button"
          class={orderType() === 'LIMIT' ? 'active' : ''}
          onClick={() => setOrderType('LIMIT')}
        >
          Limit
        </button>
      </div>

      <Show when={orderType() === 'LIMIT'}>
        <label class="field">
          <span>Limit Price (USDT)</span>
          <input
            type="number"
            min="0"
            step="0.01"
            placeholder={props.markPrice() > 0 ? String(props.markPrice()) : '0'}
            value={limitPrice()}
            onInput={(e) => setLimitPrice(e.currentTarget.value)}
          />
        </label>
      </Show>

      <label class="field">
        <span>Position Size (USDT)</span>
        <input
          type="range"
          min="10"
          max="5000"
          step="10"
          value={size()}
          onInput={(e) => setSize(Number(e.currentTarget.value))}
        />
        <div class="field__value">{formatUsd(size())}</div>
      </label>

      <label class="field">
        <span>Leverage</span>
        <div class="leverage-row">
          <For each={[1, 2, 5, 10, 20, 50, 100]}>
            {(x) => (
              <button
                type="button"
                class={`leverage-chip ${leverage() === x ? 'active' : ''}`}
                onClick={() => setLeverage(x)}
              >
                {x}x
              </button>
            )}
          </For>
        </div>
      </label>

      <div class="trade-summary">
        <div><span>Margin</span><strong>{formatUsd(margin())}</strong></div>
        <div><span>Notional</span><strong>{formatUsd(notional())}</strong></div>
        <div><span>Entry</span><strong>{formatUsd(props.markPrice())}</strong></div>
      </div>

      <button
        type="button"
        class={`submit-btn submit-btn--${side().toLowerCase()}`}
        onClick={submit}
        disabled={props.markPrice() <= 0}
      >
        {side() === 'LONG' ? 'Open Long' : 'Open Short'}
      </button>

      <Show when={message()}>
        <p class="form-message">{message()}</p>
      </Show>
    </section>
  )
}
