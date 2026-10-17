package models

import "time"

type AccountType string

const (
	AccountCard      AccountType = "card"
	AccountSavings   AccountType = "savings"
	AccountBrokerage AccountType = "brokerage"
	AccountCash      AccountType = "cash"
)

type TxType string

const (
	TxIncome   TxType = "income"
	TxExpense  TxType = "expense"
	TxTransfer TxType = "transfer"
)

type BudgetKind string

const (
	BudgetPlannedIncome    BudgetKind = "planned_income"
	BudgetRecurringExpense BudgetKind = "recurring_expense"
	BudgetLargePurchase    BudgetKind = "large_purchase"
)

type User struct {
	ID         int       `json:"id"`
	Email      *string   `json:"email,omitempty"`
	Name       string    `json:"name"`
	TelegramID *string   `json:"telegram_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Account struct {
	ID        int         `json:"id"`
	UserID    int         `json:"user_id"`
	Name      string      `json:"name"`
	Type      AccountType `json:"type"`
	Balance   int64       `json:"balance"` // kopecks
	Currency  string      `json:"currency"`
	BankName  *string     `json:"bank_name,omitempty"`
	IsActive  bool        `json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
}

type Category struct {
	ID       int      `json:"id"`
	UserID   *int     `json:"user_id,omitempty"`
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Icon     *string  `json:"icon,omitempty"`
	Keywords []string `json:"keywords"`
}

type Transaction struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	AccountID         int       `json:"account_id"`
	TransferAccountID *int      `json:"transfer_account_id,omitempty"`
	Type              TxType    `json:"type"`
	Amount            int64     `json:"amount"` // positive, kopecks
	Description       string    `json:"description"`
	Category          string    `json:"category"`
	OccurredAt        time.Time `json:"occurred_at"`
	ExternalID        *string   `json:"external_id,omitempty"`
	ImportBatchID     *int      `json:"import_batch_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type ImportBatch struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	AccountID    *int      `json:"account_id,omitempty"`
	Filename     string    `json:"filename"`
	Bank         string    `json:"bank"`
	Status       string    `json:"status"`
	RowsTotal    int       `json:"rows_total"`
	RowsImported int       `json:"rows_imported"`
	RowsSkipped  int       `json:"rows_skipped"`
	CreatedAt    time.Time `json:"created_at"`
}

type BudgetItem struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Kind       BudgetKind `json:"kind"`
	Name       string     `json:"name"`
	Amount     int64      `json:"amount"`
	Category   *string    `json:"category,omitempty"`
	DayOfMonth *int       `json:"day_of_month,omitempty"`
	DueDate    *time.Time `json:"due_date,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Goal struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	Name          string     `json:"name"`
	TargetAmount  int64      `json:"target_amount"`
	CurrentAmount int64      `json:"current_amount"`
	Deadline      *time.Time `json:"deadline,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Analytics DTOs

type AnalyticsSummary struct {
	TotalBalance      int64            `json:"total_balance"`
	IncomePeriod      int64            `json:"income_period"`
	ExpensePeriod     int64            `json:"expense_period"`
	NetPeriod         int64            `json:"net_period"`
	ByCategoryExpense map[string]int64 `json:"by_category_expense"`
	ByCategoryIncome  map[string]int64 `json:"by_category_income"`
	Accounts          []Account        `json:"accounts"`
}

type ForecastPoint struct {
	MonthsAhead int   `json:"months_ahead"`
	Balance     int64 `json:"balance"`
	Income      int64 `json:"income"`
	Expense     int64 `json:"expense"`
}

type ForecastResult struct {
	CurrentBalance int64           `json:"current_balance"`
	Points         []ForecastPoint `json:"points"`
}

type ScenarioRequest struct {
	Name        string     `json:"name"`
	Amount      int64      `json:"amount"`
	When        *time.Time `json:"when,omitempty"`
	MonthsAhead []int      `json:"months_ahead"`
}

type ScenarioResult struct {
	Name           string          `json:"name"`
	Amount         int64           `json:"amount"`
	Baseline       ForecastResult  `json:"baseline"`
	WithPurchase   ForecastResult  `json:"with_purchase"`
	DeltaByHorizon map[string]int64 `json:"delta_by_horizon"`
}
