package bindings

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
)

// CrescentMsg contains what messages that can be called from a contract
type CrescentMsg struct {
	LimitOrder *LimitOrder `json:"limit_order,omitempty"`
}

// LimitOrder constructs a message to make a limit order in the liquidity module
type LimitOrder struct {
	// orderer specifies the bech32-encoded address that makes an order
	Orderer string `json:"orderer"`
	// pair_id specifies the pair id
	PairId uint64 `json:"pair_id"`
	// direction specifies the order direction (buy or sell)
	Direction liquiditytypes.OrderDirection `json:"direction"`
	// offer_coin specifies the amount of coin the orderer offers
	OfferCoin sdk.Coin `json:"offer_coin"`
	// demand_coin_denom specifies the demand coin denom
	DemandCoinDenom string `json:"demand_coin_denom"`
	// price specifies the order price
	Price sdk.Dec `json:"price"`
	// amount specifies the amount of base coin the orderer wants to buy or sell
	Amount sdk.Int `json:"amount"`
	// order_lifespan specifies the order lifespan
	OrderLifespan time.Duration `json:"order_lifespan"`
}
