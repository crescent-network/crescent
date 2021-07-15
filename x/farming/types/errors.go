package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// farming module sentinel errors
var (
	ErrPlanNotExists             = sdkerrors.Register(ModuleName, 2, "plan not exists")
	ErrPlanTypeNotExists         = sdkerrors.Register(ModuleName, 3, "plan type not exists")
	ErrInvalidPlanType           = sdkerrors.Register(ModuleName, 4, "invalid plan type")
	ErrInvalidPlanEndTime        = sdkerrors.Register(ModuleName, 5, "invalid plan end time")
	ErrInvalidPlanEpochDays      = sdkerrors.Register(ModuleName, 6, "invalid plan epoch days")
	ErrInvalidPlanEpochRatio     = sdkerrors.Register(ModuleName, 7, "invalid plan epoch ratio")
	ErrEmptyEpochAmount          = sdkerrors.Register(ModuleName, 8, "epoch amount must not be empty")
	ErrEmptyStakingCoinWeights   = sdkerrors.Register(ModuleName, 9, "staking coin weights must not be empty")
	ErrStakingNotExists          = sdkerrors.Register(ModuleName, 10, "staking not exists")
	ErrRewardNotExists           = sdkerrors.Register(ModuleName, 11, "reward not exists")
	ErrInsufficientStakingAmount = sdkerrors.Register(ModuleName, 12, "insufficient staking amount")
)
