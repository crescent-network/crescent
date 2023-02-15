package types

// Event types for the bootstrap module.
const (
	EventTypeCreatePool       = "create_pool"
	EventTypeLimitOrder       = "limit_order"
	EventTypeUserOrderMatched = "user_order_matched"
	EventTypePoolOrderMatched = "pool_order_matched"

	AttributeKeyProposer           = "proposer"
	AttributeKeyOrderer            = "orderer"
	AttributeKeyBaseCoinDenom      = "base_coin_denom"
	AttributeKeyQuoteCoinDenom     = "quote_coin_denom"
	AttributeKeyAcceptedCoins      = "accepted_coins"
	AttributeKeyRefundedCoins      = "refunded_coins"
	AttributeKeyReserveAddress     = "reserve_address"
	AttributeKeyEscrowAddress      = "escrow_address"
	AttributeKeyRequestId          = "request_id"
	AttributeKeyPoolId             = "pool_id"
	AttributeKeyStageId            = "stage_id"
	AttributeKeyOrderId            = "order_id"
	AttributeKeyOrderIds           = "order_ids"
	AttributeKeyOrderDirection     = "order_direction"
	AttributeKeyOfferCoin          = "offer_coin"
	AttributeKeyDemandCoinDenom    = "demand_coin_denom"
	AttributeKeyPrice              = "price"
	AttributeKeyAmount             = "amount"
	AttributeKeyOpenAmount         = "open_amount"
	AttributeKeyRemainingOfferCoin = "remaining_offer_coin"
	AttributeKeyReceivedCoin       = "received_coin"
	AttributeKeyPoolIds            = "pool_ids"
	AttributeKeyStatus             = "status"
	AttributeKeyMatchedAmount      = "matched_amount"
	AttributeKeyPaidCoin           = "paid_coin"

	AttributeValueCategory = ModuleName
)
