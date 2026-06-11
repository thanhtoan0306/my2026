import { writable } from 'svelte/store'
import type { ClosedTrade, Position, Side } from '../types/trading'

const INITIAL_BALANCE = 10_000

function calcPnl(side: Side, entry: number, exit: number, size: number, leverage: number) {
  const direction = side === 'LONG' ? 1 : -1
  return ((exit - entry) / entry) * size * leverage * direction
}

export const balance = writable(INITIAL_BALANCE)
export const positions = writable<Position[]>([])
export const history = writable<ClosedTrade[]>([])

export function openPosition(params: {
  side: Side
  entryPrice: number
  size: number
  leverage: number
}) {
  const margin = params.size / params.leverage
  let ok = false

  balance.update((b) => {
    if (margin > b) return b
    ok = true
    return b - margin
  })

  if (!ok) return false

  positions.update((list) => [
    ...list,
    {
      id: crypto.randomUUID(),
      side: params.side,
      entryPrice: params.entryPrice,
      size: params.size,
      leverage: params.leverage,
      openedAt: Date.now(),
    },
  ])

  return true
}

export function closePosition(id: string, exitPrice: number) {
  let current: Position | undefined
  positions.update((list) => {
    current = list.find((p) => p.id === id)
    return list.filter((p) => p.id !== id)
  })
  if (!current) return false

  const pnl = calcPnl(current.side, current.entryPrice, exitPrice, current.size, current.leverage)
  const margin = current.size / current.leverage

  history.update((list) => [
    {
      id: current!.id,
      side: current!.side,
      entryPrice: current!.entryPrice,
      exitPrice,
      size: current!.size,
      leverage: current!.leverage,
      pnl,
      closedAt: Date.now(),
    },
    ...list,
  ].slice(0, 50))

  balance.update((b) => b + margin + pnl)
  return true
}

export function unrealizedPnl(position: Position, markPrice: number) {
  return calcPnl(position.side, position.entryPrice, markPrice, position.size, position.leverage)
}

export function resetAccount() {
  balance.set(INITIAL_BALANCE)
  positions.set([])
  history.set([])
}
