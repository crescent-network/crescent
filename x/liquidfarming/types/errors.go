package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/liquidfarming module sentinel errors
var (
	ErrSmallerThanMinimumAmount    = sdkerrors.Register(ModuleName, 2, "smaller than minimum amount")
	ErrSmallerThanWinningBidAmount = sdkerrors.Register(ModuleName, 3, "smaller than winning bid  amount")
)
