package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrPlanAlreadyTerminated = sdkerrors.Register(ModuleName, 2, "plan is already terminated")
)
