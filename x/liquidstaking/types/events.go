package types

// Event types for the liquidstaking module.
const (
	EventTypeMsgLiquidStake             = TypeMsgLiquidStake
	EventTypeMsgLiquidUnstake           = TypeMsgLiquidUnstake
	EventTypeAddLiquidValidator         = "add_liquid_validator"
	EventTypeRemoveLiquidValidator      = "remove_liquid_validator"
	EventTypeBeginRebalancing           = "begin_rebalancing"
	EventTypeReStake                    = "re_stake"
	EventTypeUnbondInactiveLiquidTokens = "unbond_inactive_liquid_tokens"

	AttributeKeyDelegator             = "delegator"
	AttributeKeyNewShares             = "new_shares"
	AttributeKeyBTokenMintedAmount    = "btoken_minted_amount"
	AttributeKeyCompletionTime        = "completion_time"
	AttributeKeyUnbondingAmount       = "unbonding_amount"
	AttributeKeyUnbondedAmount        = "unbonded_amount"
	AttributeKeyLiquidValidator       = "liquid_validator"
	AttributeKeyRedelegationCount     = "redelegation_count"
	AttributeKeyRedelegationFailCount = "redelegation_fail_count"

	AttributeValueCategory = ModuleName
)
