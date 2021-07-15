package types

// Event types for the farming module.
const (
	EventTypeCreateFixedAmountPlan = "create_fixed_amount_plan"
	EventTypeCreateRatioPlan       = "create_ratio_plan"
	EventTypeStake                 = "stake"
	EventTypeUnstake               = "unstake"
	EventTypeClaim                 = "claim"

	AttributeKeyPlanId                = "plan_id" //nolint:golint
	AttributeKeyFarmingPoolAddress    = "farming_pool_address"
	AttributeKeyStakingReserveAddress = "staking_reserve_address"
	AttributeKeyRewardPoolAddress     = "reward_pool_address"
	AttributeKeyTerminationAddress    = "termination_address"
	AttributeKeyStakingCoinWeights    = "staking_coin_weights"
	AttributeKeyStakingCoins          = "staking_coins"
	AttributeKeyUnstakingCoins        = "unstaking_coins"
	AttributeKeyRewardCoins           = "reward_coins"
	AttributeKeyStartTime             = "start_time"
	AttributeKeyEndTime               = "end_time"
	AttributeKeyEpochDays             = "epoch_days"
	AttributeKeyEpochAmount           = "epoch_amount"
	AttributeKeyEpochRatio            = "epoch_ratio"
)
