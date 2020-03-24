package entities

type DepositAddressResp struct {
	Base
	Address string `json:"address" description:"Deposit address"`
}
