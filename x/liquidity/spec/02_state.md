<!-- order: 2 -->

# State

The `liquidity` module keeps track of the Pair, Pool, Requests states. 

## Pair

Pair stores information about the coin pair in the liquidity module.
A pair is the dyadic quotation of the relative value of a base coin unit against the unit of quote coin. 

Pair type has the following structure.

```go
type Pair struct {
    Id             uint64  // id of the coin pair
    BaseCoinDenom  string  // denom of the base coin for the pair
    QuoteCoinDenom string  // denom of the quote coin for the pair
    EscrowAddress  string  // address for the escrow account
    LastOrderId    uint64  // id of the last order for the pair
    LastPrice      sdk.Dec // the last swap price of the pair
    CurrentBatchId uint64  // id of the batch for pair
}
```

## Pool

Pool stores information about the liquidity pool. 

Pool type has the following structure.

```go
// PoolType enumerates pool types.
type PoolType int32

const (
    // POOL_TYPE_UNSPECIFIED specifies unknown pool type
    PoolTypeUnspecified PoolType = 0
    // POOL_TYPE_BASIC specifies the basic pool type
    PoolTypeBasic PoolType = 1
    // POOL_TYPE_RANGED specifies the ranged pool type
    PoolTypeRanged PoolType = 2
)

type Pool struct {
    Type                  PoolTye  // type of the pool
    Id                    uint64   // id of the liquidity pool
    PairId                uint64   // id of the coin pair constituting this pool
    Creator               string   // the pool creator address
    ReserveAddress        string   // reserve account address for the liquidity pool to store reserve coins
    PoolCoinDenom         string   // denom of the pool coin
    MinPrice              *sdk.Dec // the minimum price of ranged pool, nil for basic pools
    MaxPrice              *sdk.Dec // the maximum price of ranged pool, nil for basic pools
    LastDepositRequestId  uint64   // id of the last deposit request for the pool
    LastWithdrawRequestId uint64   // id of the last withdraw request for the pool
    Disabled              bool     // true if pool is disabled, false if not disabled
}
```

# Requests

Deposit, withdrawal, or swap orders are accumulated for a pre-defined period,
which can be one or more blocks in length.
Orders are then added to the pair or pool and executed at the end of the batch.
The following requests are executed in batch-style.

## RequestStatus

```go
type RequestStatus int32

const (
    RequestStatusUnspecified RequestStatus = iota + 1
    RequestStatusNotExecuted
    RequestStatusSucceeded
    RequestStatusFailed
)
```

## DepositRequest

`DeposiRequest` defines the state of deposit message as it is processed in the next batch or batches.

When a user sends `MsgDeposit` transaction to the network, it is accumulated in a batch.
`DeposiRequest` contains the information required for deposit transaction,
the result and the status of the request.

```go
type DepositRequest struct {
    Id             uint64      // id of the deposit message in the liquidity pool
    PoolId         uint64      // id of the pool where the deposit will occur
    MsgHeight      int64       // block height where this message is appended to the batch
    Depositor      string      // address that makes a deposit to the pool
    DepositCoins   sdk.Coins   // the amount of coins to deposit
    AcceptedCoins  sdk.Coins   // the amount of accepted coins to deposit
    MintedPoolCoin sdk.Coin    // the amount of minted pool coin for the amount of accepted coins
    Status         RequestStatus
}
```

## WithdrawRequest

`WithdrawRequest` defines the state of withdraw message as it is processed in the next batch or batches.

When a user sends `MsgWithdraw` transaction to the network, it is accumulated in a batch.
`WithdrawRequest` contains the information required for withdraw transaction,
the result and the status of the request.

```go
type WithdrawRequest struct {
    Id             uint64      // id of the withdraw message in the liquidity pool
    PoolId         uint64      // id of the pool where the withdraw will occur
    MsgHeight      int64       // block height where this message is appended to the batch
    Withdrawer     string      // address that withdraws pool coin from the pool
    PoolCoin       sdk.Coin    // the amount of pool coin to withdraw
    WithdrawnCoins sdk.Coin    // the amount of reserve coins for the amount of withdrawn pool coin
    Status         RequestStatus
}
```

## OrderStatus

```go
type OrderStatus int32

const (
    OrderStatusUnspecified OrderStatus = iota + 1
    OrderStatusNotExecuted
    OrderStatusNotMatched
    OrderStatusPartiallyMatched
    OrderStatusCompleted
    OrderStatusCanceled
    OrderStatusExpired
)
```

## Order

`Order` defines the state of swap message(`MsgLimitOrder`, `MsgMarketOrder`) as it is processed
in the next batch or batches.

