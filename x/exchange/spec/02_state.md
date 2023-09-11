<!-- order: 2 -->

# State

## Market

* LastMarketId: `0x60 -> BigEndian(LastMarketId)`
* Market: `0x62 | BigEndian(MarketId) -> ProtocolBuffer(Market)`
* MarketState: `0x63 | BigEndian(MarketId) -> ProtocolBuffer(Market)`
* MarketByDenomsIndex: `0x64 | DenomLen (1 byte) | BaseDenom | QuoteDenom -> BigEndian(MarketId)`

```go
type Market struct {
    Id                  uint64
    BaseDenom           string
    QuoteDenom          string
    EscrowAddress       string
    MakerFeeRate        sdk.Dec
    TakerFeeRate        sdk.Dec
    OrderSourceFeeRatio sdk.Dec
}

type MarketState struct {
    LastPrice          *sdk.Dec
    LastMatchingHeight int64
}
```

## Order

* LastOrderId: `0x61 -> BigEndian(LastOrderId)`
* OrderKey: `0x65 | BigEndian(OrderId) -> ProtocoulBuffer(Order)`
* OrderBookOrderIndex: `0x66 | BigEndian(MarketId) | IsBuy | SortableDecBytes(Price) | BigEndian(OrderId) -> BigEndian(OrderId)`
* OrdersByOrdererIndex: `0x67 | AddrLen (1 byte) | Orderer | BigEndian(MarketId) | BigEndian(OrderId) -> nil`
* NumMMOrders: `0x68 | AddrLen (1 byte) | Orderer | BigEndian(MarketId) -> BigEndian(NumMMOrders)`

```go
type Order struct {
    Id               uint64
    Type             OrderType
    Orderer          string
    MarketId         uint64
    IsBuy            bool
    Price            sdk.Dec
    Quantity         sdk.Dec
    MsgHeight        int64
    OpenQuantity     sdk.Dec
    RemainingDeposit sdk.Dec
    Deadline         time.Time
}

type OrderType int32

const (
    OrderTypeUnspecified OrderType = 0
    OrderTypeLimit       OrderType = 1
    OrderTypeMM          OrderType = 2
)
```
