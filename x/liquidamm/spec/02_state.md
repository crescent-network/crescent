<!-- order: 2 -->

# State

## PublicPosition

* LastPublicPositionId: `0x81 -> BigEndian(LastPublicPositionId)`
* PublicPosition: `0x83 | BigEndian(PublicPositionId) -> ProtocolBuffer(PublicPosition)`
* PublicPositionsByPoolIndex: `0x84 | BigEndian(PoolId) | BigEndian(PublicPositionId) -> nil`
* PublicPositionByParamsIndex: `0x85 | AddrLen (1 byte) | Owner | BigEndian(PoolId) | Sign (1 byte) | BigEndian(LowerTick) | Sign (1 byte) | BigEndian(UpperTick) -> nil`

```go
type PublicPosition struct {
    Id                   uint64
    PoolId               uint64
    LowerTick            int32
    UpperTick            int32
    BidReserveAddress    string
    MinBidAmount         sdk.Int
    FeeRate              sdk.Dec
    LastRewardsAuctionId uint64
}
```

## RewardsAuction

* LastRewardsAuctionEndTime: `0x82 -> FormatTimeBytes(LastRewardsAuctionEndTime)`
* RewardsAuction: `0x86 | BigEndian(PublicPositionId) | BigEndian(AuctionId) -> ProtocolBuffer(RewardsAuction)`

```go
type AuctionStatus int32

const (
    AuctionStatusNil      AuctionStatus = 0
    AuctionStatusStarted  AuctionStatus = 1
    AuctionStatusFinished AuctionStatus = 2
    AuctionStatusSkipped  AuctionStatus = 3
)

type RewardsAuction struct {
    PublicPositionId uint64
    Id               uint64
    StartTime        time.Time
    EndTime          time.Time
    Status           AuctionStatus
    WinningBid       *Bid
    Rewards          sdk.Coins
    Fees             sdk.Coins
}
```

## Bid

* Bid: `0x87 | BigEndian(PublicPositionId) | BigEndian(AuctionId) | Bidder -> ProtocolBuffer(Bid)`

```go
type Bid struct {
    PublicPositionId uint64
    RewardsAuctionId uint64
    Bidder           string
    Share            sdk.Coin
}
```
