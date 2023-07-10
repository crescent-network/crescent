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
    LastPrice          *sdk.Dec
    LastMatchingHeight int64
}
```

## Order

* LastOrderId: `0x?? -> BigEndian(LastOrderId)`
* OrderKey: `0x?? | BigEndian(OrderId) -> ProtocoulBuffer(Order)`
* OrderBookOrderIndex: `0x?? | BigEndian(MarketId) | IsBuy | SortableDecBytes(Price) | BigEndian(OrderId) -> BigEndian(OrderId)`

```go
type Order struct {
    Id               uint64
    Type             OrderType
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

type OrderType int32

const (
    OrderTypeUnspecified OrderType = 0
    OrderTypeLimit       OrderType = 1
    OrderTypeMM          OrderType = 2
)
```

## Transient Balance Difference

* TransientBalance: `0x?? | AddrLen (1 byte) | Address | Denom -> ProtocolBuffer(Amount)`
