import { useCallback, useEffect, useMemo, useRef, useState, type FormEvent, type ReactNode } from 'react'
import {
  Area,
  AreaChart,
  CartesianGrid,
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  api,
  rub,
  rubShort,
  type Account,
  type Analytics,
  type BudgetItem,
  type Forecast,
  type Goal,
  type ImportBatch,
  type ScenarioResult,
  type Transaction,
} from './api'
import {
  buildSeries,
  categoryTotals,
  CHART_MODES,
  chartTitle,
  formatAxisMoney,
  isoDate,
  latestAnchor,
  loadChartPrefs,
  parseISODate,
  periodBounds,
  periodLabel,
  periodSums,
  saveChartPrefs,
  shiftAnchor,
  type ChartMode,
  type ChartPrefs,
} from './charts'
import './App.css'

type Tab = 'dashboard' | 'accounts' | 'import' | 'budget' | 'goals' | 'scenarios'

const CATEGORY_COLORS = ['#7eb6ff', '#5b8cff', '#9f7aea', '#63e6be', '#74c0fc', '#a5b4fc', '#c4b5fd', '#94a3b8']

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

const nav: { id: Tab; label: string; icon: ReactNode }[] = [
  { id: 'dashboard', label: 'Дашборд', icon: <IconGrid /> },
  { id: 'accounts', label: 'Счета', icon: <IconCards /> },
  { id: 'import', label: 'Импорт', icon: <IconUpload /> },
  { id: 'budget', label: 'Бюджет', icon: <IconBills /> },
  { id: 'goals', label: 'Цели', icon: <IconTarget /> },
  { id: 'scenarios', label: 'Сценарии', icon: <IconSpark /> },
]

