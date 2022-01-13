package types

// Event types for the liquidity module.
const (
	EventTypeCreatePool = "create_pool"

	AttributeKeyCreator         = "creator"
	AttributeKeyDepositor       = "depositor"
	AttributeKeyWithdrawer      = "withdrawer"
	AttributeKeyOrderer         = "orderer"
	AttributeKeyXCoin           = "x_coin"
	AttributeKeyYCoin           = "y_coin"
	AttributeKeyMintedPoolCoin  = "minted_pool_coin"
	AttributeKeyPoolCoin        = "pool_coin"
	AttributeKeyRequestId       = "request_id"
	AttributeKeyBatchId         = "batch_id"
	AttributeKeySwapDirection   = "swap_direction"
	AttributeKeyRemainingAmount = "remaining_amount"
	AttributeKeyReceivedAmount  = "received_amount"
)
