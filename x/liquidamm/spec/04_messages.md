<!-- order: 4 -->

# Messages

## MsgMintShare

```go
type MsgMintShare struct {
    Sender           string
    PublicPositionId uint64
    DesiredAmount    sdk.Coins
}
```

## MsgBurnShare

```go
type MsgBurnShare struct {
    Sender           string
    PublicPositionId uint64
    Share            sdk.Coin
}
```

## MsgPlaceBid

```go
type MsgPlaceBid struct {
    Sender           string
    PublicPositionId uint64
    RewardsAuctionId uint64
    Share            sdk.Coin
}
```
