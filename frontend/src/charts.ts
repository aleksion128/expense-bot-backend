export type ChartMode = 'week' | 'month' | 'weeks' | 'months'

export type ChartPrefs = {
  mode: ChartMode
  anchor: string
}

export type TxLike = {
  type: 'income' | 'expense' | 'transfer'
  amount: number
  occurred_at: string
  category?: string
}

export type ChartPoint = {
  label: string
  expense: number
  income: number
}

const STORAGE_KEY = 'budgetly.chart.v1'
const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

export const CHART_MODES: { id: ChartMode; label: string }[] = [
  { id: 'week', label: 'Неделя' },
  { id: 'month', label: 'Месяц' },
  { id: 'weeks', label: 'По неделям' },
  { id: 'months', label: 'По месяцам' },
]

function isMode(value: unknown): value is ChartMode {
  return value === 'week' || value === 'month' || value === 'weeks' || value === 'months'
}

function isISODate(value: unknown): value is string {
  return typeof value === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(value)
}

export function isoDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function parseISODate(value: string): Date {
  const [y, m, d] = value.split('-').map(Number)
  return new Date(y, (m || 1) - 1, d || 1)
}

export function loadChartPrefs(): { prefs: ChartPrefs; existed: boolean } {
  const fallback: ChartPrefs = { mode: 'month', anchor: isoDate(new Date()) }
  try {
    if (typeof localStorage === 'undefined') return { prefs: fallback, existed: false }
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { prefs: fallback, existed: false }
    const parsed = JSON.parse(raw) as Partial<ChartPrefs>
    if (!isMode(parsed.mode) || !isISODate(parsed.anchor)) return { prefs: fallback, existed: false }
    return { prefs: { mode: parsed.mode, anchor: parsed.anchor }, existed: true }
  } catch {
    return { prefs: fallback, existed: false }
  }
}

export function saveChartPrefs(prefs: ChartPrefs) {
  try {
    if (typeof localStorage === 'undefined') return
    localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs))
  } catch {
    // Private mode and quota errors should not break the dashboard.
  }
}

export function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

export function startOfWeek(d: Date): Date {
  const x = startOfDay(d)
  const mondayOffset = (x.getDay() + 6) % 7
  x.setDate(x.getDate() - mondayOffset)
  return x
}

export function startOfMonth(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), 1)
}

export function addDays(d: Date, days: number): Date {
  const x = new Date(d)
  x.setDate(x.getDate() + days)
  return x
}

export function addMonths(d: Date, months: number): Date {
  return new Date(d.getFullYear(), d.getMonth() + months, 1)
}

function daysBetween(from: Date, to: Date): number {
  const a = Date.UTC(from.getFullYear(), from.getMonth(), from.getDate())
  const b = Date.UTC(to.getFullYear(), to.getMonth(), to.getDate())
  return Math.round((b - a) / 86400000)
}

function monthShort(d: Date): string {
  return d.toLocaleDateString('ru-RU', { month: 'short' }).replace('.', '')
}

function dayMonth(d: Date): string {
  return `${d.getDate()} ${monthShort(d)}`
}

export function chartTitle(mode: ChartMode): string {
  if (mode === 'weeks') return 'Расходы и доходы по неделям'
  if (mode === 'months') return 'Расходы и доходы по месяцам'
  return 'Расходы и доходы по дням'
}

export function periodBounds(mode: ChartMode, anchor: Date): { from: Date; to: Date } {
  if (mode === 'week') {
    const from = startOfWeek(anchor)
    return { from, to: addDays(from, 7) }
  }
  if (mode === 'month') {
    const from = startOfMonth(anchor)
    return { from, to: addMonths(from, 1) }
  }
  if (mode === 'weeks') {
    const end = startOfWeek(anchor)
    return { from: addDays(end, -7 * 11), to: addDays(end, 7) }
  }
  const end = startOfMonth(anchor)
  return { from: addMonths(end, -11), to: addMonths(end, 1) }
}

export function periodLabel(mode: ChartMode, anchor: Date): string {
  const { from, to } = periodBounds(mode, anchor)
  const last = addDays(to, -1)
  if (mode === 'month') {
    const name = from.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })
    return name.charAt(0).toUpperCase() + name.slice(1)
  }
  if (mode === 'months') {
    return `${monthShort(from)} ${from.getFullYear()} – ${monthShort(last)} ${last.getFullYear()}`
  }
  const sameYear = from.getFullYear() === last.getFullYear()
  const left = sameYear ? dayMonth(from) : `${dayMonth(from)} ${from.getFullYear()}`
  return `${left} – ${dayMonth(last)} ${last.getFullYear()}`
}

export function shiftAnchor(mode: ChartMode, anchor: Date, dir: -1 | 1): Date {
  if (mode === 'week') return addDays(startOfDay(anchor), dir * 7)
  if (mode === 'month') return addMonths(startOfMonth(anchor), dir)
  if (mode === 'weeks') return addDays(startOfWeek(anchor), dir * 28)
  return addMonths(startOfMonth(anchor), dir * 3)
}

