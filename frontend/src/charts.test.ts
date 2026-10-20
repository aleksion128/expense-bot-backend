import assert from 'node:assert/strict'
import {
  buildSeries,
  categoryTotals,
  latestAnchor,
  periodLabel,
  periodSums,
  shiftAnchor,
  parseISODate,
  type TxLike,
} from './charts.ts'

function at(y: number, m: number, d: number, hour = 12): string {
  return new Date(y, m - 1, d, hour).toISOString()
}

const txs: TxLike[] = [
  { type: 'expense', amount: 1000, occurred_at: at(2026, 3, 2), category: 'Еда' },
  { type: 'expense', amount: 500, occurred_at: at(2026, 3, 8), category: 'Еда' },
  { type: 'income', amount: 4000, occurred_at: at(2026, 3, 3), category: 'Зарплата' },
  { type: 'transfer', amount: 9000, occurred_at: at(2026, 3, 4), category: 'Себе' },
  { type: 'expense', amount: 700, occurred_at: at(2026, 3, 20), category: 'Дом' },
  { type: 'expense', amount: 300, occurred_at: at(2025, 4, 1), category: 'Еда' },
  { type: 'income', amount: 200, occurred_at: at(2026, 2, 28), category: 'Кэшбэк' },
]

const march = parseISODate('2026-03-04')

const week = buildSeries(txs, 'week', march)
assert.deepEqual(
  week.map((p) => p.label),
  ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'],
)
assert.equal(week[0].expense, 1000)
assert.equal(week[1].income, 4000)
assert.equal(week[6].expense, 500)
assert.equal(week.reduce((s, p) => s + p.expense + p.income, 0), 5500)

const month = buildSeries(txs, 'month', march)
assert.equal(month.length, 31)
assert.equal(month[1].expense, 1000)
assert.equal(month[19].expense, 700)
assert.equal(month[28].income, 0)

const feb = buildSeries(txs, 'month', parseISODate('2024-02-10'))
assert.equal(feb.length, 29)

const weeks = buildSeries(txs, 'weeks', march)
assert.equal(weeks.length, 12)
assert.equal(weeks[11].expense, 1500)
assert.equal(weeks[11].income, 4000)
assert.equal(weeks[0].expense, 0)

const months = buildSeries(txs, 'months', march)
assert.equal(months.length, 12)
assert.equal(months[0].expense, 300)
assert.equal(months[10].income, 200)
assert.equal(months[11].expense, 2200)

const sums = periodSums(txs, 'month', march)
assert.deepEqual(sums, { income: 4000, expense: 2200, net: 1800 })

const cats = categoryTotals(txs, 'month', march)
assert.deepEqual(cats, [
  ['Еда', 1500],
  ['Дом', 700],
])

assert.equal(latestAnchor(txs), '2026-03-20')
assert.equal(periodLabel('month', march).toLowerCase(), 'март 2026 г.')

const shifted = shiftAnchor('month', parseISODate('2026-01-31'), 1)
assert.equal(shifted.getMonth(), 1)
assert.equal(shifted.getFullYear(), 2026)

console.log('charts ok')
