package types

// Event types for the liquidstaking module.
const (
	EventTypeMsgLiquidStake   = TypeMsgLiquidStake
	EventTypeMsgLiquidUnstake = TypeMsgLiquidUnstake

	AttributeKeyDelegator          = "delegator"
	AttributeKeyNewShares          = "new_shares"
	AttributeKeyBTokenMintedAmount = "btoken_minted_amount"
	AttributeKeyCompletionTime     = "completion_time"
	AttributeKeyUnbondingAmount    = "unbonding_amount"
	AttributeKeyUnbondedAmount     = "unbonded_amount"

	AttributeValueCategory = ModuleName
)
