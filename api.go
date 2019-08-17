/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package TokensApi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/fiksn/TokensApi/entities"

	"github.com/golang/glog"
	uuid "github.com/satori/go.uuid"
)

const (
	// Base URL for API calls
	TokensBaseUrl = "https://api.tokens.net"
	// Percentage fee for takers
	TakerFeePercent = 0.2
	// Percentage fee for makers
	MakerFeePercent = 0
)

type Interval int

const (
	Day = iota
	Hour
	Minute
)

/**
* List all existing pairs.
 */
func GetTradingPairs() (entities.TradingPairResp, error) {
	var resp entities.TradingPairResp

	jsonBlob := request(TokensBaseUrl + "/public/trading-pairs/get/all/")
	glog.V(5).Infof("GetTradingPairs resp %v", string(jsonBlob))

	err := json.Unmarshal(jsonBlob, &resp)
	if err != nil {
		glog.Warningf("GetTradingPairs unable to unmarshal json blob: %v (%v)", string(jsonBlob), err)
		return resp, err
	}

	return resp, nil
}

/**
* Get order book.
 */
func GetOrderBook(pair string) (entities.OrderBookResp, error) {
	var resp entities.OrderBookResp

	jsonBlob := request(TokensBaseUrl + fmt.Sprintf("/public/order-book/%s/", pair))
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetOrderBook resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	if err != nil {
		return resp, err
	}

	sort.Sort(entities.AskOrder(resp.Asks))
	sort.Sort(sort.Reverse(entities.AskOrder(resp.Bids)))

	return resp, nil
}

/**
* Get balance.
 */
func GetBalance(currency string) (entities.BalanceResp, error) {
	var resp entities.BalanceResp

	jsonBlob := requestAuth(TokensBaseUrl + fmt.Sprintf("/private/balance/%s/", currency))
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetBalance resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
* Get all balances.
 */
func GetAllBalances() (entities.AllBalanceResp, error) {
	var resp entities.AllBalanceResp

	jsonBlob := requestAuth(TokensBaseUrl + "/private/balance/all/")
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetAllBalances resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
* Get ticker for last day or hour.
 */
func GetTicker(pair string, interval Interval) (entities.TickerResp, error) {
	var (
		resp entities.TickerResp
		url  string
	)

	switch interval {
	case Hour:
		url = fmt.Sprintf("/public/ticker/hour/%s/", pair)
	case Day:
		url = fmt.Sprintf("/public/ticker/%s/", pair)
	default:
		return resp, errors.New("Illegal interval specified")
	}

	jsonBlob := request(TokensBaseUrl + url)
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetTicker resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
* List trades, which occured in last minute, hour or day.
 */
func GetTrades(pair string, interval Interval) (entities.TradesResp, error) {
	var (
		resp entities.TradesResp
		url  string
	)

	switch interval {
	case Hour:
		url = fmt.Sprintf("/public/trades/hour/%s/", pair)
	case Day:
		url = fmt.Sprintf("/public/trades/day/%s/", pair)
	default:
		return resp, errors.New("Illegal interval specified")
	}

	jsonBlob := request(TokensBaseUrl + url)
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetTrades resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
 * List all currencies participating in voting and number of votes for each currency.
 */
func GetVotes() (entities.VotesResp, error) {
	var resp entities.VotesResp

	jsonBlob := request(TokensBaseUrl + "/public/voting/get/all/")
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetVotes resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
 * Cancel an order by id.
 */
func CancelOrder(id uuid.UUID) (entities.Base, error) {
	var resp entities.Base

	jsonBlob := requestAuthPost(TokensBaseUrl+fmt.Sprintf("/private/orders/cancel/%s/", id), url.Values{})
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("CancelOrder resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
 * Place an order.
 */
func PlaceOrder(
	pair string,
	side entities.OrderType,
	amount decimal.Decimal,
	amountDecimals int32,
	price decimal.Decimal,
	priceDecimals int32,
	takeProfitPrice decimal.Decimal,
	expireDate *time.Time) (entities.PlaceOrderResp, error) {
	var resp entities.PlaceOrderResp

	resp.SetStatus("error")

	data := url.Values{}

	if amount.LessThanOrEqual(decimal.Zero) {
		return resp, errors.New("Negative amount is not allowed")
	}

	if price.LessThanOrEqual(decimal.Zero) {
		return resp, errors.New("Negative price is not allowed")
	}

	if side != entities.Buy && side != entities.Sell {
		return resp, errors.New("Only buy or sell orders are supported")
	}

	data.Add("tradingPair", pair)
	data.Add("side", string(side))
	data.Add("amount", amount.StringFixed(amountDecimals))
	data.Add("price", price.StringFixed(priceDecimals))

	if takeProfitPrice.GreaterThan(decimal.Zero) {
		data.Add("takeProfitPrice", takeProfitPrice.StringFixed(priceDecimals))
	}

	if expireDate != nil {
		data.Add("expireDate", strconv.FormatInt(expireDate.Unix(), 10))
	}

	jsonBlob := requestAuthPost(TokensBaseUrl+"/private/orders/add/limit/", data)
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("PlaceOrder resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
 * Get order details.
 */
func GetOrderDetails(id uuid.UUID) (entities.OrderDetailsResp, error) {
	var resp entities.OrderDetailsResp

	jsonBlob := requestAuth(TokensBaseUrl + fmt.Sprintf("/private/orders/get/%s/", id))
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetOrderDetails resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
 * Get all orders.
 */
func GetAllOrders() (entities.OrdersResp, error) {
	var resp entities.OrdersResp

	jsonBlob := requestAuth(TokensBaseUrl + "/private/orders/get/all/")
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetAllOrders resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
 * Get all orders for currency pair.
 */
func GetAllOrdersFor(pair string) (entities.OrdersResp, error) {
	var resp entities.OrdersResp

	jsonBlob := requestAuth(TokensBaseUrl + fmt.Sprintf("/private/orders/get/%s/", pair))
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetAllOrdersFor resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}

/**
 * Get transactions. Page parameter starts from 1 and goes to max (which you get from TransactionResp.TotalPages)
 */
func GetTransactions(page int) (entities.TransactionResp, error) {
	var resp entities.TransactionResp

	jsonBlob := requestAuth(TokensBaseUrl + fmt.Sprintf("/private/transactions/%d/", page))
	if jsonBlob == nil {
		return resp, errors.New("No response")
	}

	glog.V(5).Infof("GetTransactions resp %v", string(jsonBlob))

	err := deserialize(jsonBlob, &resp)
	return resp, err
}
