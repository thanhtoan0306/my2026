<script lang="ts">
  import { onMount } from 'svelte'
  import svelteLogo from './assets/svelte.svg'
  import Header from './lib/components/Header.svelte'
  import OrderBook from './lib/components/OrderBook.svelte'
  import PositionsPanel from './lib/components/PositionsPanel.svelte'
  import PriceChart from './lib/components/PriceChart.svelte'
  import PriceTicker from './lib/components/PriceTicker.svelte'
  import RecentTrades from './lib/components/RecentTrades.svelte'
  import TradeForm from './lib/components/TradeForm.svelte'
  import { startFuturesWs, markPrice, orderBook, recentTrades, ticker, wsStatus } from './lib/binance/futuresWs'
  import { candles, klinesLoading, klineStatus, startKlinesWs } from './lib/binance/klines'

  onMount(() => {
    const stopFutures = startFuturesWs()
    const stopKlines = startKlinesWs()
    return () => {
      stopFutures()
      stopKlines()
    }
  })

  const mark = $derived($markPrice?.markPrice ?? $ticker?.lastPrice ?? 0)
</script>

<div class="app svelte-theme">
  <Header status={$wsStatus} />

  <PriceChart data={$candles} loading={$klinesLoading} status={$klineStatus} />

  <main class="layout">
    <div class="layout__left">
      <PriceTicker ticker={$ticker} mark={$markPrice} />
      <OrderBook book={$orderBook} {mark} />
      <RecentTrades trades={$recentTrades} />
    </div>

    <div class="layout__right">
      <TradeForm {mark} />
      <PositionsPanel {mark} />
    </div>
  </main>

  <footer class="svelte-footer">
    <img src={svelteLogo} alt="" class="svelte-footer__logo" />
    <span>Built with <strong>Svelte</strong> · Không phải lệnh thật</span>
  </footer>
</div>

