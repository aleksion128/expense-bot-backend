# Budgetly — Personal Finance MVP

Веб-приложение для учёта личных финансов с импортом CSV/XLSX выписок банков РФ.

## Что умеет MVP

- **Счета капитала**: карта, накопительный, брокерский, наличные
- **Импорт выписок**: CSV/XLSX (Тинькофф, Альфа-Банк с преамбулой, универсальный), автокатегоризация
- **Корректный учёт**: переводы между своими счетами ≠ расходы
- **Бюджет**: плановые доходы, регулярные траты, крупные покупки
- **Прогноз**: остаток через 1 / 3 / 6 месяцев
- **Сценарии**: «А что, если куплю объектив за 20 000?»
- **Frontend**: React + Vite

## Быстрый старт

### Docker Compose

```bash
docker compose up --build
```

- API + UI (production static): http://localhost:8443  
- Dev frontend: http://localhost:5173 (проксирует `/api` на API)

### Локально (без Docker)

1. Поднимите Postgres и задайте env:

```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=finance
export POSTGRES_PASSWORD=finance
export POSTGRES_DB=finance
```

2. API:

```bash
go run .
```

3. Frontend:

```bash
cd frontend && npm install && npm run dev
```

## API (кратко)

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/accounts` | Счета |
| POST | `/api/v1/import` | multipart: `file`, `account_id`, `bank`, `transfer_account_id` |
| GET | `/api/v1/analytics` | Доходы/расходы/капитал |
| GET | `/api/v1/forecast` | Прогноз 1/3/6 мес. |
| POST | `/api/v1/scenarios` | What-if сценарий |
| GET/POST | `/api/v1/budget` | Бюджет |
| GET/POST | `/api/v1/goals` | Цели |

Демо-пользователь: `X-User-Id: 1` (по умолчанию).

Суммы в API — **копейки** (`int64`).

## Пример импорта

Файл: `sample_data/tinkoff_sample.csv`

```bash
curl -F file=@sample_data/tinkoff_sample.csv \
  -F account_id=1 \
  -F transfer_account_id=2 \
  -F bank=tinkoff \
  http://localhost:8443/api/v1/import
```

## Структура

```
├── main.go / handlers / storage / services   # Go API
├── migrations/                               # SQL
├── frontend/                                 # React UI
├── sample_data/                              # Пример выписки
└── docs/MVP.md                               # План и gap-анализ
```

## История репо

Раньше это был бэкенд Telegram-бота (`expense-bot-backend`). Ветка одна — `master`.
Логика расходов/доходов/целей переработана в единый журнал транзакций + веб-UI.
