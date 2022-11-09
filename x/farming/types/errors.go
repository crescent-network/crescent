package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// farming module sentinel errors
var (
	ErrInvalidPlanType                 = sdkerrors.Register(ModuleName, 2, "invalid plan type")
	ErrInvalidPlanName                 = sdkerrors.Register(ModuleName, 3, "invalid plan name")
	ErrInvalidPlanEndTime              = sdkerrors.Register(ModuleName, 4, "invalid plan end time")
	ErrInvalidStakingCoinWeights       = sdkerrors.Register(ModuleName, 5, "invalid staking coin weights")
	ErrInvalidTotalEpochRatio          = sdkerrors.Register(ModuleName, 6, "invalid total epoch ratio")
	ErrStakingNotExists                = sdkerrors.Register(ModuleName, 7, "staking not exists")
	ErrConflictPrivatePlanFarmingPool  = sdkerrors.Register(ModuleName, 8, "the address is already in use, please use a different plan name")
	ErrInvalidStakingReservedAmount    = sdkerrors.Register(ModuleName, 9, "staking reserved amount invariant broken")
	ErrInvalidRemainingRewardsAmount   = sdkerrors.Register(ModuleName, 10, "remaining rewards amount invariant broken")
	ErrInvalidOutstandingRewardsAmount = sdkerrors.Register(ModuleName, 11, "outstanding rewards amount invariant broken")
	ErrNumPrivatePlansLimit            = sdkerrors.Register(ModuleName, 12, "cannot create private plans more than the limit")
	ErrNumMaxDenomsLimit               = sdkerrors.Register(ModuleName, 13, "number of denoms cannot exceed the limit")
	ErrInvalidEpochAmount              = sdkerrors.Register(ModuleName, 14, "invalid epoch amount")
	ErrRatioPlanDisabled               = sdkerrors.Register(ModuleName, 15, "creation of ratio plans is disabled")
	ErrInvalidUnharvestedRewardsAmount = sdkerrors.Register(ModuleName, 16, "invalid unharvested rewards amount")
	ErrModuleDisabled                  = sdkerrors.Register(ModuleName, 17, "farming module has been disabled")
)
