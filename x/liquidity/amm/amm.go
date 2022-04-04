package amm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// MaxCoinAmount is the hard cap of coin amount used in the amm package.
	MaxCoinAmount = sdk.NewIntWithDecimal(1, 40)
)
