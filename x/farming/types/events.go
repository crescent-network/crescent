package types

// Event types for the farming module.
const (
	EventTypeCreateFixedAmountPlan = "create_fixed_amount_plan"
	EventTypeCreateRatioPlan       = "create_ratio_plan"
	EventTypeStake                 = "stake"
	EventTypeUnstake               = "unstake"
	EventTypeHarvest               = "harvest"
	EventTypeRemovePlan            = "remove_plan"
	EventTypeRewardsWithdrawn      = "rewards_withdrawn"
	EventTypePlanTerminated        = "plan_terminated"
	EventTypeRewardsAllocated      = "rewards_allocated"

	AttributeKeyPlanId             = "plan_id" //nolint:golint
	AttributeKeyPlanName           = "plan_name"
	AttributeKeyFarmingPoolAddress = "farming_pool_address"
	AttributeKeyTerminationAddress = "termination_address"
	AttributeKeyStakingCoins       = "staking_coins"
	AttributeKeyUnstakingCoins     = "unstaking_coins"
	AttributeKeyRewardCoins        = "reward_coins"
	AttributeKeyStartTime          = "start_time"
	AttributeKeyEndTime            = "end_time"
	AttributeKeyEpochAmount        = "epoch_amount"
	AttributeKeyEpochRatio         = "epoch_ratio"
	AttributeKeyFarmer             = "farmer"
	AttributeKeyAmount             = "amount"
	AttributeKeyStakingCoinDenom   = "staking_coin_denom"
	AttributeKeyStakingCoinDenoms  = "staking_coin_denoms"
)
