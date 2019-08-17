/*
 * Copyright (C) 2019 Gregor Pogačnik
 */
package TokensApi

import (
	"flag"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/fiksn/TokensApi/entities"
	uuid "github.com/satori/go.uuid"
)

const (
	fiat   string = "usdt"
	crypto string = "btc"
)

var (
	e2e = flag.Bool("e2e", false, "Perform end to end testing")
)

func init() {
	flag.Parse()
}

func containsID(orders []entities.OpenOrder, id uuid.UUID) bool {

	for _, order := range orders {
		if order.Id == id {
			return true
		}
	}

	return false
}

func initCredentials() bool {
	credentials, err := parseJsonCfg("./credentials")
	if err != nil {
		fmt.Println("Your credentials are required for this test - place a file credentials into current directory")
		return false
	}

	Init(credentials.APIKey, credentials.APISecret)
	return true
}

func DisabledTestFilledOrder(t *testing.T) {
	if !*e2e {
		fmt.Println("End to end testing not perfomed you need to pass -e2e to go test")
		return
	}

	if !initCredentials() {
		return
	}

	id, err := uuid.FromString("28d6c834-b825-42f8-8117-9cb99439608d")
	if err != nil {
		t.Error("UUID conversion error", err)
		return
	}

	result, err := GetOrderDetails(id)
	if err != nil {
		t.Error("Could GetOrderDetails", err)
		return
	}

	fmt.Printf("Details %v\n", result)
}

func TestFiatToCryptoOrder(t *testing.T) {
	// WARNING: this might spend 5 USDT (but at least you'll get a very good price ;) )
	const (
		num = 5.0
		eps = 0.000001
	)

	if !*e2e {
		fmt.Println("End to end testing not perfomed you need to pass -e2e to go test")
		return
	}

	if !initCredentials() {
		return
	}

	orders, err := GetOrderBook(crypto + fiat)
	if err != nil {
		t.Error("GetOrderBook failed", err)
		return
	}

	balance, err := GetBalance(fiat)
	if err != nil {
		t.Error("GetBalance failed", err)
		return
	}

	available, err := balance.Available.Float64()
	fmt.Printf("Available balance %v %v\n", available, fiat)
	if err != nil {
		t.Error("GetBalance failed with no available money", err)
		return
	}

	if available < num {
		fmt.Printf("Not enough funds to conduct test %v %v < %v %v", available, fiat, num, fiat)
		return
	}

	fairAskPrice, err := orders.Asks[0][entities.Price].Float64()
	if err != nil {
		t.Error("Could not determine fair ask price", err)
		return
	}
	fairBidPrice, err := orders.Bids[0][entities.Price].Float64()
	if err != nil {
		t.Error("Could not determine fair bid price", err)
		return
	}

	// Slash the price in half (since I am buying use a ridiculous value) - /2 for average and another /2 for the slashing
	myPrice := (fairAskPrice + fairBidPrice) / 4.0
	volume := num / myPrice

	pairs, err := GetTradingPairs()
	if err != nil {
		t.Error("GetTradingPairs failed", err)
		return
	}

	pair := pairs[crypto+fiat]

	placement, err := PlaceOrderTyped(&pair, entities.Buy, Amount(volume), Price(myPrice), nil, nil)
	if err != nil {
		t.Error("PlaceOrderTyped failed", err)
	}

	fmt.Printf("Placed order %v volume = %v price = %v spent = %v %v\n", placement.OrderId, volume, myPrice, myPrice*volume, fiat)

	// After every action it might take a bit for it to get reflected
	time.Sleep(2 * time.Second)

	balance, err = GetBalance(fiat)
	if err != nil {
		t.Error("GetBalance failed", err)
		return
	}

	newAvailable, err := balance.Available.Float64()
	fmt.Printf("New available balance %v %v\n", newAvailable, fiat)

	if newAvailable >= available {
		t.Error("Before there were more available funds", available, newAvailable)
	}

	details, err := GetOrderDetails(placement.OrderId)
	if err != nil {
		t.Error("GetOrderDetails failed", err)
	}

	fmt.Printf("Details %v\n", details)

	if details.OrderStatus != entities.Open {
		t.Error("Order has wrong status", details.OrderStatus, entities.Open)
	}
	if details.OpenOrder.Id != placement.OrderId {
		t.Error("Order details has wrong id", details.OpenOrder.Id, placement.OrderId)
	}
	if details.OpenOrder.Type != entities.Buy {
		t.Error("Order details has wrong type", details.OpenOrder.Type)
	}

	theirPrice, err := details.OpenOrder.Price.Float64()
	if err != nil {
		t.Error("Conversion failed", err)
	}

	if theirPrice != myPrice {
		t.Error("Order details has wrong price", theirPrice, myPrice)
	}

	theirAmount, err := details.OpenOrder.Amount.Float64()
	if err != nil {
		t.Error("Conversion failed", err)
	}

	if math.Abs(theirAmount-volume) > eps {
		t.Error("Order details has wrong amount", theirAmount, volume)
	}

	theirAmount, err = details.OpenOrder.RemainingAmount.Float64()
	if err != nil {
		t.Error("Conversion failed", err)
	}
	if math.Abs(theirAmount-volume) > eps {
		t.Error("Order details has wrong remaining amount", theirAmount, volume)
	}

	if details.OpenOrder.CurrencyPair != crypto+fiat {
		t.Error("Order details has wrong currency pair", details.OpenOrder.CurrencyPair)
	}

	allOrders, err := GetAllOrders()
	if err != nil {
		t.Error("GetAllOrders failed", err)
	}

	if !containsID(allOrders.OpenOrders, placement.OrderId) {
		t.Error("Order was not found among all orders")
	}

	pairOrders, err := GetAllOrdersFor(crypto + fiat)
	if err != nil {
		t.Error("GetAllOrdersFor failed", err)
	}

	if !containsID(pairOrders.OpenOrders, placement.OrderId) {
		t.Error("Order was not found among pair orders")
	}

	// You might want to check the webinterface here
	time.Sleep(10 * time.Second)

	fmt.Println("Cancelling order")
	_, err = CancelOrder(placement.OrderId)
	if err != nil {
		t.Error("CancelOrder failed", err)
	}

	// After every action it might take a bit for it to get reflected
	time.Sleep(2 * time.Second)

	balance, err = GetBalance(fiat)
	if err != nil {
		t.Error("GetBalance failed", err)
		return
	}

	newAvailable, err = balance.Available.Float64()
	fmt.Printf("New available balance %v %v\n", newAvailable, fiat)
	if newAvailable != available {
		t.Error("After cancellation the same amount of funds should be available", available, newAvailable)
	}

	allOrders, err = GetAllOrders()
	if err != nil {
		t.Error("GetAllOrders failed", err)
	}

	if containsID(allOrders.OpenOrders, placement.OrderId) {
		t.Error("Order was found among all orders after cancellation")
	}

	pairOrders, err = GetAllOrdersFor(crypto + fiat)
	if err != nil {
		t.Error("GetAllOrdersFor failed", err)
	}

	if containsID(pairOrders.OpenOrders, placement.OrderId) {
		t.Error("Order was found among pair orders after cancellation")
	}

}
