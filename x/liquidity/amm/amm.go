package amm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// The minimum and maximum coin amount used in the amm package.
var (
	MinCoinAmount           = sdk.NewInt(100)
	MaxCoinAmount           = sdk.NewIntWithDecimal(1, 40)
	PoolOrderPriceDiffRatio = sdk.NewDecWithPrec(5, 4) // 5bp(0.05%)
)
