/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package entities

import (
	"github.com/shopspring/decimal"
)

type TradingPair struct {
	PriceDecimals   int32         `json:"priceDecimals" description:"Decimals for price"`
	AmountDecimals  int32         `json:"amountDecimals" description:"Decimals for amount"`
	MinAmount       decimal.Decimal `json:"minAmount" description:"Minimum amount of base currency."`
	BaseCurrency    string      `json:"baseCurrency" description:"Base currency."`
	CounterCurrency string      `json:"counterCurrency" description:"Counter currency."`
	Title           string      `json:"title" description:"Title."`
}

type TradingPairResp map[string]TradingPair

func (me TradingPair) String() string {
	return "Trading pair " + me.Title
}
