/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package entities

import (
	"fmt"
	"github.com/shopspring/decimal"
)

type BalanceResp struct {
	Base
	Currency string `json:"currency" description:"Currency"`
	Balance
}

type Balance struct {
	Total     decimal.Decimal `json:"total,string" description:"Total amount."`
	Available decimal.Decimal `json:"available,string" description:"Available amount."`
}

type Balances map[string]*Balance

type AllBalanceResp struct {
	Base
	Balances `json:"balances" description:"Balances"`
}

func (me BalanceResp) String() string {
	return fmt.Sprintf("Balance %v %v/%v", me.Currency, me.Available, me.Total)
}
