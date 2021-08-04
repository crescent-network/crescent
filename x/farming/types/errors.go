package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// farming module sentinel errors
var (
	ErrPlanNotExists        = sdkerrors.Register(ModuleName, 2, "plan does not exist")
	ErrInvalidPlanType      = sdkerrors.Register(ModuleName, 3, "invalid plan type")
	ErrInvalidPlanEndTime   = sdkerrors.Register(ModuleName, 4, "invalid plan end time")
	ErrStakingNotExists     = sdkerrors.Register(ModuleName, 5, "staking not exists")
	ErrRewardNotExists      = sdkerrors.Register(ModuleName, 6, "reward not exists")
	ErrFeeCollectionFailure = sdkerrors.Register(ModuleName, 7, "fee collection failure")
	ErrInvalidNameLength    = sdkerrors.Register(ModuleName, 8, "invalid name length")
	ErrDuplicatePlanName    = sdkerrors.Register(ModuleName, 9, "duplicate plan name")
)
