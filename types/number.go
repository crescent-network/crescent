package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/cremath"
)

var (
	ZeroInt    = sdk.ZeroInt()
	ZeroDec    = sdk.ZeroDec()
	OneInt     = sdk.OneInt()
	OneDec     = sdk.OneDec()
	ZeroBigDec = cremath.ZeroBigDec()
	OneBigDec  = cremath.OneBigDec()
)
