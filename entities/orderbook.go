/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package entities

import (
	"github.com/shopspring/decimal"

	"github.com/golang/glog"
)

type ColumnQuotation int

const (
	Volume ColumnQuotation = 0
	Price  ColumnQuotation = 1
)

// volume, price
type Quotation [][2]decimal.Decimal

type OrderBook struct {
	Bids Quotation `json:"bids" description:"Bids (buy requests)"`
	Asks Quotation `json:"asks" description:"Asks (sell requests)"`
}

type OrderBookResp struct {
	Base
	Bids Quotation `json:"bids" description:"Bids (buy requests)"`
	Asks Quotation `json:"asks" description:"Asks (sell requests)"`
}

type AskOrder Quotation // also ascending

func (v AskOrder) Len() int      { return len(v) }
func (v AskOrder) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (a AskOrder) Less(i, j int) bool {
	first, err := decimal.NewFromString(a[i][Price].String())
	second, err := decimal.NewFromString(a[j][Price].String())

	if err != nil {
		return false
	}

	return first.Cmp(second) < 0
}

/**
* Get liquidity in counter currency (e.g., USDT).
 */
func (me *Quotation) GetLiquidity() decimal.Decimal {
	sum := decimal.New(0, 0)

	for _, val := range *me {

		onePrice, err := decimal.NewFromString(val[Price].String())
		if err != nil {
			glog.Warningf("GetLiquidity fatal error %v", err)
			return sum
		}

		oneVolume, err := decimal.NewFromString(val[Volume].String())
		if err != nil {
			glog.Warningf("GetLiquidity fatal error %v", err)
			return sum
		}

		// sum += (oneVolume * onePrice)
		sum = sum.Add(oneVolume.Mul(onePrice))
	}

	return sum
}

func (me *Quotation) Copy() *Quotation {
	created := make(Quotation, len(*me))
	copy(created, *me)
	return &created
}
