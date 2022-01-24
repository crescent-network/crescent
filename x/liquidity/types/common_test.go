package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func newBuyOrder(price sdk.Dec, baseCoinAmt sdk.Int) *types.BaseOrder {
	return types.NewBaseOrder(types.SwapDirectionBuy, price, baseCoinAmt, price.MulInt(baseCoinAmt).TruncateInt())
}

//nolint
func newSellOrder(price sdk.Dec, baseCoinAmt sdk.Int) *types.BaseOrder {
	return types.NewBaseOrder(types.SwapDirectionSell, price, baseCoinAmt, baseCoinAmt)
}

func newInt(i int64) sdk.Int {
	return sdk.NewInt(i)
}

func parseDec(s string) sdk.Dec {
	return sdk.MustNewDecFromStr(s)
}
