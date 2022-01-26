package types

// Event types for the liquidity module.
const (
	EventTypeCreatePair      = "create_pair"
	EventTypeCreatePool      = "create_pool"
	EventTypeDepositBatch    = "deposit_batch"
	EventTypeWithdrawBatch   = "withdraw_batch"
	EventTypeSwapBatch       = "swap_batch"
	EventTypeCancelSwapBatch = "cancel_swap_batch"

	AttributeKeyCreator         = "creator"
	AttributeKeyDepositor       = "depositor"
	AttributeKeyWithdrawer      = "withdrawer"
	AttributeKeyOrderer         = "orderer"
	AttributeKeyBaseCoinDenom   = "base_coin_denom"
	AttributeKeyQuoteCoinDenom  = "quote_coin_denom"
	AttributeKeyDepositCoins    = "deposit_coins"
	AttributeKeyMintedPoolCoin  = "minted_pool_coin"
	AttributeKeyPoolCoin        = "pool_coin"
	AttributeKeyReserveAddress  = "reserve_address"
	AttributeKeyEscrowAddress   = "escrow_address"
	AttributeKeyRequestId       = "request_id"
	AttributeKeyPoolId          = "pool_id"
	AttributeKeyPairId          = "pair_id"
	AttributeKeyBatchId         = "batch_id"
	AttributeKeySwapRequestId   = "swap_request_id"
	AttributeKeySwapDirection   = "swap_direction"
	AttributeKeyOfferCoin       = "offer_coin"
	AttributeKeyDemandCoinDenom = "demand_coin_denom"
	AttributeKeyPrice           = "price"
	AttributeKeyAmount          = "amount"
	AttributeKeyExpireAt        = "expire_at"
)
