package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/claim module sentinel errors
var (
	ErrAlreadyClaimedAll = sdkerrors.Register(ModuleName, 2, "already claimed all actions")
	ErrAlreadyClaimed    = sdkerrors.Register(ModuleName, 3, "already claimed action")
	ErrTerminatedAirdrop = sdkerrors.Register(ModuleName, 4, "terminated airdrop event")
)
