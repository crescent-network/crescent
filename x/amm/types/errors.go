package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrAddZeroLiquidity = sdkerrors.Register(ModuleName, 2, "added liquidity is zero")
)
