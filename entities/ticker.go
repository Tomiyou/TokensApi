/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package entities

import (
	"github.com/shopspring/decimal"
)

type TickerResp struct {
	Base
	Bid        decimal.Decimal `json:"bid,string" description:"Current best bid."`
	Ask        decimal.Decimal `json:"ask,string" description:"Current best bid."`
	Low        decimal.Decimal `json:"low,string" description:"Lowest value of requested interval."`
	High       decimal.Decimal `json:"high,string" description:"Highest value of requested interval."`
	Vwap       decimal.Decimal `json:"vwap,string" description:"Volume weighted average."`
	Volume     decimal.Decimal `json:"volume,string" description:"Volume in the requested interval"`
	VolumeUsdt decimal.Decimal `json:"volume_usdt,string" description:"Volume in the requested interval (in USDT)"`
}