<style>
  :global(:root) {
    color: #f0ebff;
    background: #0a0614;
    font-family: 'Segoe UI', 'SF Pro Text', system-ui, sans-serif;
    line-height: 1.5;
    font-weight: 400;
    font-synthesis: none;
    text-rendering: optimizeLegibility;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;

    --bg: #0a0614;
    --panel: #15102a;
    --panel-border: #2d2150;
    --panel-glow: rgba(124, 58, 237, 0.12);
    --text: #f0ebff;
    --muted: #9d8ec7;
    --up: #34d399;
    --down: #fb7185;
    --accent: #ff3e00;
    --accent-violet: #a855f7;
    --long: #2dd4bf;
    --short: #f472b6;
  }

  :global(*) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    min-width: 320px;
    background:
      radial-gradient(ellipse at 0% 0%, rgba(124, 58, 237, 0.22), transparent 42%),
      radial-gradient(ellipse at 100% 0%, rgba(255, 62, 0, 0.14), transparent 38%),
      radial-gradient(ellipse at 50% 100%, rgba(168, 85, 247, 0.08), transparent 45%),
      var(--bg);
  }

  :global(button, input) {
    font: inherit;
  }

  :global(.text-up) {
    color: var(--up);
  }

  :global(.text-down) {
    color: var(--down);
  }

  :global(.app) {
    max-width: 1400px;
    margin: 0 auto;
    padding: 1rem 1.25rem 2rem;
  }

  :global(.svelte-theme .panel) {
    box-shadow:
      0 0 0 1px var(--panel-border),
      0 8px 32px rgba(0, 0, 0, 0.35),
      inset 0 1px 0 rgba(255, 255, 255, 0.03);
  }

  .svelte-footer {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    margin-top: 1.5rem;
    padding-top: 1rem;
    border-top: 1px solid var(--panel-border);
    color: var(--muted);
    font-size: 0.78rem;
  }

  .svelte-footer strong {
    color: var(--accent);
    font-weight: 700;
  }

  .svelte-footer__logo {
    width: 1rem;
    height: auto;
    opacity: 0.85;
  }

  :global(.header) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    flex-wrap: wrap;
    margin-bottom: 1.25rem;
  }

  :global(.header__brand) {
    display: flex;
    align-items: center;
    gap: 0.85rem;
  }

  :global(.header__brand h1) {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 700;
  }

  :global(.header__brand p) {
    margin: 0.15rem 0 0;
    color: var(--muted);
    font-size: 0.85rem;
  }

  :global(.header__logo) {
    width: 2.75rem;
    height: 2.75rem;
    display: grid;
    place-items: center;
    border-radius: 0.85rem;
    background: linear-gradient(135deg, #7c3aed, #ff3e00);
    color: #111;
    font-size: 1.35rem;
    font-weight: 800;
  }

  :global(.header__meta) {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  :global(.pill) {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.45rem 0.75rem;
    border-radius: 999px;
    background: var(--panel);
    border: 1px solid var(--panel-border);
    font-size: 0.85rem;
  }

  :global(.pill--muted) {
    color: var(--muted);
  }

  :global(.pill--balance strong) {
    color: var(--accent);
  }

  :global(.status-dot) {
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 50%;
    background: var(--muted);
  }

  :global(.status-dot--connected) {
    background: var(--up);
    box-shadow: 0 0 8px rgba(52, 211, 153, 0.75);
  }

  :global(.status-dot--connecting) {
    background: var(--accent);
    animation: pulse 1s infinite;
  }

  :global(.status-dot--disconnected) {
    background: var(--down);
  }

  @keyframes pulse {
    50% {
      opacity: 0.45;
    }
  }

  :global(.chart-panel) {
    margin-bottom: 0;
  }

  :global(.chart-panel__title) {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  :global(.interval-toggle) {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  :global(.interval-toggle button) {
    padding: 0.3rem 0.55rem;
    border-radius: 0.45rem;
    border: 1px solid var(--panel-border);
    background: rgba(255, 255, 255, 0.03);
    color: var(--muted);
    font-size: 0.75rem;
    cursor: pointer;
  }

  :global(.interval-toggle button.active) {
    border-color: var(--accent);
    color: #ffb899;
    background: rgba(255, 62, 0, 0.14);
    box-shadow: 0 0 10px rgba(255, 62, 0, 0.2);
  }

  :global(.chart-wrap) {
    position: relative;
    min-height: 420px;
  }

  :global(.chart-container) {
    width: 100%;
    height: 420px;
  }

  :global(.chart-loading) {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    background: rgba(18, 24, 33, 0.72);
    color: var(--muted);
    font-size: 0.85rem;
    border-radius: 0.75rem;
  }

  :global(.layout) {
    display: grid;
    grid-template-columns: minmax(0, 1.4fr) minmax(320px, 0.9fr);
    gap: 1rem;
    margin-top: 1rem;
  }

  :global(.layout__left),
  :global(.layout__right) {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  :global(.panel) {
    background: var(--panel);
    border: 1px solid var(--panel-border);
    border-radius: 1rem;
    padding: 1rem;
  }

  :global(.panel__head) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    margin-bottom: 0.85rem;
  }

  :global(.panel__head h2) {
    margin: 0;
    font-size: 0.95rem;
  }

  :global(.panel__hint) {
    color: var(--muted);
    font-size: 0.75rem;
  }

  :global(.price-panel__main) {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  :global(.label) {
    margin: 0 0 0.25rem;
    color: var(--muted);
    font-size: 0.8rem;
  }

  :global(.price) {
    margin: 0;
    font-size: 2rem;
    font-weight: 700;
    letter-spacing: -0.03em;
  }

  :global(.price--loading) {
    color: var(--muted);
  }

  :global(.price--up) {
    color: var(--up);
  }

  :global(.price--down) {
    color: var(--down);
  }

  :global(.change-badge) {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    padding: 0.5rem 0.75rem;
    border-radius: 0.75rem;
    font-weight: 600;
    font-size: 0.95rem;
  }

  :global(.change-badge span) {
    font-size: 0.8rem;
    font-weight: 500;
    opacity: 0.85;
  }

  :global(.change-badge--up) {
    background: rgba(52, 211, 153, 0.12);
    color: var(--up);
  }

  :global(.change-badge--down) {
    background: rgba(251, 113, 133, 0.12);
    color: var(--down);
  }

  :global(.stats-grid) {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.75rem;
  }

  :global(.stat) {
    padding: 0.65rem 0.75rem;
    border-radius: 0.75rem;
    background: rgba(255, 255, 255, 0.03);
  }

  :global(.stat span) {
    display: block;
    color: var(--muted);
    font-size: 0.72rem;
    margin-bottom: 0.2rem;
  }

  :global(.stat strong) {
    font-size: 0.85rem;
  }

  :global(.orderbook__header),
  :global(.orderbook__row),
  :global(.trades-table__header),
  :global(.trades-table__row) {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
    font-size: 0.78rem;
  }

  :global(.orderbook__header),
  :global(.trades-table__header) {
    color: var(--muted);
    margin-bottom: 0.35rem;
  }

  :global(.orderbook__row),
  :global(.trades-table__row) {
    position: relative;
    padding: 0.18rem 0;
  }

  :global(.orderbook__row span:last-child),
  :global(.trades-table__row span:last-child) {
    text-align: right;
  }

  :global(.orderbook__bar) {
    position: absolute;
    top: 0;
    bottom: 0;
    right: 0;
    opacity: 0.18;
    pointer-events: none;
  }

  :global(.orderbook__bar--ask) {
    background: var(--down);
  }

  :global(.orderbook__bar--bid) {
    background: var(--up);
  }

  :global(.orderbook__row--ask span:first-child) {
    color: var(--down);
  }

  :global(.orderbook__row--bid span:first-child) {
    color: var(--up);
  }

  :global(.orderbook__mid) {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    margin: 0.5rem 0;
    padding: 0.45rem;
    border-radius: 0.5rem;
    background: rgba(124, 58, 237, 0.12);
    border: 1px solid rgba(255, 62, 0, 0.2);
    color: #ffb899;
    font-size: 0.85rem;
  }

  :global(.trades-table__row--buy span:first-child) {
    color: var(--up);
  }

  :global(.trades-table__row--sell span:first-child) {
    color: var(--down);
  }

  :global(.side-toggle),
  :global(.type-toggle) {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
    margin-bottom: 0.85rem;
  }

  :global(.side-toggle__btn),
  :global(.type-toggle button),
  :global(.leverage-chip),
  :global(.submit-btn),
  :global(.close-btn),
  :global(.link-btn) {
    border: 1px solid var(--panel-border);
    background: rgba(255, 255, 255, 0.03);
    color: var(--text);
    border-radius: 0.65rem;
    cursor: pointer;
    transition: background 0.15s, border-color 0.15s;
  }

  :global(.side-toggle__btn),
  :global(.type-toggle button) {
    padding: 0.65rem;
    font-weight: 600;
  }

  :global(.side-toggle__btn--long.active) {
    background: rgba(45, 212, 191, 0.14);
    border-color: var(--long);
    color: var(--long);
  }

  :global(.side-toggle__btn--short.active) {
    background: rgba(244, 114, 182, 0.14);
    border-color: var(--short);
    color: var(--short);
  }

  :global(.type-toggle button.active) {
    border-color: var(--accent-violet);
    color: var(--accent-violet);
  }

  :global(.field) {
    display: block;
    margin-bottom: 0.85rem;
  }

  :global(.field span) {
    display: block;
    margin-bottom: 0.35rem;
    color: var(--muted);
    font-size: 0.8rem;
  }

  :global(.field input[type='number']) {
    width: 100%;
    padding: 0.65rem 0.75rem;
    border-radius: 0.65rem;
    border: 1px solid var(--panel-border);
    background: rgba(0, 0, 0, 0.25);
    color: var(--text);
  }

  :global(.field input[type='range']) {
    width: 100%;
  }

  :global(.field__value) {
    margin-top: 0.35rem;
    font-weight: 600;
  }

  :global(.leverage-row) {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  :global(.leverage-chip) {
    padding: 0.35rem 0.55rem;
    font-size: 0.78rem;
  }

  :global(.leverage-chip.active) {
    border-color: var(--accent-violet);
    color: var(--accent-violet);
    background: rgba(168, 85, 247, 0.1);
  }

  :global(.trade-summary) {
    display: grid;
    gap: 0.45rem;
    margin: 0.85rem 0;
    padding: 0.75rem;
    border-radius: 0.75rem;
    background: rgba(255, 255, 255, 0.03);
    font-size: 0.82rem;
  }

  :global(.trade-summary div) {
    display: flex;
    justify-content: space-between;
  }

  :global(.trade-summary span) {
    color: var(--muted);
  }

  :global(.submit-btn) {
    width: 100%;
    padding: 0.85rem;
    font-weight: 700;
    border: none;
  }

  :global(.submit-btn--long) {
    background: linear-gradient(135deg, #0d9488, var(--long));
    color: #042f2e;
    box-shadow: 0 4px 16px rgba(45, 212, 191, 0.25);
  }

  :global(.submit-btn--short) {
    background: linear-gradient(135deg, #db2777, var(--short));
    color: #1a0510;
    box-shadow: 0 4px 16px rgba(244, 114, 182, 0.25);
  }

  :global(.submit-btn:disabled) {
    opacity: 0.45;
    cursor: not-allowed;
  }

  :global(.form-message) {
    margin: 0.75rem 0 0;
    font-size: 0.82rem;
    color: var(--muted);
  }

  :global(.empty-state) {
    margin: 0;
    color: var(--muted);
    font-size: 0.85rem;
  }

  :global(.positions-list) {
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
  }

  :global(.position-card) {
    padding: 0.75rem;
    border-radius: 0.75rem;
    border: 1px solid var(--panel-border);
    background: rgba(0, 0, 0, 0.18);
  }

  :global(.position-card--long) {
    border-left: 3px solid var(--long);
  }

  :global(.position-card--short) {
    border-left: 3px solid var(--short);
  }

  :global(.position-card__head),
  :global(.position-card__foot) {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  :global(.position-card__grid) {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.45rem;
    margin: 0.65rem 0;
    font-size: 0.78rem;
  }

  :global(.position-card__grid span) {
    display: block;
    color: var(--muted);
    margin-bottom: 0.15rem;
  }

  :global(.close-btn),
  :global(.link-btn) {
    padding: 0.35rem 0.65rem;
    font-size: 0.78rem;
  }

  :global(.link-btn) {
    background: transparent;
    color: var(--muted);
  }

  :global(.history-block) {
    margin-top: 1rem;
    padding-top: 0.85rem;
    border-top: 1px solid var(--panel-border);
  }

  :global(.history-block h3) {
    margin: 0 0 0.5rem;
    font-size: 0.82rem;
    color: var(--muted);
  }

  :global(.history-row) {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 0.5rem;
    font-size: 0.75rem;
    padding: 0.25rem 0;
  }

  @media (max-width: 960px) {
    :global(.layout) {
      grid-template-columns: 1fr;
    }

    :global(.stats-grid) {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 520px) {
    :global(.price) {
      font-size: 1.55rem;
    }

    :global(.stats-grid) {
      grid-template-columns: 1fr;
    }
  }
</style>
