package types_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

var testAddr = sdk.AccAddress(crypto.AddressHash([]byte("test")))

func newBuyOrder(price sdk.Dec, amt sdk.Int) *types.BaseOrder {
	return types.NewBaseOrder(types.SwapDirectionBuy, price, amt, price.MulInt(amt).TruncateInt())
}

func newSellOrder(price sdk.Dec, amt sdk.Int) *types.BaseOrder {
	return types.NewBaseOrder(types.SwapDirectionSell, price, amt, amt)
}

func newBuyUserOrder(reqId uint64, price sdk.Dec, amt sdk.Int) *types.UserOrder {
	return &types.UserOrder{
		BaseOrder: types.BaseOrder{
			Direction:                types.SwapDirectionBuy,
			Price:                    price,
			Amount:                   amt,
			OpenAmount:               amt,
			OfferCoinAmount:          price.MulInt(amt).Ceil().TruncateInt(),
			RemainingOfferCoinAmount: price.MulInt(amt).Ceil().TruncateInt(),
			ReceivedAmount:           sdk.ZeroInt(),
		},
		RequestId: reqId,
		Orderer:   testAddr,
	}
}

//nolint
func newSellUserOrder(reqId uint64, price sdk.Dec, amt sdk.Int) *types.UserOrder {
	return &types.UserOrder{
		BaseOrder: types.BaseOrder{
			Direction:                types.SwapDirectionSell,
			Price:                    price,
			Amount:                   amt,
			OpenAmount:               amt,
			OfferCoinAmount:          amt,
			RemainingOfferCoinAmount: amt,
			ReceivedAmount:           sdk.ZeroInt(),
		},
		RequestId: reqId,
		Orderer:   testAddr,
	}
}

func newBuyPoolOrder(poolId uint64, price sdk.Dec, amt sdk.Int) *types.PoolOrder {
	return &types.PoolOrder{
		BaseOrder: types.BaseOrder{
			Direction:                types.SwapDirectionBuy,
			Price:                    price,
			Amount:                   amt,
			OpenAmount:               amt,
			OfferCoinAmount:          price.MulInt(amt).Ceil().TruncateInt(),
			RemainingOfferCoinAmount: price.MulInt(amt).Ceil().TruncateInt(),
			ReceivedAmount:           sdk.ZeroInt(),
		},
		PoolId:         poolId,
		ReserveAddress: testAddr,
	}
}

//nolint
func newSellPoolOrder(poolId uint64, price sdk.Dec, amt sdk.Int) *types.PoolOrder {
	return &types.PoolOrder{
		BaseOrder: types.BaseOrder{
			Direction:                types.SwapDirectionSell,
			Price:                    price,
			Amount:                   amt,
			OpenAmount:               amt,
			OfferCoinAmount:          amt,
			RemainingOfferCoinAmount: amt,
			ReceivedAmount:           sdk.ZeroInt(),
		},
		PoolId:         poolId,
		ReserveAddress: testAddr,
	}
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
