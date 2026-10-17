-- Personal Finance MVP schema

CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    email       TEXT UNIQUE,
    name        TEXT NOT NULL DEFAULT 'Demo User',
    telegram_id TEXT UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accounts (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    type       TEXT NOT NULL CHECK (type IN ('card', 'savings', 'brokerage', 'cash')),
    balance    BIGINT NOT NULL DEFAULT 0, -- kopecks
    currency   TEXT NOT NULL DEFAULT 'RUB',
    bank_name  TEXT,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
    id         SERIAL PRIMARY KEY,
    user_id    INT REFERENCES users(id) ON DELETE CASCADE, -- NULL = system default
    name       TEXT NOT NULL,
    kind       TEXT NOT NULL CHECK (kind IN ('expense', 'income', 'transfer')),
    icon       TEXT,
    keywords   TEXT[] NOT NULL DEFAULT '{}',
    UNIQUE (user_id, name, kind)
);

CREATE TABLE IF NOT EXISTS import_batches (
    id            SERIAL PRIMARY KEY,
    user_id       INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id    INT REFERENCES accounts(id) ON DELETE SET NULL,
    filename      TEXT NOT NULL,
    bank          TEXT NOT NULL DEFAULT 'generic',
    status        TEXT NOT NULL DEFAULT 'done',
    rows_total    INT NOT NULL DEFAULT 0,
    rows_imported INT NOT NULL DEFAULT 0,
    rows_skipped  INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id                  SERIAL PRIMARY KEY,
    user_id             INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id          INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    transfer_account_id INT REFERENCES accounts(id) ON DELETE SET NULL,
    type                TEXT NOT NULL CHECK (type IN ('income', 'expense', 'transfer')),
    amount              BIGINT NOT NULL, -- always positive, kopecks
    description         TEXT NOT NULL DEFAULT '',
    category            TEXT NOT NULL DEFAULT 'Другое',
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    external_id         TEXT, -- for import dedup
    import_batch_id     INT REFERENCES import_batches(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_tx_user_date ON transactions(user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_tx_user_type ON transactions(user_id, type);
CREATE INDEX IF NOT EXISTS idx_tx_account ON transactions(account_id);

CREATE TABLE IF NOT EXISTS budget_items (
    id            SERIAL PRIMARY KEY,
    user_id       INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind          TEXT NOT NULL CHECK (kind IN ('planned_income', 'recurring_expense', 'large_purchase')),
    name          TEXT NOT NULL,
    amount        BIGINT NOT NULL,
    category      TEXT,
    day_of_month  INT CHECK (day_of_month IS NULL OR (day_of_month BETWEEN 1 AND 28)),
    due_date      DATE,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS goals (
    id             SERIAL PRIMARY KEY,
    user_id        INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    target_amount  BIGINT NOT NULL,
    current_amount BIGINT NOT NULL DEFAULT 0,
    deadline       DATE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Demo user + default accounts + system categories
INSERT INTO users (id, email, name)
VALUES (1, 'demo@local', 'Demo User')
ON CONFLICT DO NOTHING;

SELECT setval('users_id_seq', GREATEST((SELECT MAX(id) FROM users), 1));

INSERT INTO accounts (user_id, name, type, balance, bank_name)
SELECT 1, 'Карта', 'card', 0, 'Тинькофф'
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE user_id = 1 AND name = 'Карта');

INSERT INTO accounts (user_id, name, type, balance, bank_name)
SELECT 1, 'Накопительный', 'savings', 0, 'Тинькофф'
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE user_id = 1 AND name = 'Накопительный');

INSERT INTO accounts (user_id, name, type, balance, bank_name)
SELECT 1, 'Брокерский', 'brokerage', 0, 'Тинькофф Инвестиции'
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE user_id = 1 AND name = 'Брокерский');

INSERT INTO accounts (user_id, name, type, balance, bank_name)
SELECT 1, 'Наличные', 'cash', 0, NULL
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE user_id = 1 AND name = 'Наличные');

INSERT INTO categories (user_id, name, kind, keywords) VALUES
(NULL, 'Продукты', 'expense', ARRAY['пятерочка','перекресток','магнит','лента','вкусвилл','ашан','дикси','продукты']),
(NULL, 'Кафе и рестораны', 'expense', ARRAY['кофе','ресторан','кафе','яндекс еда','delivery','vkusno','бургер','суши','пицца']),
(NULL, 'Транспорт', 'expense', ARRAY['яндекс го','uber','такси','метро','транспорт','азс','бензин','заправ']),
(NULL, 'ЖКХ и связь', 'expense', ARRAY['жкх','квартира','мтс','мегафон','билайн','теле2','интернет','электричество','газ']),
(NULL, 'Здоровье', 'expense', ARRAY['аптека','клиника','стомат','врач','медицин']),
(NULL, 'Покупки', 'expense', ARRAY['wildberries','ozon','avito','lamoda','dns','мвидео','ситилинк']),
(NULL, 'Развлечения', 'expense', ARRAY['кино','spotify','netflix','иви','кинопоиск','steam','playstation']),
(NULL, 'Перевод себе', 'transfer', ARRAY['перевод себе','между своими','own transfer','internal']),
(NULL, 'Зарплата', 'income', ARRAY['зарплата','зп','salary','аванс']),
(NULL, 'Другое', 'expense', ARRAY[]::TEXT[]),
(NULL, 'Прочий доход', 'income', ARRAY[]::TEXT[])
ON CONFLICT DO NOTHING;
