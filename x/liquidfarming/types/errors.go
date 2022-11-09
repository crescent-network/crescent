package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/liquidfarming module sentinel errors
var (
	ErrSmallerThanMinimumAmount      = sdkerrors.Register(ModuleName, 2, "smaller than the minimum amount")
	ErrNotBiggerThanWinningBidAmount = sdkerrors.Register(ModuleName, 3, "not bigger than the winning bid amount")
)
