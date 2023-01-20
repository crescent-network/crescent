package types

// Event types for the bootstrap module.
const (
	EventTypeApplyBootstrap       = "apply_market_maker"
	EventTypeClaimIncentives      = "claim_incentives"
	EventTypeIncludeBootstrap     = "include_market_maker"
	EventTypeExcludeBootstrap     = "exclude_market_maker"
	EventTypeRejectBootstrap      = "reject_market_maker"
	EventTypeDistributeIncentives = "distribute_incentives"

	AttributeKeyAddress         = "address"
	AttributeKeyPairIds         = "pair_ids"
	AttributeKeyPairId          = "pair_id"
	AttributeKeyBudgetAddress   = "budget_address"
	AttributeKeyTotalIncentives = "total_incentives"

	AttributeValueCategory = ModuleName
)
