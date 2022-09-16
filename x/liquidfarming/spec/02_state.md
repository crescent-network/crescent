<!-- order: 2 -->

# State

The `liquidfarming` module keeps track of the states of pool coins and LFCoins.

## LiquidFarms

```go
// LiquidFarms tracks the list of the activated LiquidFarms
type LiquidFarms []LiquidFarm
```

## RewardsAuction

```go
// AuctionStatus enumerates the valid status of an auction.
type AuctionStatus int32

const (
	AuctionStatusNil      AuctionStatus = 0
	AuctionStatusStarted  AuctionStatus = 1
	AuctionStatusFinished AuctionStatus = 2
)

type RewardsAuction struct {
	Id                   uint64
	PoolId               uint64 // Corresponding pool id of the target liquid farm
	BiddingCoinDenom     string // corresponding pool coin denom
	PayingReserveAddress string
	StartTime            time.Time
	EndTime              time.Time
	Status               AuctionStatus
	Winner               string // winner's account address
	Rewards              sdk.Coins
}
```

## CompoundingRewards

```go
// RewardsQueued records the amount of pool coins in `FarmQueued` status
// that was converted from the rewards coins by the auction.
type RewardsQueued struct {
	Amount sdk.Int 
}
```

## Bid

```go
// Bid defines a standard bid for an auction.
type Bid struct {
	PoolId      uint64
	Bidder      string
	BiddingCoin sdk.Coin
}
```

## Parameter

- ModuleName: `liquidfarming`
- RouterKey: `liquidfarming`
- StoreKey: `liquidfarming`
- QuerierRoute: `liquidfarming`

## Store

- LastRewardsAuctionIdKey: `[]byte{0xe1} | PoolId -> Uint64Value(uint64)`
- LiquidFarmKey: `[]byte{0xe3} | PoolId -> ProtocolBuffer(LiquidFarm)`
- CompoundingRewardsKey: `[]byte{0xe6} | PoolId -> ProtocolBuffer(CompoundingRewards)`
- RewardsAuctionKey: `[]byte{0xe7} | PoolId | AuctionId -> ProtocolBuffer(RewardsAuction)`
- BidKey: `[]byte{0xea} | PoolId | BidderAddressLen (1 byte) | BidderAddress -> ProtocolBuffer(Bid)`
- WinningBidKey: `[]byte{0xeb} | PoolId | AuctionId -> ProtocolBuffer(WinningBid)`