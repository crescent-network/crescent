package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// farming module sentinel errors
var (
	ErrPlanNotExists                  = sdkerrors.Register(ModuleName, 2, "plan does not exist")
	ErrInvalidPlanType                = sdkerrors.Register(ModuleName, 3, "invalid plan type")
	ErrInvalidPlanEndTime             = sdkerrors.Register(ModuleName, 4, "invalid plan end time")
	ErrStakingNotExists               = sdkerrors.Register(ModuleName, 5, "staking not exists")
	ErrRewardNotExists                = sdkerrors.Register(ModuleName, 6, "reward not exists")
	ErrFeeCollectionFailure           = sdkerrors.Register(ModuleName, 7, "fee collection failure")
	ErrInvalidPlanNameLength          = sdkerrors.Register(ModuleName, 8, "invalid plan name length")
	ErrDuplicatePlanName              = sdkerrors.Register(ModuleName, 9, "duplicate plan name")
	ErrInvalidPlanName                = sdkerrors.Register(ModuleName, 10, "invalid plan name")
	ErrConflictPrivatePlanFarmingPool = sdkerrors.Register(ModuleName, 11, "the address is already in use, please use a different plan name")
	ErrInvalidStakingReservedAmount   = sdkerrors.Register(ModuleName, 12, "staking reserved amount invariant broken")
	ErrInvalidRemainingRewardsAmount  = sdkerrors.Register(ModuleName, 13, "remaining rewards amount invariant broken")
)
