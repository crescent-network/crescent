package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

// x/liquidamm module sentinel errors
var (
	ErrInsufficientBidAmount = sdkerrors.Register(ModuleName, 2, "insufficient bid amount")
)
