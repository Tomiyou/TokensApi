/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package entities

import (
	"github.com/shopspring/decimal"

	"github.com/satori/go.uuid"
)

type OrderType string

const (
	Buy  OrderType = "buy"
	Sell OrderType = "sell"
)

type OrderDetails string

const (
	Open     OrderDetails = "open"
	Filled   OrderDetails = "filled"
	Canceled OrderDetails = "canceled" // sic
	Expired  OrderDetails = "expired"
)

type OrdersResp struct {
	Base
	// All orders here are Open
	OpenOrders []OpenOrder `json:"openOrders" description:"Open orders."`
}

type OpenOrder struct {
	Id              uuid.UUID   `json:"id" description:"ID"`
	Created         Timestamp   `json:"created" description:"Timestamp of order"`
	Type            OrderType   `json:"type" description:"Either buy or sell"`
	Price           decimal.Decimal `json:"price,string" description:"Price of order."`
	TakeProfitPrice decimal.Decimal `json:"takeeprofit,string" description:"Price to take profit."`
	Amount          decimal.Decimal `json:"amount,string" description:"Amount of order."`
	RemainingAmount decimal.Decimal `json:"remainingAmount,string" description:"Remaining amount of order."`
	CurrencyPair    string      `json:"currencyPair" description:"Currency pair of order."`
}

type OrderDetailsResp struct {
	Base
	// This order might have a different status than "Open", check OrderStatus
	OpenOrder
	Trades []ExtendedTrade `json:"trades" description:"Trades."`

	OrderStatus OrderDetails `json:"orderStatus" description:"Details about the order"`
}

type ExtendedTrade struct {
	Trade
	Value decimal.Decimal `json:"value,string" description:"Value is price * amount, however note that currenly it is in limited number of decimals (same as price?)"`
}

type PlaceOrderResp struct {
	Base
	OrderId uuid.UUID `json:"orderId" description:"ID"`
}