export default function App() {
  const [tab, setTab] = useState<Tab>('dashboard')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [manualOpen, setManualOpen] = useState(false)

  const [analytics, setAnalytics] = useState<Analytics | null>(null)
  const [forecast, setForecast] = useState<Forecast | null>(null)
  const [accounts, setAccounts] = useState<Account[]>([])
  const [txs, setTxs] = useState<Transaction[]>([])
  const [txTruncated, setTxTruncated] = useState(false)
  const [budget, setBudget] = useState<BudgetItem[]>([])
  const [goals, setGoals] = useState<Goal[]>([])
  const [imports, setImports] = useState<ImportBatch[]>([])
  const [chartPrefs, setChartPrefs] = useState<ChartPrefs>(() => loadChartPrefs().prefs)
  const [chartReady, setChartReady] = useState(() => loadChartPrefs().existed)
  const allowAutoAnchor = useRef(!loadChartPrefs().existed)
  const txsRef = useRef<Transaction[]>([])

  const reload = useCallback(async () => {
    setError(null)
    setLoading(true)
    try {
      const [a, f, acc, txPage, b, g, i] = await Promise.all([
        api.analytics(),
        api.forecast(),
        api.accounts(),
        api.transactionsAll(),
        api.budget(),
        api.goals(),
        api.imports(),
      ])
      setAnalytics(a)
      setForecast(f)
      setAccounts(acc)
      setTxs(txPage.items)
      txsRef.current = txPage.items
      setTxTruncated(txPage.truncated)
      setBudget(b)
      setGoals(g)
      setImports(i)
      if (allowAutoAnchor.current) {
        allowAutoAnchor.current = false
        const anchor = latestAnchor(txPage.items)
        if (anchor) setChartPrefs((prev) => ({ ...prev, anchor }))
      }
      setChartReady(true)
      return txPage.items
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка загрузки')
      return txsRef.current
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void reload()
  }, [reload])

  useEffect(() => {
    if (!chartReady) return
    saveChartPrefs(chartPrefs)
  }, [chartPrefs, chartReady])

  const updateChartPrefs = useCallback((next: ChartPrefs) => {
    allowAutoAnchor.current = false
    setChartReady(true)
    setChartPrefs(next)
  }, [])

  const reloadAndFocusNew = useCallback(async () => {
    const prevIds = new Set(txsRef.current.map((t) => t.id))
    const next = await reload()
    const fresh = next.filter((t) => !prevIds.has(t.id))
    const anchor = latestAnchor(fresh)
    if (!anchor) return
    allowAutoAnchor.current = false
    setChartReady(true)
    setChartPrefs((prev) => ({ ...prev, anchor }))
  }, [reload])

  const upcoming = useMemo(() => buildUpcomingBills(budget), [budget])
  const estimated = forecast?.points.find((p) => p.months_ahead === 1)?.balance ?? analytics?.total_balance ?? 0
  const chartAnchor = useMemo(() => parseISODate(chartPrefs.anchor), [chartPrefs.anchor])
  const selectedPeriod = useMemo(
    () => periodSums(txs, chartPrefs.mode, chartAnchor),
    [txs, chartPrefs.mode, chartAnchor],
  )
  const selectedPeriodLabel = useMemo(
    () => periodLabel(chartPrefs.mode, chartAnchor),
    [chartPrefs.mode, chartAnchor],
  )

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-mark">B</span>
          Budgetly
        </div>
        <nav className="nav">
          {nav.map((item) => (
            <button
              key={item.id}
              className={tab === item.id ? 'nav-item active' : 'nav-item'}
              onClick={() => setTab(item.id)}
            >
              {item.icon}
              {item.label}
            </button>
          ))}
        </nav>
        <div className="sidebar-footer">
          <button className="nav-item" onClick={() => void reload()} disabled={loading}>
            <IconRefresh />
            Обновить
          </button>
        </div>
      </aside>

      <main className="main">
        <header className="topbar">
          <div className="greeting">
            <div className="avatar">Д</div>
            <div>
              <h1>Привет! Отличный месяц для бюджета</h1>
              <p>
                {loading
                  ? 'Обновляем данные…'
                  : analytics
                    ? `Капитал ${rubShort(analytics.total_balance)} · ${selectedPeriodLabel}: ${rubShort(selectedPeriod.net)}`
                    : 'Загрузите выписку, чтобы увидеть аналитику'}
              </p>
            </div>
          </div>
          <div className="top-actions">
            <button className="btn" onClick={() => setManualOpen(true)}>
              Добавить вручную +
            </button>
            <button className="icon-btn" title="Уведомления" type="button">
              <IconBell />
            </button>
          </div>
        </header>

        {error && <div className="banner error">{error}</div>}

        {tab === 'dashboard' && analytics && (
          <DashboardView
            analytics={analytics}
            estimated={estimated}
            txs={txs}
            truncated={txTruncated}
            prefs={chartPrefs}
            onPrefs={updateChartPrefs}
            upcoming={upcoming}
            onSeeAll={() => setTab('budget')}
          />
        )}
        {tab === 'accounts' && <AccountsPanel accounts={accounts} onChanged={reload} />}
        {tab === 'import' && <ImportPanel accounts={accounts} imports={imports} onChanged={reloadAndFocusNew} />}
        {tab === 'budget' && <BudgetPanel items={budget} onChanged={reload} />}
        {tab === 'goals' && <GoalsPanel goals={goals} onChanged={reload} />}
        {tab === 'scenarios' && <ScenariosPanel />}
      </main>

      {manualOpen && (
        <ManualModal
          accounts={accounts}
          onClose={() => setManualOpen(false)}
          onSaved={async () => {
            setManualOpen(false)
            setTab('dashboard')
            await reloadAndFocusNew()
          }}
        />
      )}
    </div>
  )
}

