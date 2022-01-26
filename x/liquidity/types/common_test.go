package types_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func newBuyOrder(price sdk.Dec, amt sdk.Int) *types.BaseOrder {
	return types.NewBaseOrder(types.SwapDirectionBuy, price, amt, price.MulInt(amt).TruncateInt())
}

func newSellOrder(price sdk.Dec, amt sdk.Int) *types.BaseOrder {
	return types.NewBaseOrder(types.SwapDirectionSell, price, amt, amt)
}

func newInt(i int64) sdk.Int {
	return sdk.NewInt(i)
}

func parseDec(s string) sdk.Dec {
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

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
