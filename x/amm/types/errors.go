package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrConditionsNotMet = sdkerrors.Register(ModuleName, 2, "conditions not met")
)
