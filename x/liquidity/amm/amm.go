package amm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// The minimum and maximum coin amount used in the amm package.
var (
	MinCoinAmount             = sdk.NewInt(100)
	MaxCoinAmount             = sdk.NewIntWithDecimal(1, 40)
	MinPoolOrderPriceGapRatio = sdk.NewDecWithPrec(5, 4) // 5bp(0.05%)
	MaxPoolOrderPriceGapRatio = sdk.NewDecWithPrec(5, 3) // 50bp(0.5%)
)
