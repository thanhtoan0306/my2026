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
