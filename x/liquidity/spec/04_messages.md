<!-- order: 4 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps `liquidity` module messages from transactions.


## MsgCreatePair
A coin pair is created with `MsgCfeatePair` message.
```go
type MsgCreatePair struct {
    Creator        string // the bech32-encoded address of the pair creator
    BaseCoinDenom  string // the base coin denom of the pair
    QuoteCoinDenom string // the quote coin denom of the pair
}
```
### Validity Checks
Validity checks are performed for MsgCreatePair messages. The transaction that is triggered with MsgCreatePair fails if:
- `Creator` address does not exist
- The coin pair already exists
- The balance of `Creator` does not have enough coins for PairCreationFee

## MsgCreatePool
A liquidity pool is created and initial coins are deposited with the `MsgCreatePool` message.
```go
type MsgCreatePool struct {
    Creator      string    // the bech32-encoded address of the pool creator
    PairId       uint64    // the pair id; pool(s) belong to a single pair
    DepositCoins sdk.Coins // the amount of coins to deposit
}
```
### Validity Checks
Validity checks are performed for MsgCreatePool messages. The transaction that is triggered with MsgCreatePool fails if:
- `Creator` address does not exist
- `PairId` does not exist in parameters
- Coin denoms from `DepositCoins` does not equal to coin pair with `PairID`
- Amount of one of `DepositCoins` is less than `MinInitialDepositAmount`
- Valid `Pool` with same pair already exists
- One or more coins `DepositCoins` do not exist in `bank` module
- The balance of `Creator` does not have enough amount of coins for DepositCoins
- The balance of `Creator` does not have enough coins for `PoolCreationFee`


## MsgDeposit
Coins are deposited in a batch to a liquidity pool with the `MsgDeposit` message.
```go
type MsgDeposit struct {
    Depositor    string    // the bech32-encoded address that makes a deposit to the pool
    PoolId       uint64    // the pool id
    DepositCoins sdk.Coins // the amount of coins to deposit
}
```
### Validity Checks
Validity checks are performed for MsgDeposit messages. The transaction that is triggered with the MsgDeposit message fails if:
- `Depositor` address does not exist
- `PoolId` does not exist
- The pool with `PoolId` is disabled
- The denoms of `DepositCoins` are not composed of existing coin pair of the specified `Pool`
- The balance of `Depositor` does not have enough coins for `DepositCoins`

## MsgWithdraw
Withdraw coins in batch from liquidity pool with the `MsgWithdraw` message.
```go
type MsgWithdraw struct {
    Withdrawer string   // the bech32-encoded address that withdraws pool coin from the pool
    PoolId     uint64   // the pool id
    PoolCoin   sdk.Coin // the amount of pool coin
}
```
### Validity Checks
Validity checks are performed for MsgWithdraw messages. The transaction that is triggered with the MsgWithdraw message fails if:
- `Withdrawer` address does not exist
- `PoolId` does not exist
- The pool with `PoolId` is disabled
- The denom of `PoolCoin` not equal to pool coin denom with `PoolId`
- The balance of `Withdrawer` does not have enough coins for `PoolCoin`

## MsgLimitOrder
Swap coins through limit order with `MsgLimitOrder` message.
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
### Validity Checks
Validity checks are performed for MsgLimitOrder messages. The transaction that is triggered with the MsgLimitOrder message fails if:
- `Orderer` address does not exist
- The pair with `PairId` does not exist
- `Price` is not tick price
- `OrderLifespan` is greater than `MaxOrderLifespan`
- `Direction` is invalid
- Denom of `OfferCoin` or `DemandCoinDenom` do not exist in `bank` module
- Denom of `OfferCoin` and `DemandCoinDenom` are not entered properly according to the `Direction`
- `Price` is not in the range of (1-`MaxPriceLimitRatio`)*`LastPrice` to (1+`MaxPriceLimitRatio`)*`LastPrice`
- The balance of `Orderer` does not have enough coins for `OfferCoin`

## MsgMarketOrder
Swap coins through market order with `MsgMarketOrder` message.
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
### Validity Checks
Validity checks are performed for MsgMarketOrder messages. The transaction that is triggered with the MsgMarketOrder message fails if:
- `Orderer` address does not exist
- The pair with `PairId` does not exist
- `Price` is not tick price
- `OrderLifespan` is greater than `MaxOrderLifespan`
- `Direction` is invalid
- Denom of `OfferCoin` or `DemandCoinDenom` do not exist in `bank` module
- Denom of `OfferCoin` and `DemandCoinDenom` are not entered properly according to the `Direction`
- The balance of `Orderer` does not have enough coins for `OfferCoin`


## MsgCancelOrder
Cancel a swap order with `MsgCancelOrder` message.
```go
type MsgCancelOrder struct {
    Orderer       string  // the bech32-encoded address that makes an order
    PairId        uint64  // the pair id
    SwapRequestId uint64  // the swap request id
}
```
### Validity Checks
Validity checks are performed for MsgCancelOrder messages. The transaction that is triggered with the MsgCancelOrder message fails if:
- `Orderer` address does not exist
- The pair with `PairId` does not exist
- `SwapRequestId` does not exist in pair with `PairId`
- `Orderer` is not the orderer from swap request with `SwapRequestId`
- Swap request with `SwapRequestId` is already canceled

## MsgCancelAllOrders
Cancel all swap order with `MsgCancelAllOrder` message.
```go
type MsgCancelAllOrders struct {
    Orderer string   // the bech32-encoded address that makes an order
    PairIds []uint64 // the pair ids
}
```
### Validity Checks
Validity checks are performed for MsgCancelOrder messages. The transaction that is triggered with the MsgCancelOrder message fails if:
- `Orderer` address does not exist
- `PairId` in `PairIds` does not exist
- There is no swap request in one of pair with `PairId` in `PairIds`
- There is no swap request is not already canceled in one of pair with `PairId` in `PairIds`
- There is no swap request with `BatchId` lower than pair's `CurrentBatchId` in one of pair with `PairId` in `PairIds`

