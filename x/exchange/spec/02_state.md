<!-- order: 2 -->

# State

## Market

* LastMarketId: `0x?? -> BigEndian(LastMarketId)`
* Market: `0x?? | BigEndian(MarketId) -> ProtocolBuffer(Market)`
* MarketState: `0x?? | BigEndian(MarketId) -> ProtocolBuffer(Market)`
* MarketByDenomsIndex: `0x?? | DenomLen (1 byte) | BaseDenom | QuoteDenom -> BigEndian(MarketId)`

```go
type Market struct {
    Id            uint64
    BaseDenom     string
    QuoteDenom    string
    EscrowAddress string
    MakerFeeRate  sdk.Dec
    TakerFeeRate  sdk.Dec
}

type MarketState struct {
    LastPrice *sdk.Dec
}
```

## Order

* LastOrderId: `0x?? -> BigEndian(LastOrderId)`
* OrderKey: `0x?? | BigEndian(OrderId) -> ProtocoulBuffer(Order)`
* OrderBookOrder: `0x?? | BigEndian(MarketId) | IsBuy | SortableDecBytes(Price) | BigEndian(OrderId) -> BigEndian(OrderId)`

```go
type Order struct {
    Id               uint64
    Orderer          string
    MarketId         uint64
    IsBuy            bool
    Price            sdk.Dec
    Quantity         sdk.Int
    MsgHeight        int64
    OpenQuantity     sdk.Int
    RemainingDeposit sdk.Int
    Deadline         time.Time
}
```

## Transient Balance Difference

* TransientBalance: `0x?? | AddrLen (1 byte) | Address | Denom -> ProtocolBuffer(Amount)`
