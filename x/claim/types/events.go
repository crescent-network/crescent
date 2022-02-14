package types

// Event types for the claim module.
const (
	EventTypeClaim = "claim"

	AttributeKeyRecipient             = "recipient"
	AttributeKeyInitialClaimableCoins = "initial_claimable_coins"
	AttributeKeyClaimableCoins        = "claimable_coins"
	AttributeKeyDepositActionClaimed  = "deposit_action_claimed"
	AttributeKeySwapActionClaimed     = "swap_action_claimed"
	AttributeKeyFarmingActionClaimed  = "farming_action_claimed"
	AttributeKeyUnclaimedCoins        = "unclaimed_coins"
)
