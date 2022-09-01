<!-- order: 4 -->

# Messages

Messages (Msg) are objects that trigger state transitions.
Msgs are wrapped in transactions (Txs) that clients submit to the network.
The Cosmos SDK wraps and unwraps `liquidity` module messages from transactions.

## MsgCreatePair

A coin pair is created with `MsgCreatePair` message.

```go
type MsgCreatePair struct {
    Creator        string // the bech32-encoded address of the pair creator
    BaseCoinDenom  string // the base coin denom of the pair
    QuoteCoinDenom string // the quote coin denom of the pair
}
```

### Validity Checks

Validity checks are performed for `MsgCreatePair` messages.
The transaction that is triggered with `MsgCreatePair` fails if:
- `Creator` address is invalid
- The coin pair already exists
- The balance of `Creator` does not have enough coins for `PairCreationFee`

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

Validity checks are performed for `MsgCreatePool` messages.
The transaction that is triggered with `MsgCreatePool` fails if:
- `Creator` address is invalid
- Pair with `PairId` does not exist
- Coin denoms from `DepositCoins` aren't equal to coin pair with `PairID`
- Amount of one of `DepositCoins` is less than `MinInitialDepositAmount`
- Active(not disabled) basic pool with same pair already exists
- The balance of `Creator` does not have enough amount of coins for `DepositCoins`
- The balance of `Creator` does not have enough coins for `PoolCreationFee`

## MsgCreateRangedPool

A ranged liquidity pool is created and initial coins are deposited with the `MsgCreateRangedPool` message.

```go
type MsgCreateRangedPool struct {
    Creator      string    // the bech32-encoded address of the pool creator
    PairId       uint64    // the pair id; pool(s) belong to a single pair
    DepositCoins sdk.Coins // the amount of coins to deposit
    MinPrice     sdk.Dec   // the minimum price of the ranged pool
    MaxPrice     sdk.Dec   // the maximum price of the ranged pool
    InitialPrice sdk.Dec   // the initial pool price
}
```