function DashboardView({
  analytics,
  estimated,
  txs,
  truncated,
  prefs,
  onPrefs,
  upcoming,
  onSeeAll,
}: {
  analytics: Analytics
  estimated: number
  txs: Transaction[]
  truncated: boolean
  prefs: ChartPrefs
  onPrefs: (next: ChartPrefs) => void
  upcoming: { name: string; date: string; amount: number }[]
  onSeeAll: () => void
}) {
  const anchor = useMemo(() => parseISODate(prefs.anchor), [prefs.anchor])
  const series = useMemo(() => buildSeries(txs, prefs.mode, anchor), [txs, prefs.mode, anchor])
  const sums = useMemo(() => periodSums(txs, prefs.mode, anchor), [txs, prefs.mode, anchor])
  const label = periodLabel(prefs.mode, anchor)
  const categories = useMemo(() => categoryTotals(txs, prefs.mode, anchor), [txs, prefs.mode, anchor])
  const topCategories = categories.slice(0, 5)
  const rest = categories.slice(5).reduce((sum, [, value]) => sum + value, 0)
  const donutSource = rest > 0 ? [...topCategories, ['Остальное', rest] as [string, number]] : topCategories
  const donut = donutSource.map(([name, value], i) => ({
    name,
    value,
    color: CATEGORY_COLORS[i % CATEGORY_COLORS.length],
  }))
  const empty = sums.income === 0 && sums.expense === 0
  const latest = latestAnchor(txs)
  const bounds = periodBounds(prefs.mode, anchor)
  const latestDate = latest ? parseISODate(latest) : null
  const latestOutside = latestDate != null && (latestDate < bounds.from || latestDate >= bounds.to)

  function setMode(mode: ChartMode) {
    onPrefs({ ...prefs, mode })
  }

  function move(dir: -1 | 1) {
    onPrefs({ ...prefs, anchor: isoDate(shiftAnchor(prefs.mode, anchor, dir)) })
  }

  return (
    <>
      <section className="kpi-row">
        <Kpi label="Доходы" hint={label} value={rub(sums.income)} tone="up" />
        <Kpi label="Расходы" hint={label} value={rub(sums.expense)} tone="down" />
        <Kpi label="Капитал" hint="все счета" value={rub(analytics.total_balance)} tone="plus" />
        <Kpi label="Прогноз на месяц" hint="с учётом бюджета" value={rub(estimated)} tone="plus" />
      </section>

      <section className="dash-grid">
        <div className="card">
          <div className="card-head chart-head">
            <div>
              <h3>{chartTitle(prefs.mode)}</h3>
              <p className="muted small">{label}</p>
            </div>
            <div className="chart-tools">
              <div className="seg" role="tablist" aria-label="Период графика">
                {CHART_MODES.map((mode) => (
                  <button
                    key={mode.id}
                    type="button"
                    role="tab"
                    aria-selected={prefs.mode === mode.id}
                    className={prefs.mode === mode.id ? 'active' : undefined}
                    onClick={() => setMode(mode.id)}
                  >
                    {mode.label}
                  </button>
                ))}
              </div>
              <div className="chart-nav">
                <button type="button" onClick={() => move(-1)} aria-label="Предыдущий период">
                  ‹
                </button>
                <button type="button" onClick={() => move(1)} aria-label="Следующий период">
                  ›
                </button>
              </div>
            </div>
          </div>
          <div className="chart-box">
            <div className="chart-wrap">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={series} margin={{ top: 10, right: 8, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="spendFill" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#7eb6ff" stopOpacity={0.35} />
                      <stop offset="100%" stopColor="#7eb6ff" stopOpacity={0} />
                    </linearGradient>
                    <linearGradient id="incomeFill" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#3dd68c" stopOpacity={0.28} />
                      <stop offset="100%" stopColor="#3dd68c" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid stroke="#2a3042" strokeDasharray="3 6" vertical={false} />
                  <XAxis
                    dataKey="label"
                    stroke="#8b93a7"
                    tickLine={false}
                    axisLine={false}
                    interval={prefs.mode === 'month' ? 2 : 0}
                    tick={{ fontSize: 12 }}
                  />
                  <YAxis
                    stroke="#8b93a7"
                    tickLine={false}
                    axisLine={false}
                    width={64}
                    tickFormatter={(v) => formatAxisMoney(Number(v))}
                  />
                  <Tooltip
                    contentStyle={{
                      background: '#171a24',
                      border: '1px solid #2a3042',
                      borderRadius: 12,
                    }}
                    formatter={(value, name) => [rub(Number(value ?? 0)), name]}
                    labelStyle={{ color: '#8b93a7' }}
                  />
                  <Legend wrapperStyle={{ color: '#8b93a7' }} />
                  <Area
                    type="monotone"
                    dataKey="income"
                    name="Доходы"
                    stroke="#3dd68c"
                    strokeWidth={2}
                    fill="url(#incomeFill)"
                    dot={false}
                  />
                  <Area
                    type="monotone"
                    dataKey="expense"
                    name="Расходы"
                    stroke="#9ec5ff"
                    strokeWidth={3}
                    fill="url(#spendFill)"
                    dot={prefs.mode === 'week' ? { r: 4, fill: '#fff', stroke: '#9ec5ff', strokeWidth: 2 } : false}
                    activeDot={{ r: 6 }}
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
            {empty && (
              <div className="chart-empty">
                <span>В этом периоде нет доходов и расходов</span>
                {latest && latestOutside && (
                  <button type="button" className="btn small" onClick={() => onPrefs({ ...prefs, anchor: latest })}>
                    Показать последние операции
                  </button>
                )}
              </div>
            )}
          </div>
          <div className="chart-summary">
            <span>
              Расходы <strong>{rub(sums.expense)}</strong>
            </span>
            <span>
              Доходы <strong>{rub(sums.income)}</strong>
            </span>
            <span>
              Итого <strong>{rub(sums.net)}</strong>
            </span>
          </div>
          <p className="muted small chart-note">
            Переводы между своими счетами не входят в график.
            {truncated ? ' Показаны последние 20 000 операций.' : ''}
          </p>
        </div>

        <div className="card">
          <div className="card-head">
            <h3>По категориям</h3>
          </div>
          <p className="muted small chart-note">{label}</p>
          {donut.length === 0 ? (
            <p className="muted">Пока нет расходов за период</p>
          ) : (
            <>
              <div className="donut-wrap">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={donut}
                      dataKey="value"
                      nameKey="name"
                      innerRadius={48}
                      outerRadius={72}
                      paddingAngle={3}
                    >
                      {donut.map((d) => (
                        <Cell key={d.name} fill={d.color} stroke="transparent" />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        background: '#171a24',
                        border: '1px solid #2a3042',
                        borderRadius: 12,
                      }}
                      formatter={(value) => rub(Number(value ?? 0))}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <ul className="legend">
                {donut.map((d) => (
                  <li key={d.name}>
                    <span className="legend-left">
                      <span className="dot" style={{ background: d.color }} />
                      {d.name}
                    </span>
                    <strong>{rub(d.value)}</strong>
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>

        <div className="card span-right">
          <div className="card-head">
            <h3>Ближайшие платежи</h3>
            <div className="month-select">
              <IconCalendar />
              {new Date().toLocaleString('ru-RU', { month: 'long' })}
            </div>
          </div>
          {upcoming.length === 0 ? (
            <p className="muted">Добавьте регулярные траты во вкладке «Бюджет»</p>
          ) : (
            <ul className="bills">
              {upcoming.map((b) => (
                <li key={`${b.name}-${b.date}`} className="bill">
                  <div>
                    <strong>{b.name}</strong>
                    <span className="muted small">{b.date}</span>
                  </div>
                  <strong>{rub(b.amount)}</strong>
                </li>
              ))}
            </ul>
          )}
          <button className="linkish" onClick={onSeeAll}>
            Смотреть все
          </button>
        </div>
      </section>
    </>
  )
}

function Kpi({
  label,
  hint,
  value,
  tone,
}: {
  label: string
  hint?: string
  value: string
  tone: 'up' | 'down' | 'plus'
}) {
  return (
    <div className="kpi">
      <div className="kpi-top">
        <span className="kpi-label">{label}</span>
        <span className={`kpi-icon ${tone}`}>
          {tone === 'up' ? '↗' : tone === 'down' ? '↘' : '+'}
        </span>
      </div>
      <div className="kpi-value">{value}</div>
      {hint && <div className="kpi-hint">{hint}</div>}
    </div>
  )
}

function AccountsPanel({
  accounts,
  onChanged,
}: {
  accounts: Account[]
  onChanged: () => Promise<unknown>
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
    <section className="page-grid">
      <div className="card">
        <h3>Мои счета</h3>
        <ul className="list">
          {accounts.map((a) => (
            <li key={a.id}>
              <div>
                <strong>{a.name}</strong>
                <div className="muted small">
                  {accountTypeLabel[a.type]}
                  {a.bank_name ? ` · ${a.bank_name}` : ''}
                </div>
              </div>
              <strong>{rub(a.balance)}</strong>
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
  onChanged: () => Promise<unknown>
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
        `Импортировано ${batch.rows_imported} из ${batch.rows_total} (пропущено ${batch.rows_skipped}), банк: ${batch.bank}` +
          (batch.rows_imported > 0 ? '. График переключён на период новых операций.' : ''),
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
    <section className="page-grid">
      <div className="card">
        <h3>Загрузить выписку</h3>
        <p className="muted">
          CSV/XLSX банков РФ. Переводы между своими счетами не считаются расходами.
        </p>
        <form className="form" onSubmit={submit}>
          <label>
            Счёт
            <select value={accountId} onChange={(e) => setAccountId(Number(e.target.value))} required>
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
              <option value="alfa">Альфа-Банк</option>
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
  onChanged: () => Promise<unknown>
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
    <section className="page-grid">
      <div className="card">
        <h3>Бюджет</h3>
        <ul className="list">
          {items.map((b) => (
            <li key={b.id}>
              <div>
                <strong>{b.name}</strong>
                <div className="muted small">
                  {budgetKindLabel[b.kind]}
                  {b.day_of_month ? ` · день ${b.day_of_month}` : ''}
                  {b.due_date ? ` · до ${b.due_date.slice(0, 10)}` : ''}
                </div>
              </div>
              <div className="row-actions">
                <strong>{rub(b.amount)}</strong>
                <button className="btn ghost small" onClick={() => void api.deleteBudget(b.id).then(onChanged)}>
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

function GoalsPanel({ goals, onChanged }: { goals: Goal[]; onChanged: () => Promise<unknown> }) {
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
    <section className="page-grid">
      <div className="card">
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
                  <button className="btn ghost small" onClick={() => void api.deleteGoal(g.id).then(onChanged)}>
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
    <section className="page-grid">
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
        <div className="card">
          <h3>
            Сценарий: {result.name} (−{rub(result.amount)})
          </h3>
          <div className="forecast-row">
            {result.baseline.points.map((p, idx) => {
              const withP = result.with_purchase.points[idx]
              return (
                <div key={p.months_ahead} className="forecast-item">
                  <div className="muted small">через {p.months_ahead} мес.</div>
                  <div className="muted small">без покупки</div>
                  <div className="kpi-value" style={{ fontSize: '1.1rem' }}>
                    {rubShort(p.balance)}
                  </div>
                  <div className="muted small">с покупкой</div>
                  <div className="kpi-value bad" style={{ fontSize: '1.1rem' }}>
                    {rubShort(withP.balance)}
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

function ManualModal({
  accounts,
  onClose,
  onSaved,
}: {
  accounts: Account[]
  onClose: () => void
  onSaved: () => Promise<void>
}) {
  const [accountId, setAccountId] = useState(accounts[0]?.id ?? 0)
  const [type, setType] = useState<'expense' | 'income' | 'transfer'>('expense')
  const [amount, setAmount] = useState('')
  const [description, setDescription] = useState('')
  const [category, setCategory] = useState('Другое')
  const [transferId, setTransferId] = useState(accounts[1]?.id ?? 0)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!accountId) return
    setBusy(true)
    setErr(null)
    try {
      await api.createTransaction({
        account_id: accountId,
        type,
        amount: Math.round(parseFloat(amount.replace(',', '.')) * 100),
        description,
        category: type === 'transfer' ? 'Перевод себе' : category,
        transfer_account_id: type === 'transfer' ? transferId || undefined : undefined,
      })
      await onSaved()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'ошибка')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h3>Добавить вручную</h3>
        <form className="form" onSubmit={submit}>
          <label>
            Тип
            <select value={type} onChange={(e) => setType(e.target.value as typeof type)}>
              <option value="expense">Расход</option>
              <option value="income">Доход</option>
              <option value="transfer">Перевод</option>
            </select>
          </label>
          <label>
            Счёт
            <select value={accountId} onChange={(e) => setAccountId(Number(e.target.value))}>
              {accounts.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
            </select>
          </label>
          {type === 'transfer' && (
            <label>
              Куда
              <select value={transferId} onChange={(e) => setTransferId(Number(e.target.value))}>
                {accounts.map((a) => (
                  <option key={a.id} value={a.id}>
                    {a.name}
                  </option>
                ))}
              </select>
            </label>
          )}
          <label>
            Сумма, ₽
            <input value={amount} onChange={(e) => setAmount(e.target.value)} required />
          </label>
          <label>
            Описание
            <input value={description} onChange={(e) => setDescription(e.target.value)} />
          </label>
          {type !== 'transfer' && (
            <label>
              Категория
              <input value={category} onChange={(e) => setCategory(e.target.value)} />
            </label>
          )}
          {err && <div className="banner error">{err}</div>}
          <div className="modal-actions">
            <button type="button" className="btn ghost" onClick={onClose}>
              Отмена
            </button>
            <button className="btn" disabled={busy}>
              Сохранить
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function buildUpcomingBills(items: BudgetItem[]) {
  const now = new Date()
  const out: { name: string; date: string; amount: number }[] = []
  for (const b of items) {
    if (!b.is_active) continue
    if (b.kind === 'recurring_expense') {
      const day = b.day_of_month ?? 1
      const due = new Date(now.getFullYear(), now.getMonth(), day)
      if (due < now) due.setMonth(due.getMonth() + 1)
      out.push({
        name: b.name,
        date: due.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' }),
        amount: b.amount,
      })
    }
    if (b.kind === 'large_purchase' && b.due_date) {
      out.push({
        name: b.name,
        date: b.due_date.slice(0, 10).split('-').reverse().slice(0, 2).join('-'),
        amount: b.amount,
      })
    }
  }
  return out.slice(0, 6)
}

function IconGrid() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <rect x="3" y="3" width="7" height="7" rx="1.5" />
      <rect x="14" y="3" width="7" height="7" rx="1.5" />
      <rect x="3" y="14" width="7" height="7" rx="1.5" />
      <rect x="14" y="14" width="7" height="7" rx="1.5" />
    </svg>
  )
}
function IconCards() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <rect x="2" y="5" width="20" height="14" rx="2" />
      <path d="M2 10h20" />
    </svg>
  )
}
function IconUpload() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <path d="M12 16V4" />
      <path d="M7 9l5-5 5 5" />
      <path d="M4 20h16" />
    </svg>
  )
}
function IconBills() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <path d="M7 3h10v18l-2-1-2 1-2-1-2 1-2-1-2 1V3z" />
      <path d="M9 8h6M9 12h6M9 16h4" />
    </svg>
  )
}
function IconTarget() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <circle cx="12" cy="12" r="8" />
      <circle cx="12" cy="12" r="4" />
      <circle cx="12" cy="12" r="1" />
    </svg>
  )
}
function IconSpark() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <path d="M12 3v4M12 17v4M3 12h4M17 12h4M6 6l2.5 2.5M15.5 15.5L18 18M18 6l-2.5 2.5M6 18l2.5-2.5" />
    </svg>
  )
}
function IconRefresh() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <path d="M20 12a8 8 0 1 1-2.3-5.7" />
      <path d="M20 4v5h-5" />
    </svg>
  )
}
function IconBell() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <path d="M6 9a6 6 0 1 1 12 0c0 7 3 7 3 7H3s3 0 3-7" />
      <path d="M10 19a2 2 0 0 0 4 0" />
    </svg>
  )
}
function IconCalendar() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <rect x="3" y="5" width="18" height="16" rx="2" />
      <path d="M3 10h18M8 3v4M16 3v4" />
    </svg>
  )
}
