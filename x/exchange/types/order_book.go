package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OrderBook struct {
	Sells []OrderBookPriceLevel
	Buys  []OrderBookPriceLevel
}

type OrderBookPriceLevel struct {
	Price    sdk.Dec
	Quantity sdk.Int
}
