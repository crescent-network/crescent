<!-- order: 3 -->

# Messages

## MsgCreateMarket

```go
type MsgCreateMarket struct {
    Sender     string
    BaseDenom  string
    QuoteDenom string
}
```

## MsgPlaceLimitOrder

```go
type MsgPlaceLimitOrder struct {
    Sender   string
    MarketId uint64
    IsBuy    bool
    Price    sdk.Dec
    Quantity sdk.Int
    IsBatch  bool
    Lifespan time.Duration
}
```

## MsgPlaceMarketOrder

```go
type MsgPlaceMarketOrder struct {
    Sender   string
    MarketId uint64
    IsBuy    bool
    Quantity sdk.Int
}
```

## MsgCancelOrder

```go
type MsgCancelOrder struct {
    Sender  string
    OrderId uint64
}
```

## MsgSwapExactAmountIn

```go
type MsgSwapExactAmountIn struct {
    Sender    string
    Routes    []uint64
    Input     types.Coin
    MinOutput types.Coin
}
```
