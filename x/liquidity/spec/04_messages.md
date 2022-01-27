<!-- order: 4 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps `liquidity` module messages from transactions.


## MsgCreatePair

```go
type MsgCreatePair struct {
	Creator        string // the bech32-encoded address of the pair creator
	BaseCoinDenom  string // the base coin denom of the pair
	QuoteCoinDenom string // the quote coin denom of the pair
}
```

## MsgCreatePool

```go
type MsgCreatePool struct {
	Creator      string    // the bech32-encoded address of the pool creator
	PairId       uint64    // the pair id; pool(s) belong to a single pair
	DepositCoins sdk.Coins // the amount of coins to deposit
}
```

## MsgDepositBatch

```go
type MsgDepositBatch struct {
	Depositor    string    // the bech32-encoded address that makes a deposit to the pool
	PoolId       uint64    // the pool id
	DepositCoins sdk.Coins // the amount of coins to deposit
}
```

## MsgWithdrawBatch

```go
type MsgWithdrawBatch struct {
	Withdrawer string   // the bech32-encoded address that withdraws pool coin from the pool
	PoolId     uint64   // the pool id
	PoolCoin   sdk.Coin // the amount of pool coin
}
```

## MsgLimitOrderBatch

```go
type MsgLimitOrderBatch struct {
	Orderer         string        // the bech32-encoded address that makes an order
	PairId          uint64        // the pair id
	Direction       SwapDirection // the swap direction; buy or sell
	OfferCoin       sdk.Coin      // the amount of coin that the orderer offers
	DemandCoinDenom string        // the demand coin denom that the orderer wants to swap for
	Price           sdk.Dec       // the order price; the exchange ratio is the amount of quote coin over the amount of base coin
	Amount          sdk.Int       // the amount of base coin that the orderer wants to buy or sell
	OrderLifespan   time.Duration // the order lifespan
}
```

## MsgMarketOrderBatch

```go
type MsgLimitOrderBatch struct {
	Orderer         string        // the bech32-encoded address that makes an order
	PairId          uint64        // the pair id
	Direction       SwapDirection // the swap direction; buy or sell
	OfferCoin       sdk.Coin      // the amount of coin that the orderer offers
	DemandCoinDenom string        // the demand coin denom that the orderer wants to swap for
	Amount          sdk.Int       // the amount of base coin that the orderer wants to buy or sell
	OrderLifespan   time.Duration // the order lifespan
}
```

## MsgCancelOrder

```go
type MsgCancelOrder struct {
	Orderer       string // the bech32-encoded address that makes an order
	PairId        uint64 // the pair id
	SwapRequestId uint64 // the swap request id
}
```

## MsgCancelAllOrders

```go
type MsgCancelAllOrders struct {
	Orderer string   // the bech32-encoded address that makes an order
	PairIds []uint64 // the pair ids
}
```

