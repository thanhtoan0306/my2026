import type { WsStatus } from '../types/trading'
import { tradingStore } from '../stores/tradingStore'
import { formatUsd } from '../utils/format'

interface HeaderProps {
  status: () => WsStatus
  symbol: () => string
}

export function Header(props: HeaderProps) {
  const statusLabel = () => {
    switch (props.status()) {
      case 'connected':
        return 'Live'
      case 'connecting':
        return 'Connecting…'
      default:
        return 'Disconnected'
    }
  }

  return (
    <header class="header">
      <div class="header__brand">
        <span class="header__logo">₿</span>
        <div>
          <h1>BTC Futures</h1>
          <p>Paper trading · Binance WebSocket</p>
        </div>
      </div>

      <div class="header__meta">
        <div class="pill">
          <span class={`status-dot status-dot--${props.status()}`} />
          {statusLabel()}
        </div>
        <div class="pill pill--muted">{props.symbol()}</div>
        <div class="pill pill--balance">
          Balance <strong>{formatUsd(tradingStore.state.balance)}</strong>
        </div>
      </div>
    </header>
  )
}
