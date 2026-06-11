import { createStore } from 'solid-js/store'
import type { ClosedTrade, Position, Side } from '../types/trading'

const INITIAL_BALANCE = 10_000

function calcPnl(side: Side, entry: number, exit: number, size: number, leverage: number) {
  const direction = side === 'LONG' ? 1 : -1
  return ((exit - entry) / entry) * size * leverage * direction
}

function createTradingStore() {
  const [state, setState] = createStore({
    balance: INITIAL_BALANCE,
    positions: [] as Position[],
    history: [] as ClosedTrade[],
  })

  const openPosition = (params: {
    side: Side
    entryPrice: number
    size: number
    leverage: number
  }) => {
    const margin = params.size / params.leverage
    if (margin > state.balance) return false

    const position: Position = {
      id: crypto.randomUUID(),
      side: params.side,
      entryPrice: params.entryPrice,
      size: params.size,
      leverage: params.leverage,
      openedAt: Date.now(),
    }

    setState('balance', (b) => b - margin)
    setState('positions', (list) => [...list, position])
    return true
  }

  const closePosition = (id: string, exitPrice: number) => {
    const position = state.positions.find((p) => p.id === id)
    if (!position) return false

    const pnl = calcPnl(position.side, position.entryPrice, exitPrice, position.size, position.leverage)
    const margin = position.size / position.leverage

    const closed: ClosedTrade = {
      id: position.id,
      side: position.side,
      entryPrice: position.entryPrice,
      exitPrice,
      size: position.size,
      leverage: position.leverage,
      pnl,
      closedAt: Date.now(),
    }

    setState('positions', (list) => list.filter((p) => p.id !== id))
    setState('history', (list) => [closed, ...list].slice(0, 50))
    setState('balance', (b) => b + margin + pnl)
    return true
  }

  const unrealizedPnl = (position: Position, markPrice: number) =>
    calcPnl(position.side, position.entryPrice, markPrice, position.size, position.leverage)

  const resetAccount = () => {
    setState({
      balance: INITIAL_BALANCE,
      positions: [],
      history: [],
    })
  }

  return { state, openPosition, closePosition, unrealizedPnl, resetAccount }
}

export const tradingStore = createTradingStore()
