package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

// x/liquidamm module sentinel errors
var (
	ErrPublicPositionExists = sdkerrors.Register(ModuleName, 2, "public position with same parameters already exists")
	ErrInsufficientBidAmount = sdkerrors.Register(ModuleName, 3, "insufficient bid amount")
)
