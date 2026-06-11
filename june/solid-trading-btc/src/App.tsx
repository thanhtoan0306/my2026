import { Header } from './components/Header'
import { OrderBook } from './components/OrderBook'
import { PositionsPanel } from './components/PositionsPanel'
import { PriceChart } from './components/PriceChart'
import { PriceTicker } from './components/PriceTicker'
import { RecentTrades } from './components/RecentTrades'
import { TradeForm } from './components/TradeForm'
import { useBinanceFuturesWs } from './hooks/useBinanceFuturesWs'
import { useBinanceKlines } from './hooks/useBinanceKlines'
import './App.css'

function App() {
  const ws = useBinanceFuturesWs()
  const klines = useBinanceKlines('15m')
  const mark = () => ws.markPrice()?.markPrice ?? ws.ticker()?.lastPrice ?? 0

  return (
    <div class="app">
      <Header status={ws.status} symbol={() => 'BTCUSDT Perp'} />

      <PriceChart
        candles={klines.candles}
        loading={klines.loading}
        status={klines.status}
        interval={klines.interval}
        setInterval={klines.setInterval}
        intervals={klines.intervals}
      />

      <main class="layout">
        <div class="layout__left">
          <PriceTicker ticker={ws.ticker} markPrice={ws.markPrice} />
          <OrderBook orderBook={ws.orderBook} markPrice={mark} />
          <RecentTrades trades={ws.recentTrades} />
        </div>

        <div class="layout__right">
          <TradeForm markPrice={mark} />
          <PositionsPanel markPrice={mark} />
        </div>
      </main>
    </div>
  )
}

export default App
