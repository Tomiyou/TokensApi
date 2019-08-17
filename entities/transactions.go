/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package entities

import (
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	Deposit         TransactionType = "Deposit"
	Whitdrawal      TransactionType = "Whitdrawal"
	WithdrawalFee   TransactionType = "Whitdrawal fee"
	Trading         TransactionType = "Trade" // cannot reuse Trade
	TradingFee      TransactionType = "Trading fee"
	AccountTransfer TransactionType = "Account transfer"
	Vote            TransactionType = "Vote"
)

type TransactionResp struct {
	Page
	Transactions []Transaction `json:"transactions" description:"Transactions"`
}

type Transaction struct {
	Id       int64           `json:"id" description:"ID"`
	Datetime Timestamp       `json:"datetime" description:"Timestamp of transaction"`
	Amount   decimal.Decimal     `json:"amount,string" description:"Amount of transaction."`
	Type     TransactionType `json:"type" description:"Type of transaction"`
	Currency string          `json:"currency" description:"Currency of transaction"`
}