When a user sends `MsgLimitOrder` or `MsgMarketOrder` transaction to the network,
it is accumulated in a batch.
`Order` contains the information required for swap transaction,
the result and the status of the request.

```go
type OrderDirection int32

const (
    OrderDirectionUnspecified  SwapDirection = iota + 1
    OrderDirectionBuy
    OrderDirectionSell
)

type Order struct {
    Id                 uint64          // id of the swap message for the pair
    PairId             uint64          // id of the pair where the swap order is placed
    MsgHeight          int64           // block height where this message is appended to the batch
    Orderer            string          // address from which swap order was requested
    Direction          OrderDirection
    OfferCoin          sdk.Coin        // amount of coin provided when requesting a swap
    RemainingOfferCoin sdk.Coin        // remaining amount of offer coin after matching
    ReceivedCoin       sdk.Coin        // amount of received coin after matching
    Price              sdk.Dec         // order price of the swap message
    Amount             sdk.Int         // order amount in base coin of the swap message
    OpenAmount         sdk.Int         // remaining order amount in base coin after matching
    BatchId            uint64          // batch id of the pair when swap order is submitted
    ExpireAt           time.Time       // swap orders are cancelled when current block time is greater than ExpireAt
    Status             OrderStatus
}
```

## MMOrderIndex

`MMOrderIndex` holds the order IDs of a group of limit orders which are
created by `MsgMMOrder`.
This struct exists to cancel previous MM order quickly.

```go
type MMOrderIndex struct {
    Orderer string
    PairId uint64
    OrderIds []uint64
}
```

# Parameter

- ModuleName: `liquidity`
- RouterKey: `liquidity`
- StoreKey: `liquidity`
- QuerierRoute: `liquidity`

# Store

Stores are KVStores in the `multistore`.
The key to find the store is the first parameter in the list.

### The key for the latest pair id

- LastPairIdKey: `[]byte{0xa0} -> ProtocolBuffer(uint64)`

### The key for the latest pool id

- LastPoolIdKey: `[]byte{0xa1} -> ProtocolBuffer(uint64)`

### The key to get the pair object 

- PairKey: `[]byte{0xa5} | PairId -> ProtocolBuffer(Pair)`

### The index key to get the pair object by base and quote denoms

- PairIndexKey: `[]byte{0xa6} | BaseCoinDenomLen (1 byte) | BaseCoinDenom | QuoteCoinDenomLen (1 byte) | QuoteCoinDenom -> ProtocolBuffer(uint64)`

### The index key to lookup pairs with the given denom

- PairsByDenomsIndexKey: `[]byte{0xa7} | DenomALen (1 byte) | DenomA | DenomBLen (1 byte) | DenomB | PairId -> nil`

### The key to get the pool object

- PoolKey: `[]byte{0xab} | PoolId -> ProtocolBuffer(Pool)`

### The index key to get the pool object from the reserve address

- PoolByReserveAddressIndexKey: `[]byte{0xac} | ReserveAddressLen (1 byte) | ReserveAddress -> ProtocolBuffer(uint64)`

### The index key to lookup pools by pair id

- PoolsByPairIndexKey: `[]byte{0xad} | PairId | PoolId -> nil`

### The key to get the deposit request by pool id and deposit request id

- DepositRequestKey: `[]byte{0xb0} | PoolId | DepositRequestId -> ProtocolBuffer(DepositRequest)`

### The index key to get the deposit request by depositor address, pool id and request id

- DepositRequestIndexKey: `[]byte{0xb4} | DepositorAddressLen (1 byte) | DepositorAddress | PoolId | ReqId -> nil`

### The key to get the withdraw request by pool id and withdraw request id

- WithdrawRequestKey: `[]byte{0xb1} | PoolId | WithdrawRequestId -> ProtocolBuffer(WithdrawRequest)`

### The index key to get the withdraw request by withdrawer address, pool id and request id

- WithdrawRequestIndexKey: `[]byte{0xb5} | WithdrawerAddressLen (1 byte) | WithdrawerAddress | PoolId | ReqId -> nil`

### The key to get the order by pair id and order id

- OrderKey: `[]byte{0xb2} | PairId | Id -> ProtocolBuffer(Order)`

### The index key to get the order by orderer address, pair id and order id

- OrderIndexKey: `[]byte{0xb3} | OrdererAddressLen (1 byte) | OrdererAddress | PairId | OrderId -> nil`

### The key to get the MM order index by orderer address and pair id

- MMOrderIndexKey: `[]byte{0xb6} | OrdererAddressLen (1 byte) | OrdererAddress | PairId`
