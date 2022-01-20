package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func newBuyOrder(price string, amount int64) *types.BaseOrder {
	return types.NewBaseOrder(types.SwapDirectionBuy, parseDec(price), sdk.NewInt(amount))
}

//nolint
func newSellOrder(price string, amount int64) *types.BaseOrder {
	return types.NewBaseOrder(types.SwapDirectionSell, parseDec(price), sdk.NewInt(amount))
}

func parseDec(s string) sdk.Dec {
	return sdk.MustNewDecFromStr(s)
}
