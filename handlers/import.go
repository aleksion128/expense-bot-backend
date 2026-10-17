package handlers

import (
	"bot_api_microservice/middlewares"
	"bot_api_microservice/models"
	importsvc "bot_api_microservice/services/import"
	"bot_api_microservice/storage"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ImportStatement(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)

	accountID, err := strconv.Atoi(c.PostForm("account_id"))
	if err != nil || accountID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id required"})
		return
	}
	bank := c.DefaultPostForm("bank", "auto")
	markTransfers := c.DefaultPostForm("mark_transfers", "true") != "false"

	var transferID *int
	if v := c.PostForm("transfer_account_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			transferID = &n
		}
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required (CSV/XLSX)"})
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read file"})
		return
	}

	batch, err := importsvc.Import(db, importsvc.ImportRequest{
		UserID:            uid,
		AccountID:         accountID,
		TransferAccountID: transferID,
		Filename:          header.Filename,
		Bank:              bank,
		Data:              data,
		MarkTransfers:     markTransfers,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, batch)
}

func ListImports(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	items, err := storage.ListImportBatches(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []models.ImportBatch{}
	}
	c.JSON(http.StatusOK, items)
}
