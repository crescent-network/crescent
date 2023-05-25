package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInsufficientOutput   = sdkerrors.Register(ModuleName, 2, "insufficient output amount")
	ErrOrderPriceOutOfRange = sdkerrors.Register(ModuleName, 3, "order price out of range")
)
