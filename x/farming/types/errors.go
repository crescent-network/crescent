package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// farming module sentinel errors
var (
	ErrPlanNotExists           = sdkerrors.Register(ModuleName, 2, "plan does not exist")
	ErrInvalidPlanType         = sdkerrors.Register(ModuleName, 3, "invalid plan type")
	ErrInvalidPlanEndTime      = sdkerrors.Register(ModuleName, 4, "invalid plan end time")
	ErrInvalidPlanEpochRatio   = sdkerrors.Register(ModuleName, 5, "invalid plan epoch ratio")
	ErrEmptyEpochAmount        = sdkerrors.Register(ModuleName, 6, "epoch amount must not be empty")
	ErrEmptyStakingCoinWeights = sdkerrors.Register(ModuleName, 7, "staking coin weights must not be empty")
	ErrStakingNotExists        = sdkerrors.Register(ModuleName, 8, "staking not exists")
	ErrRewardNotExists         = sdkerrors.Register(ModuleName, 9, "reward not exists")
	ErrFeeCollectionFailure    = sdkerrors.Register(ModuleName, 10, "fee collection failure")
)