Read more about ranged pool creation in the [Liquidity pool white paper](../../../docs/whitepapers/liquidity/pool.md#creation-of-ranged-liquidity-pool).

### Validity Checks

Validity checks are performed for `MsgCreateRangedPool` messages.
The transaction that is triggered with `MsgCreateRangedPool` fails if:
- `Creator` address is invalid
- Pair with `PairId` does not exist
- Coin denoms from `DepositCoins` aren't equal to coin pair with `PairID`
- Amount of one of `DepositCoins` is less than `MinInitialDepositAmount`
- The balance of `Creator` does not have enough amount of coins for `DepositCoins`
- The balance of `Creator` does not have enough coins for `PoolCreationFee`
- Relationship among `InitialPrice`, `MinPrice` and `MaxPrice` is invalid.

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

Validity checks are performed for `MsgDeposit` messages.
The transaction that is triggered with the `MsgDeposit` message fails if:
- `Depositor` address is invalid
- Pool with `PoolId` does not exist
- The pool with `PoolId` is disabled
- The denoms of `DepositCoins` are different from the pair of the pool specified by `PoolId`
- The balance of `Depositor` does not have enough coins for `DepositCoins`

Read more about deposit and withdraw in the [Liquidity pool white paper](../../../docs/whitepapers/liquidity/pool.md#deposit-and-withdraw-ratio).

## MsgWithdraw

Withdraw coins in batch from liquidity pool with the `MsgWithdraw` message.

```go
type MsgWithdraw struct {
    Withdrawer string   // the bech32-encoded address that withdraws pool coin from the pool
    PoolId     uint64   // the pool id
    PoolCoin   sdk.Coin // the amount of pool coin
}
```

Read more about deposit and withdraw in the [Liquidity pool white paper](../../../docs/whitepapers/liquidity/pool.md#deposit-and-withdraw-ratio).

### Validity Checks

Validity checks are performed for `MsgWithdraw` messages.
The transaction that is triggered with the `MsgWithdraw` message fails if:
- `Withdrawer` address is invalid
- Pool with `PoolId` does not exist
- The pool with `PoolId` is disabled
- The denom of `PoolCoin` isn't equal to pool coin denom with `PoolId`
- The balance of `Withdrawer` does not have enough coins for `PoolCoin`

## MsgLimitOrder

Swap coins through limit order with `MsgLimitOrder` message.

```go
type MsgLimitOrder struct {
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

`Price` that isn't fit on price ticks is automatically converted by this rule:

- For buy orders, the resulting price will be the highest tick price lower than(or equal to) `Price`
- For sell orders, the resulting price will be the lowest tick price higher than(or equal to) `Price`

Note that an order will be executed for at least one batch, even if `OrderLifespan` is specified as `0`.

### Validity Checks

Validity checks are performed for `MsgLimitOrder` messages.
The transaction that is triggered with the `MsgLimitOrder` message fails if:
- `Orderer` address is invalid
- Pair with `PairId` does not exist
- `OrderLifespan` is greater than `MaxOrderLifespan`
- `Direction` is invalid
- Denom of `OfferCoin` or `DemandCoinDenom` doesn't match with the pair specified `PairId`
- Denom of `OfferCoin` and `DemandCoinDenom` are not entered properly according to the `Direction`
- `Price` is not in the range of (1-`MaxPriceLimitRatio`)*`LastPrice` to (1+`MaxPriceLimitRatio`)*`LastPrice`
- The balance of `Orderer` does not have enough coins for `OfferCoin`

## MsgMarketOrder

Swap coins through market order with `MsgMarketOrder` message.

```go
type MsgMarketOrder struct {
    Orderer         string        // the bech32-encoded address that makes an order
    PairId          uint64        // the pair id
    Direction       SwapDirection // the swap direction; buy or sell
    OfferCoin       sdk.Coin      // the amount of coin that the orderer offers
    DemandCoinDenom string        // the demand coin denom that the orderer wants to swap for
    Amount          sdk.Int       // the amount of base coin that the orderer wants to buy or sell
    OrderLifespan   time.Duration // the order lifespan
}
```

Market orders can only be made when the pair has last price.
Market orders are converted to limit orders by following rule:

- Buy market orders are converted to limit orders with price of `LastPrice * (1+MaxPriceLimitRatio)`
- Sell market orders are converted to limit orders with price of `LastPrice * (1-MaxPriceLimitRatio)`

After the conversion, market orders are treated same as limit orders.

Note that an order will be executed for at least one batch, even if `OrderLifespan` is specified as `0`.

### Validity Checks

Validity checks are performed for `MsgMarketOrder` messages.
The transaction that is triggered with the `MsgMarketOrder` message fails if:
- `Orderer` address is invalid
- Pair with `PairId` does not exist
- `OrderLifespan` is greater than `MaxOrderLifespan`
- `Direction` is invalid
- Denom of `OfferCoin` or `DemandCoinDenom` doesn't match with the pair specified `PairId`
- Denom of `OfferCoin` and `DemandCoinDenom` are not entered properly according to the `Direction`
- The balance of `Orderer` does not have enough coins for `OfferCoin`

## MsgMMOrder

Make an MM(market making) order, which places multiple limit orders at once based
on the parameters.

```go
type MsgMMOrder struct {
    Orderer       string
    PairId        uint64
    MaxSellPrice  sdk.Dec
    MinSellPrice  sdk.Dec
    SellAmount    sdk.Int
    MaxBuyPrice   sdk.Dec
    MinBuyPrice   sdk.Dec
    BuyAmount     sdk.Int
    OrderLifespan time.Duration
}
```

Limit orders are created at even intervals, for each buy/sell side.
If the amount is zero, then no orders are made for that order direction.
The maximum number of orders for each side is limited by the `MaxNumMarketMakingOrderTicks`
parameter.
At any point, there can be only one MM order from an orderer.
If the orderer makes another MM order, then the previous order will be canceled.

## MsgCancelOrder

Cancel an order with `MsgCancelOrder` message.

```go
type MsgCancelOrder struct {
    Orderer string // the bech32-encoded address that makes an order
    PairId  uint64 // the pair id
    OrderId uint64 // the order id
}
```

Orders are executed for at least one batch.
That means, users cannot cancel orders that has just been made.

### Validity Checks

Validity checks are performed for `MsgCancelOrder` messages.
The transaction that is triggered with the `MsgCancelOrder` message fails if:
- `Orderer` address is invalid
- Pair with `PairId` does not exist
- Order with `OrderId` does not exist in pair with `PairId`
- `Orderer` is not the orderer from order with `OrderId`
- Order with `OrderId` is already canceled

## MsgCancelAllOrders

Cancel all orders with `MsgCancelAllOrders` message.

```go
type MsgCancelAllOrders struct {
    Orderer string   // the bech32-encoded address that makes an order
    PairIds []uint64 // the pair ids
}
```

`MsgCancelAllOrders` cancels only orders that can be canceled with `MsgCancelOrder`.

### Validity Checks

Validity checks are performed for `MsgCancelAllOrders` messages.
The transaction that is triggered with the `MsgCancelAllOrders` message fails if:
- `Orderer` address is invalid
- Pair with `PairId` in `PairIds` does not exist

## MsgCancelMMOrder

Cancel an MM(market making) order.

```go
type MsgCancelMMOrder struct {
    Orderer string
    PairId  uint64
}
```

Cancel previously made MM order by specifying the pair id.
