package types

// Event types for the liquidstaking module.
const (
	EventTypeMsgLiquidStake   = TypeMsgLiquidStake
	EventTypeMsgLiquidUnstake = TypeMsgLiquidUnstake

	AttributeKeyNewShares          = "new_shares"
	AttributeKeyBTokenMintedAmount = "btoken_minted_amount"
	AttributeKeyCompletionTime     = "completion_time"
	AttributeKeyUnbondingAmount    = "unbonding_amount"

	AttributeValueCategory = ModuleName
)
