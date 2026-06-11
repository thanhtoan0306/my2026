<script lang="ts">
  import svelteLogo from '../../assets/svelte.svg'
  import { balance } from '../stores/tradingStore'
  import { formatUsd } from '../utils/format'
  import type { WsStatus } from '../types/trading'

  interface Props {
    status: WsStatus
    symbol?: string
  }

  let { status, symbol = 'BTCUSDT Perp' }: Props = $props()

  const statusLabel = $derived(
    status === 'connected' ? 'Live' : status === 'connecting' ? 'Connecting…' : 'Disconnected',
  )
</script>

<header class="header">
  <div class="header__brand">
    <div class="header__logo-wrap">
      <img src={svelteLogo} alt="Svelte" class="header__svelte-logo" />
    </div>
    <div>
      <div class="header__title-row">
        <h1>Svelte <span class="header__accent">BTC Futures</span></h1>
        <span class="svelte-badge">Svelte 5</span>
      </div>
      <p>Reactive paper trading · Binance WebSocket</p>
    </div>
  </div>

  <div class="header__meta">
    <div class="pill pill--svelte">⚡ Svelte Edition</div>
    <div class="pill">
      <span class="status-dot status-dot--{status}"></span>
      {statusLabel}
    </div>
    <div class="pill pill--muted">{symbol}</div>
    <div class="pill pill--balance">
      Balance <strong>{formatUsd($balance)}</strong>
    </div>
  </div>
</header>

<style>
  .header__logo-wrap {
    width: 3rem;
    height: 3rem;
    display: grid;
    place-items: center;
    border-radius: 1rem;
    background: linear-gradient(145deg, rgba(124, 58, 237, 0.35), rgba(255, 62, 0, 0.18));
    border: 1px solid rgba(255, 62, 0, 0.35);
    box-shadow:
      0 0 24px rgba(255, 62, 0, 0.15),
      inset 0 1px 0 rgba(255, 255, 255, 0.06);
  }

  .header__svelte-logo {
    width: 1.55rem;
    height: auto;
    filter: drop-shadow(0 0 6px rgba(255, 62, 0, 0.45));
  }

  .header__title-row {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    flex-wrap: wrap;
  }

  .header__accent {
    background: linear-gradient(90deg, #ff3e00, #ff8a65);
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
  }

  .svelte-badge {
    font-size: 0.62rem;
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 0.22rem 0.5rem;
    border-radius: 999px;
    background: linear-gradient(135deg, #ff3e00, #c026d3);
    color: #fff;
    box-shadow: 0 0 12px rgba(255, 62, 0, 0.35);
  }

  .pill--svelte {
    border-color: rgba(255, 62, 0, 0.45);
    color: #ffc9b5;
    background: rgba(255, 62, 0, 0.1);
    font-weight: 600;
    font-size: 0.78rem;
  }
</style>
