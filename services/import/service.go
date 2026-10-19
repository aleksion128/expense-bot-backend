package importsvc

import (
	"bot_api_microservice/models"
	"bot_api_microservice/services/categorize"
	"bot_api_microservice/storage"
	"database/sql"
	"fmt"
)

type ImportRequest struct {
	UserID            int
	AccountID         int
	TransferAccountID *int
	Filename          string
	Bank              string
	Data              []byte
	MarkTransfers     bool
}

func Import(db *sql.DB, req ImportRequest) (models.ImportBatch, error) {
	if _, err := storage.GetAccount(db, req.UserID, req.AccountID); err != nil {
		return models.ImportBatch{}, fmt.Errorf("account not found")
	}

	parsed, err := ParseFile(req.Filename, req.Data, req.Bank)
	if err != nil {
		return models.ImportBatch{}, err
	}

	cats, err := storage.ListCategories(db, req.UserID)
	if err != nil {
		return models.ImportBatch{}, err
	}
	accounts, _ := storage.ListAccounts(db, req.UserID)
	var hints []string
	for _, a := range accounts {
		hints = append(hints, a.Name)
		if a.BankName != nil {
			hints = append(hints, *a.BankName)
		}
	}

	batch, err := storage.CreateImportBatch(db, models.ImportBatch{
		UserID:    req.UserID,
		AccountID: &req.AccountID,
		Filename:  req.Filename,
		Bank:      parsed.Bank,
		Status:    "processing",
	})
	if err != nil {
		return batch, err
	}

	tx, err := db.Begin()
	if err != nil {
		return batch, err
	}
	defer func() { _ = tx.Rollback() }()

	imported, skipped := 0, 0
	for _, row := range parsed.Rows {
		txType := models.TxExpense
		amount := row.Amount
		if amount > 0 {
			txType = models.TxIncome
		} else {
			amount = -amount
		}

		var transferTo *int
		if req.MarkTransfers && categorize.DetectTransfer(row.Description, hints) {
			txType = models.TxTransfer
			if req.TransferAccountID != nil {
				transferTo = req.TransferAccountID
			} else {
				// fallback: first non-source account
				for _, a := range accounts {
					if a.ID != req.AccountID {
						id := a.ID
						transferTo = &id
						break
					}
				}
			}
			if transferTo == nil {
				// cannot book transfer without destination — skip as transfer expense neutral
				skipped++
				continue
			}
		}

		cat := categorize.Categorize(row.Description, txType, cats)
		if row.Category != "" && (cat == "Другое" || cat == "Прочий доход") {
			cat = row.Category
		}
		desc := row.Description
		if desc == "" {
			desc = row.Category
		}
		ext := row.ExternalID
		t := models.Transaction{
			UserID:            req.UserID,
			AccountID:         req.AccountID,
			TransferAccountID: transferTo,
			Type:              txType,
			Amount:            amount,
			Description:       desc,
			Category:          cat,
			OccurredAt:        row.OccurredAt,
			ExternalID:        &ext,
			ImportBatchID:     &batch.ID,
		}
		ok, err := storage.InsertImportedTransaction(tx, t)
		if err != nil {
			return batch, err
		}
		if ok {
			imported++
		} else {
			skipped++
		}
	}

	total := len(parsed.Rows)
	if err := storage.UpdateImportBatchStats(tx, batch.ID, total, imported, skipped, "done"); err != nil {
		return batch, err
	}
	if err := tx.Commit(); err != nil {
		return batch, err
	}

	batch.RowsTotal = total
	batch.RowsImported = imported
	batch.RowsSkipped = skipped
	batch.Status = "done"
	batch.Bank = parsed.Bank
	return batch, nil
}
