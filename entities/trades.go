/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package entities

import (
	"github.com/shopspring/decimal"
)

type TradesResp struct {
	Base
	Trades []Trade `json:"trades" description:"Timestamp"`
}

type Trade struct {
	Id       int64       `json:"id" description:"ID"`
	Datetime Timestamp   `json:"datetime" description:"Timestamp of trade"`
	Price    decimal.Decimal `json:"price,string" description:"Price of trade."`
	Amount   decimal.Decimal `json:"amount,string" description:"Amount of trade."`
	Type     OrderType   `json:"type" description:"Either buy or sell"`
}
