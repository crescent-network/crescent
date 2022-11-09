package types

// Event types for the marketmaker module.
const (
	EventTypeApplyMarketMaker     = "apply_market_maker"
	EventTypeClaimIncentives      = "claim_incentives"
	EventTypeIncludeMarketMaker   = "include_market_maker"
	EventTypeExcludeMarketMaker   = "exclude_market_maker"
	EventTypeRejectMarketMaker    = "reject_market_maker"
	EventTypeDistributeIncentives = "distribute_incentives"

	AttributeKeyAddress         = "address"
	AttributeKeyPairIds         = "pair_ids"
	AttributeKeyPairId          = "pair_id"
	AttributeKeyBudgetAddress   = "budget_address"
	AttributeKeyTotalIncentives = "total_incentives"

	AttributeValueCategory = ModuleName
)
