export function formatPrice(value: number, digits = 2) {
  return value.toLocaleString('en-US', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  })
}

export function formatUsd(value: number, digits = 2) {
  const prefix = value >= 0 ? '' : '-'
  return `${prefix}$${formatPrice(Math.abs(value), digits)}`
}

export function formatPercent(value: number, digits = 2) {
  const prefix = value >= 0 ? '+' : ''
  return `${prefix}${value.toFixed(digits)}%`
}

export function formatQty(value: number) {
  if (value >= 1) return value.toFixed(4)
  return value.toFixed(6)
}

export function formatFundingRate(rate: number) {
  return `${(rate * 100).toFixed(4)}%`
}

export function formatTime(ts: number) {
  return new Date(ts).toLocaleTimeString('en-US', { hour12: false })
}

export function shortTime(ts: number) {
  return new Date(ts).toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

export function countdownTo(ts: number) {
  const diff = Math.max(0, ts - Date.now())
  const h = Math.floor(diff / 3_600_000)
  const m = Math.floor((diff % 3_600_000) / 60_000)
  const s = Math.floor((diff % 60_000) / 1000)
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}
