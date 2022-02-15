package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/claim module sentinel errors
var (
	ErrTerminatedAirdrop = sdkerrors.Register(ModuleName, 2, "terminated airdrop event")
)
