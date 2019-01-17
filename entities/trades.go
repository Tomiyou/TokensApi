/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package entities

import "encoding/json"

type OrderType string

const (
	BUY  OrderType = "buy"
	SELL OrderType = "sell"
)

type TradesResp struct {
	Base
	Trades []Trade `json:"trades" description:"Timestamp"`
}

type Trade struct {
	Id       int64       `json:"id" description:"ID"`
	Datetime timestamp   `json:"datetime" description:"Timestamp of trade"`
	Price    json.Number `json:"price,string" description:"Price of trade."`
	Amount   json.Number `json:"amount,string" description:"Amount of trade."`
	Type     OrderType   `json:"type" description:"Either buy or sell"`
}
