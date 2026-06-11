export type KlineInterval = '1m' | '5m' | '15m' | '1h' | '4h' | '1d'

export interface Candle {
  time: number
  open: number
  high: number
  low: number
  close: number
  volume: number
}

export type Side = 'LONG' | 'SHORT'
export type OrderType = 'MARKET' | 'LIMIT'
export type WsStatus = 'connecting' | 'connected' | 'disconnected'

export interface TickerData {
  symbol: string
  lastPrice: number
  priceChange: number
  priceChangePercent: number
  high: number
  low: number
  volume: number
  quoteVolume: number
}

export interface MarkPriceData {
  symbol: string
  markPrice: number
  indexPrice: number
  fundingRate: number
  nextFundingTime: number
}

export interface OrderBookLevel {
  price: number
  quantity: number
}

export interface OrderBookData {
  bids: OrderBookLevel[]
  asks: OrderBookLevel[]
}

export interface AggTrade {
  id: number
  price: number
  quantity: number
  time: number
  isBuyerMaker: boolean
}

export interface Position {
  id: string
  side: Side
  entryPrice: number
  size: number
  leverage: number
  openedAt: number
}

export interface ClosedTrade {
  id: string
  side: Side
  entryPrice: number
  exitPrice: number
  size: number
  leverage: number
  pnl: number
  closedAt: number
}

export interface MarketSnapshot {
  status: WsStatus
  ticker: TickerData | null
  markPrice: MarkPriceData | null
  orderBook: OrderBookData
  recentTrades: AggTrade[]
}
