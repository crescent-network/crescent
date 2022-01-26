package types

// Event types for the liquidstaking module.
const (
	EventTypeMsgLiquidStake   = TypeMsgLiquidStake
	EventTypeMsgLiquidUnstake = TypeMsgLiquidUnstake

	// TODO: check conventions AttributeValue or AttributeKey
	AttributeKeyNewShares          = "new_shares"
	AttributeKeyBTokenMintedAmount = "btoken_minted_amount"
	AttributeKeyCompletionTime     = "completion_time"
	AttributeKeyUnbondingAmount    = "unbonding_amount"
	//AttributeValueName               = "name"
	//AttributeValueDestinationAddress = "destination_address"
	//AttributeValueSourceAddress      = "source_address"
	//AttributeValueRate               = "rate"
	//AttributeValueAmount             = "amount"

	AttributeValueCategory = ModuleName
)
