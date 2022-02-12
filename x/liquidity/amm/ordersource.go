package amm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OrderView interface {
	HighestBuyPrice() (sdk.Dec, bool)
	LowestSellPrice() (sdk.Dec, bool)
	BuyAmountOver(price sdk.Dec) sdk.Int
	SellAmountUnder(price sdk.Dec) sdk.Int
}

type OrderSource interface {
	OrderView
	BuyOrdersOver(price sdk.Dec) []Order
	SellOrdersUnder(price sdk.Dec) []Order
}
