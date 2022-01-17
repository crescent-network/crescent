package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func newBuyOrder(price string, amount int64) *types.Order {
	return types.NewOrder(types.SwapDirectionBuy, newDec(price), sdk.NewInt(amount))
}

//nolint
func newSellOrder(price string, amount int64) *types.Order {
	return types.NewOrder(types.SwapDirectionSell, newDec(price), sdk.NewInt(amount))
}

func newDec(s string) sdk.Dec {
	return sdk.MustNewDecFromStr(s)
}

func parseCoin(s string) sdk.Coin {
	coin, err := sdk.ParseCoinNormalized(s)
	if err != nil {
		panic(err)
	}
	return coin
}

func parseCoins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return coins
}
