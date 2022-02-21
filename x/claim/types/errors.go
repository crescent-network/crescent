package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/claim module sentinel errors
var (
	ErrAlreadyClaimed    = sdkerrors.Register(ModuleName, 2, "already claimed condition")
	ErrTerminatedAirdrop = sdkerrors.Register(ModuleName, 3, "terminated airdrop event")
	ErrConditionRequired = sdkerrors.Register(ModuleName, 4, "condition must be executed first")
)
