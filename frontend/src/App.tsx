import { useCallback, useEffect, useMemo, useState, type FormEvent } from 'react'
import {
  api,
  rub,
  type Account,
  type Analytics,
  type BudgetItem,
  type Forecast,
  type Goal,
  type ImportBatch,
  type ScenarioResult,
  type Transaction,
} from './api'
import './App.css'

type Tab = 'dashboard' | 'accounts' | 'import' | 'budget' | 'goals' | 'scenarios'

const accountTypeLabel: Record<Account['type'], string> = {
  card: 'Карта',
  savings: 'Накопительный',
  brokerage: 'Брокерский',
  cash: 'Наличные',
}

const budgetKindLabel: Record<BudgetItem['kind'], string> = {
  planned_income: 'Плановый доход',
  recurring_expense: 'Регулярный расход',
  large_purchase: 'Крупная покупка',
}

export default function App() {
  const [tab, setTab] = useState<Tab>('dashboard')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const [analytics, setAnalytics] = useState<Analytics | null>(null)
  const [forecast, setForecast] = useState<Forecast | null>(null)
  const [accounts, setAccounts] = useState<Account[]>([])
  const [txs, setTxs] = useState<Transaction[]>([])
  const [budget, setBudget] = useState<BudgetItem[]>([])
  const [goals, setGoals] = useState<Goal[]>([])
  const [imports, setImports] = useState<ImportBatch[]>([])

  const reload = useCallback(async () => {
    setError(null)
    setLoading(true)
    try {
      const [a, f, acc, t, b, g, i] = await Promise.all([
        api.analytics(),
        api.forecast(),
        api.accounts(),
        api.transactions(),
        api.budget(),
        api.goals(),
        api.imports(),
      ])
      setAnalytics(a)
      setForecast(f)
      setAccounts(acc)
      setTxs(t)
      setBudget(b)
      setGoals(g)
      setImports(i)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка загрузки')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void reload()
  }, [reload])

  const expenseCategories = useMemo(() => {
    if (!analytics) return []
    return Object.entries(analytics.by_category_expense).sort((a, b) => b[1] - a[1])
  }, [analytics])

  return (
    <div className="app">
      <header className="header">
        <div>
          <h1>Budgetly</h1>
          <p className="muted">Личные финансы · импорт выписок · прогноз</p>
        </div>
        <button className="btn ghost" onClick={() => void reload()} disabled={loading}>
          Обновить
        </button>
      </header>

      <nav className="tabs">
        {(
          [
            ['dashboard', 'Дашборд'],
            ['accounts', 'Счета'],
            ['import', 'Импорт'],
            ['budget', 'Бюджет'],
            ['goals', 'Цели'],
            ['scenarios', 'Сценарии'],
          ] as const
        ).map(([id, label]) => (
          <button
            key={id}
            className={tab === id ? 'tab active' : 'tab'}
            onClick={() => setTab(id)}
          >
            {label}
          </button>
        ))}
      </nav>

      {error && <div className="banner error">{error}</div>}
      {loading && !analytics && <div className="banner">Загрузка…</div>}

      {tab === 'dashboard' && analytics && (
        <Dashboard
          analytics={analytics}
          forecast={forecast}
          txs={txs}
          expenseCategories={expenseCategories}
        />
      )}
      {tab === 'accounts' && (
        <AccountsPanel accounts={accounts} onChanged={reload} />
      )}
      {tab === 'import' && (
        <ImportPanel accounts={accounts} imports={imports} onChanged={reload} />
      )}
      {tab === 'budget' && <BudgetPanel items={budget} onChanged={reload} />}
      {tab === 'goals' && <GoalsPanel goals={goals} onChanged={reload} />}
      {tab === 'scenarios' && <ScenariosPanel />}
    </div>
  )
}

function Dashboard({
  analytics,
  forecast,
  txs,
  expenseCategories,
}: {
  analytics: Analytics
  forecast: Forecast | null
  txs: Transaction[]
  expenseCategories: [string, number][]
}) {
  return (
    <section className="grid">
      <div className="card highlight">
        <div className="label">Капитал</div>
        <div className="value">{rub(analytics.total_balance)}</div>
      </div>
      <div className="card">
        <div className="label">Доходы за месяц</div>
        <div className="value good">{rub(analytics.income_period)}</div>
      </div>
      <div className="card">
        <div className="label">Расходы за месяц</div>
        <div className="value bad">{rub(analytics.expense_period)}</div>
      </div>
      <div className="card">
        <div className="label">Чистый поток</div>
        <div className={`value ${analytics.net_period >= 0 ? 'good' : 'bad'}`}>
          {rub(analytics.net_period)}
        </div>
      </div>

      <div className="card wide">
        <h3>Прогноз остатка</h3>
        <div className="forecast-row">
          {forecast?.points.map((p) => (
            <div key={p.months_ahead} className="forecast-item">
              <div className="label">через {p.months_ahead} мес.</div>
              <div className="value">{rub(p.balance)}</div>
              <div className="muted small">
                +{rub(p.income)} / −{rub(p.expense)}
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="card">
        <h3>Счета</h3>
        <ul className="list">
          {analytics.accounts.map((a) => (
            <li key={a.id}>
              <span>
                {a.name}{' '}
                <span className="muted">({accountTypeLabel[a.type]})</span>
              </span>
              <strong>{rub(a.balance)}</strong>
            </li>
          ))}
        </ul>
      </div>

      <div className="card">
        <h3>Расходы по категориям</h3>
        {expenseCategories.length === 0 && <p className="muted">Пока нет расходов</p>}
        <ul className="list">
          {expenseCategories.map(([cat, sum]) => (
            <li key={cat}>
              <span>{cat}</span>
              <strong>{rub(sum)}</strong>
            </li>
          ))}
        </ul>
      </div>

      <div className="card wide">
        <h3>Последние операции</h3>
        <table>
          <thead>
            <tr>
              <th>Дата</th>
              <th>Тип</th>
              <th>Описание</th>
              <th>Категория</th>
              <th>Сумма</th>
            </tr>
          </thead>
          <tbody>
            {txs.map((t) => (
              <tr key={t.id}>
                <td>{new Date(t.occurred_at).toLocaleDateString('ru-RU')}</td>
                <td>
                  <span className={`pill ${t.type}`}>
                    {t.type === 'income' ? 'доход' : t.type === 'expense' ? 'расход' : 'перевод'}
                  </span>
                </td>
                <td>{t.description || '—'}</td>
                <td>{t.category}</td>
                <td className={t.type === 'income' ? 'good' : t.type === 'expense' ? 'bad' : ''}>
                  {t.type === 'expense' ? '−' : t.type === 'income' ? '+' : ''}
                  {rub(t.amount)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}

function AccountsPanel({
  accounts,
  onChanged,
}: {
  accounts: Account[]
  onChanged: () => Promise<void>
}) {
  const [name, setName] = useState('')
  const [type, setType] = useState<Account['type']>('card')
  const [balance, setBalance] = useState('0')
  const [bank, setBank] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  async function create(e: FormEvent) {
    e.preventDefault()
    setBusy(true)
    setErr(null)
    try {
      await api.createAccount({
        name,
        type,
        balance: Math.round(parseFloat(balance.replace(',', '.')) * 100) || 0,
        bank_name: bank || undefined,
        currency: 'RUB',
      })
      setName('')
      setBalance('0')
      setBank('')
      await onChanged()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'ошибка')
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className="grid">
      <div className="card wide">
        <h3>Капитал по счетам</h3>
        <ul className="list big">
          {accounts.map((a) => (
            <li key={a.id}>
              <div>
                <strong>{a.name}</strong>
                <div className="muted">
                  {accountTypeLabel[a.type]}
                  {a.bank_name ? ` · ${a.bank_name}` : ''}
                </div>
              </div>
              <div className="value">{rub(a.balance)}</div>
            </li>
          ))}
        </ul>
      </div>
      <div className="card">
        <h3>Добавить счёт</h3>
        <form className="form" onSubmit={create}>
          <label>
            Название
            <input value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <label>
            Тип
            <select value={type} onChange={(e) => setType(e.target.value as Account['type'])}>
              <option value="card">Карта</option>
              <option value="savings">Накопительный</option>
              <option value="brokerage">Брокерский</option>
              <option value="cash">Наличные</option>
            </select>
          </label>
          <label>
            Баланс, ₽
            <input value={balance} onChange={(e) => setBalance(e.target.value)} />
          </label>
          <label>
            Банк
            <input value={bank} onChange={(e) => setBank(e.target.value)} />
          </label>
          {err && <div className="banner error">{err}</div>}
          <button className="btn" disabled={busy}>
            Создать
          </button>
        </form>
      </div>
    </section>
  )
}

function ImportPanel({
  accounts,
  imports,
  onChanged,
}: {
  accounts: Account[]
  imports: ImportBatch[]
  onChanged: () => Promise<void>
}) {
  const [accountId, setAccountId] = useState(accounts[0]?.id ?? 0)
  const [transferId, setTransferId] = useState(accounts[1]?.id ?? 0)
  const [bank, setBank] = useState('auto')
  const [file, setFile] = useState<File | null>(null)
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    if (!accountId && accounts[0]) setAccountId(accounts[0].id)
    if (!transferId && accounts[1]) setTransferId(accounts[1].id)
  }, [accounts, accountId, transferId])

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!file || !accountId) return
    setBusy(true)
    setErr(null)
    setMsg(null)
    try {
      const batch = await api.importFile(file, accountId, bank, transferId || undefined)
      setMsg(
        `Импортировано ${batch.rows_imported} из ${batch.rows_total} (пропущено ${batch.rows_skipped}), банк: ${batch.bank}`,
      )
      setFile(null)
      await onChanged()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'ошибка импорта')
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className="grid">
      <div className="card">
        <h3>Загрузить выписку</h3>
        <p className="muted">
          CSV/XLSX из банков РФ. Тинькофф определяется автоматически. Переводы между своими
          счетами не считаются расходами.
        </p>
        <form className="form" onSubmit={submit}>
          <label>
            Счёт списания/зачисления
            <select
              value={accountId}
              onChange={(e) => setAccountId(Number(e.target.value))}
              required
            >
              {accounts.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Счёт для переводов себе
            <select value={transferId} onChange={(e) => setTransferId(Number(e.target.value))}>
              <option value={0}>—</option>
              {accounts.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Банк
            <select value={bank} onChange={(e) => setBank(e.target.value)}>
              <option value="auto">Авто</option>
              <option value="tinkoff">Тинькофф / T-Bank</option>
              <option value="sber">Сбер</option>
              <option value="generic">Универсальный CSV</option>
            </select>
          </label>
          <label>
            Файл
            <input
              type="file"
              accept=".csv,.xlsx,.xls,text/csv"
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
              required
            />
          </label>
          {msg && <div className="banner ok">{msg}</div>}
          {err && <div className="banner error">{err}</div>}
          <button className="btn" disabled={busy || !file}>
            {busy ? 'Импорт…' : 'Импортировать'}
          </button>
        </form>
      </div>
      <div className="card">
        <h3>История загрузок</h3>
        {imports.length === 0 && <p className="muted">Пока пусто</p>}
        <ul className="list">
          {imports.map((i) => (
            <li key={i.id}>
              <div>
                <strong>{i.filename}</strong>
                <div className="muted small">
                  {new Date(i.created_at).toLocaleString('ru-RU')} · {i.bank}
                </div>
              </div>
              <span>
                {i.rows_imported}/{i.rows_total}
              </span>
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}

function BudgetPanel({
  items,
  onChanged,
}: {
  items: BudgetItem[]
  onChanged: () => Promise<void>
}) {
  const [kind, setKind] = useState<BudgetItem['kind']>('planned_income')
  const [name, setName] = useState('')
  const [amount, setAmount] = useState('')
  const [day, setDay] = useState('1')
  const [due, setDue] = useState('')
  const [err, setErr] = useState<string | null>(null)

  async function create(e: FormEvent) {
    e.preventDefault()
    setErr(null)
    try {
      const body: Partial<BudgetItem> = {
        kind,
        name,
        amount: Math.round(parseFloat(amount.replace(',', '.')) * 100),
        is_active: true,
      }
      if (kind !== 'large_purchase') body.day_of_month = Number(day) || 1
      if (kind === 'large_purchase' && due) body.due_date = due
      await api.createBudget(body)
      setName('')
      setAmount('')
      await onChanged()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'ошибка')
    }
  }

  return (
    <section className="grid">
      <div className="card wide">
        <h3>Бюджет</h3>
        <ul className="list">
          {items.map((b) => (
            <li key={b.id}>
              <div>
                <strong>{b.name}</strong>
                <div className="muted">
                  {budgetKindLabel[b.kind]}
                  {b.day_of_month ? ` · день ${b.day_of_month}` : ''}
                  {b.due_date ? ` · до ${b.due_date.slice(0, 10)}` : ''}
                </div>
              </div>
              <div className="row-actions">
                <strong>{rub(b.amount)}</strong>
                <button
                  className="btn ghost small"
                  onClick={() => void api.deleteBudget(b.id).then(onChanged)}
                >
                  ✕
                </button>
              </div>
            </li>
          ))}
        </ul>
      </div>
      <div className="card">
        <h3>Добавить</h3>
        <form className="form" onSubmit={create}>
          <label>
            Тип
            <select value={kind} onChange={(e) => setKind(e.target.value as BudgetItem['kind'])}>
              <option value="planned_income">Плановый доход</option>
              <option value="recurring_expense">Регулярный расход</option>
              <option value="large_purchase">Крупная покупка</option>
            </select>
          </label>
          <label>
            Название
            <input value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <label>
            Сумма, ₽
            <input value={amount} onChange={(e) => setAmount(e.target.value)} required />
          </label>
          {kind !== 'large_purchase' ? (
            <label>
              День месяца
              <input value={day} onChange={(e) => setDay(e.target.value)} />
            </label>
          ) : (
            <label>
              Дата
              <input type="date" value={due} onChange={(e) => setDue(e.target.value)} required />
            </label>
          )}
          {err && <div className="banner error">{err}</div>}
          <button className="btn">Сохранить</button>
        </form>
      </div>
    </section>
  )
}

function GoalsPanel({ goals, onChanged }: { goals: Goal[]; onChanged: () => Promise<void> }) {
  const [name, setName] = useState('')
  const [target, setTarget] = useState('')
  const [current, setCurrent] = useState('0')

  async function create(e: FormEvent) {
    e.preventDefault()
    await api.createGoal({
      name,
      target_amount: Math.round(parseFloat(target.replace(',', '.')) * 100),
      current_amount: Math.round(parseFloat(current.replace(',', '.')) * 100) || 0,
    })
    setName('')
    setTarget('')
    setCurrent('0')
    await onChanged()
  }

  return (
    <section className="grid">
      <div className="card wide">
        <h3>Финансовые цели</h3>
        {goals.length === 0 && <p className="muted">Целей пока нет</p>}
        <div className="goals">
          {goals.map((g) => {
            const pct = g.target_amount
              ? Math.min(100, Math.round((g.current_amount / g.target_amount) * 100))
              : 0
            return (
              <div key={g.id} className="goal">
                <div className="row-actions">
                  <strong>{g.name}</strong>
                  <button
                    className="btn ghost small"
                    onClick={() => void api.deleteGoal(g.id).then(onChanged)}
                  >
                    ✕
                  </button>
                </div>
                <div className="muted">
                  {rub(g.current_amount)} из {rub(g.target_amount)}
                </div>
                <div className="bar">
                  <div style={{ width: `${pct}%` }} />
                </div>
              </div>
            )
          })}
        </div>
      </div>
      <div className="card">
        <h3>Новая цель</h3>
        <form className="form" onSubmit={create}>
          <label>
            Название
            <input value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <label>
            Цель, ₽
            <input value={target} onChange={(e) => setTarget(e.target.value)} required />
          </label>
          <label>
            Уже накоплено, ₽
            <input value={current} onChange={(e) => setCurrent(e.target.value)} />
          </label>
          <button className="btn">Добавить</button>
        </form>
      </div>
    </section>
  )
}

function ScenariosPanel() {
  const [name, setName] = useState('Объектив')
  const [amount, setAmount] = useState('20000')
  const [result, setResult] = useState<ScenarioResult | null>(null)
  const [err, setErr] = useState<string | null>(null)

  async function run(e: FormEvent) {
    e.preventDefault()
    setErr(null)
    try {
      const res = await api.scenario({
        name,
        amount: Math.round(parseFloat(amount.replace(',', '.')) * 100),
      })
      setResult(res)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'ошибка')
    }
  }

  return (
    <section className="grid">
      <div className="card">
        <h3>А что, если…</h3>
        <p className="muted">Сравним прогноз с покупкой и без неё.</p>
        <form className="form" onSubmit={run}>
          <label>
            Что покупаем
            <input value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <label>
            Сумма, ₽
            <input value={amount} onChange={(e) => setAmount(e.target.value)} required />
          </label>
          {err && <div className="banner error">{err}</div>}
          <button className="btn">Посчитать</button>
        </form>
      </div>
      {result && (
        <div className="card wide">
          <h3>
            Сценарий: {result.name} (−{rub(result.amount)})
          </h3>
          <div className="forecast-row">
            {result.baseline.points.map((p, idx) => {
              const withP = result.with_purchase.points[idx]
              return (
                <div key={p.months_ahead} className="forecast-item">
                  <div className="label">через {p.months_ahead} мес.</div>
                  <div className="muted small">без покупки</div>
                  <div className="value">{rub(p.balance)}</div>
                  <div className="muted small">с покупкой</div>
                  <div className="value bad">{rub(withP.balance)}</div>
                  <div className="muted small">
                    дельта {rub(result.delta_by_horizon[`${p.months_ahead}m`] || 0)}
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </section>
  )
}
