package types

// Event types for the liquidity module.
const (
	EventTypeCreatePool      = "create_pool"
	EventTypeDepositBatch    = "deposit_batch"
	EventTypeWithdrawBatch   = "withdraw_batch"
	EventTypeSwapBatch       = "swap_batch"
	EventTypeCancelSwapBatch = "cancel_swap_batch"

	AttributeKeyCreator         = "creator"
	AttributeKeyDepositor       = "depositor"
	AttributeKeyWithdrawer      = "withdrawer"
	AttributeKeyOrderer         = "orderer"
	AttributeKeyXCoin           = "x_coin"
	AttributeKeyYCoin           = "y_coin"
	AttributeKeyMintedPoolCoin  = "minted_pool_coin"
	AttributeKeyPoolCoin        = "pool_coin"
	AttributeKeyRequestId       = "request_id"
	AttributeKeyPoolId          = "pool_id"
	AttributeKeyPairId          = "pair_id"
	AttributeKeyBatchId         = "batch_id"
	AttributeKeySwapRequestId   = "swap_request_id"
	AttributeKeySwapDirection   = "swap_direction"
	AttributeKeyRemainingAmount = "remaining_amount"
	AttributeKeyReceivedAmount  = "received_amount"
)
