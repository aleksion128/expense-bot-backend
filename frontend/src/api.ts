const API = '/api/v1'

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API}${path}`, {
    ...init,
    headers: {
      ...(init?.body instanceof FormData ? {} : { 'Content-Type': 'application/json' }),
      'X-User-Id': '1',
      ...(init?.headers || {}),
    },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'request failed')
  }
  return res.json()
}

export type Account = {
  id: number
  name: string
  type: 'card' | 'savings' | 'brokerage' | 'cash'
  balance: number
  currency: string
  bank_name?: string
}

export type Transaction = {
  id: number
  account_id: number
  transfer_account_id?: number
  type: 'income' | 'expense' | 'transfer'
  amount: number
  description: string
  category: string
  occurred_at: string
}

export type BudgetItem = {
  id: number
  kind: 'planned_income' | 'recurring_expense' | 'large_purchase'
  name: string
  amount: number
  category?: string
  day_of_month?: number
  due_date?: string
  is_active: boolean
}

export type Goal = {
  id: number
  name: string
  target_amount: number
  current_amount: number
  deadline?: string
}

export type Analytics = {
  total_balance: number
  income_period: number
  expense_period: number
  net_period: number
  by_category_expense: Record<string, number>
  by_category_income: Record<string, number>
  accounts: Account[]
}

export type Forecast = {
  current_balance: number
  points: { months_ahead: number; balance: number; income: number; expense: number }[]
}

export type ScenarioResult = {
  name: string
  amount: number
  baseline: Forecast
  with_purchase: Forecast
  delta_by_horizon: Record<string, number>
}

export type ImportBatch = {
  id: number
  filename: string
  bank: string
  status: string
  rows_total: number
  rows_imported: number
  rows_skipped: number
  created_at: string
}

export const api = {
  analytics: () => req<Analytics>('/analytics'),
  forecast: () => req<Forecast>('/forecast'),
  accounts: () => req<Account[]>('/accounts'),
  createAccount: (body: Partial<Account>) =>
    req<Account>('/accounts', { method: 'POST', body: JSON.stringify(body) }),
  updateAccount: (id: number, body: Partial<Account>) =>
    req<Account>(`/accounts/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  transactions: (limit = 200) => req<Transaction[]>(`/transactions?limit=${limit}`),
  createTransaction: (body: Partial<Transaction>) =>
    req<Transaction>('/transactions', { method: 'POST', body: JSON.stringify(body) }),
  budget: () => req<BudgetItem[]>('/budget'),
  createBudget: (body: Partial<BudgetItem>) =>
    req<BudgetItem>('/budget', { method: 'POST', body: JSON.stringify(body) }),
  deleteBudget: (id: number) => req<{ ok: boolean }>(`/budget/${id}`, { method: 'DELETE' }),
  goals: () => req<Goal[]>('/goals'),
  createGoal: (body: Partial<Goal>) =>
    req<Goal>('/goals', { method: 'POST', body: JSON.stringify(body) }),
  deleteGoal: (id: number) => req<{ ok: boolean }>(`/goals/${id}`, { method: 'DELETE' }),
  scenario: (body: { name: string; amount: number }) =>
    req<ScenarioResult>('/scenarios', { method: 'POST', body: JSON.stringify(body) }),
  imports: () => req<ImportBatch[]>('/imports'),
  importFile: async (file: File, accountId: number, bank: string, transferAccountId?: number) => {
    const fd = new FormData()
    fd.append('file', file)
    fd.append('account_id', String(accountId))
    fd.append('bank', bank)
    fd.append('mark_transfers', 'true')
    if (transferAccountId) fd.append('transfer_account_id', String(transferAccountId))
    return req<ImportBatch>('/import', { method: 'POST', body: fd })
  },
}

export function rub(kopecks: number, fractionDigits = 2): string {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(kopecks / 100)
}

export function rubShort(kopecks: number): string {
  return rub(kopecks, Math.abs(kopecks) % 100 === 0 ? 0 : 2)
}