function emptyPoints(mode: ChartMode, anchor: Date): ChartPoint[] {
  if (mode === 'week') {
    return WEEKDAYS.map((label) => ({ label, expense: 0, income: 0 }))
  }
  if (mode === 'month') {
    const from = startOfMonth(anchor)
    const days = daysBetween(from, addMonths(from, 1))
    return Array.from({ length: days }, (_, i) => ({
      label: String(i + 1),
      expense: 0,
      income: 0,
    }))
  }
  if (mode === 'weeks') {
    const end = startOfWeek(anchor)
    const start = addDays(end, -7 * 11)
    return Array.from({ length: 12 }, (_, i) => ({
      label: dayMonth(addDays(start, i * 7)),
      expense: 0,
      income: 0,
    }))
  }
  const end = startOfMonth(anchor)
  const start = addMonths(end, -11)
  const crossesYear = start.getFullYear() !== end.getFullYear()
  return Array.from({ length: 12 }, (_, i) => {
    const month = addMonths(start, i)
    const label = crossesYear ? `${monthShort(month)} ${String(month.getFullYear()).slice(2)}` : monthShort(month)
    return { label, expense: 0, income: 0 }
  })
}

function pointIndex(mode: ChartMode, anchor: Date, occurred: Date): number | null {
  if (Number.isNaN(occurred.getTime())) return null
  if (mode === 'week') {
    const diff = daysBetween(startOfWeek(anchor), occurred)
    return diff >= 0 && diff < 7 ? diff : null
  }
  if (mode === 'month') {
    const start = startOfMonth(anchor)
    if (occurred.getFullYear() !== start.getFullYear() || occurred.getMonth() !== start.getMonth()) return null
    return occurred.getDate() - 1
  }
  if (mode === 'weeks') {
    const start = addDays(startOfWeek(anchor), -7 * 11)
    const diff = daysBetween(start, occurred)
    if (diff < 0 || diff >= 84) return null
    return Math.floor(diff / 7)
  }
  const start = addMonths(startOfMonth(anchor), -11)
  const idx = (occurred.getFullYear() - start.getFullYear()) * 12 + (occurred.getMonth() - start.getMonth())
  return idx >= 0 && idx < 12 ? idx : null
}

export function buildSeries(txs: TxLike[], mode: ChartMode, anchor: Date): ChartPoint[] {
  const points = emptyPoints(mode, anchor)
  for (const tx of txs) {
    if (tx.type === 'transfer') continue
    const idx = pointIndex(mode, anchor, new Date(tx.occurred_at))
    if (idx == null || !points[idx]) continue
    if (tx.type === 'expense') points[idx].expense += tx.amount
    if (tx.type === 'income') points[idx].income += tx.amount
  }
  return points
}

export function periodSums(txs: TxLike[], mode: ChartMode, anchor: Date) {
  const { from, to } = periodBounds(mode, anchor)
  let income = 0
  let expense = 0
  for (const tx of txs) {
    const occurred = new Date(tx.occurred_at)
    if (occurred < from || occurred >= to) continue
    if (tx.type === 'income') income += tx.amount
    if (tx.type === 'expense') expense += tx.amount
  }
  return { income, expense, net: income - expense }
}

export function categoryTotals(txs: TxLike[], mode: ChartMode, anchor: Date): [string, number][] {
  const { from, to } = periodBounds(mode, anchor)
  const totals = new Map<string, number>()
  for (const tx of txs) {
    if (tx.type !== 'expense') continue
    const occurred = new Date(tx.occurred_at)
    if (occurred < from || occurred >= to) continue
    const name = tx.category?.trim() || 'Без категории'
    totals.set(name, (totals.get(name) || 0) + tx.amount)
  }
  return [...totals.entries()].sort((a, b) => b[1] - a[1])
}

export function latestAnchor(txs: TxLike[]): string | null {
  let best: Date | null = null
  for (const tx of txs) {
    if (tx.type === 'transfer') continue
    const occurred = new Date(tx.occurred_at)
    if (Number.isNaN(occurred.getTime())) continue
    if (!best || occurred > best) best = occurred
  }
  return best ? isoDate(best) : null
}

export function formatAxisMoney(kopecks: number): string {
  const rubles = kopecks / 100
  const abs = Math.abs(rubles)
  if (abs >= 1_000_000) return `${trimNum(rubles / 1_000_000)} млн`
  if (abs >= 1_000) return `${trimNum(rubles / 1_000)} тыс`
  return String(Math.round(rubles))
}

function trimNum(value: number): string {
  const rounded = Math.round(value * 10) / 10
  return Number.isInteger(rounded) ? String(rounded) : rounded.toFixed(1).replace('.', ',')
}
