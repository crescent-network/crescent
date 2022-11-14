<!-- order: 2 -->

# State

The `liquidfarming` module keeps track of the states of pool coins and LFCoins.

## LiquidFarm

```go
// LiquidFarms tracks the list of the activated LiquidFarms
type LiquidFarms struct {
	liquidfarms []LiquidFarm
}

// LiquidFarm defines liquid farm.
type LiquidFarm struct {
	PoolId				uint64		// the pool id
	MinFarmAmount		sdk.Int		// the minimum farm amount; it allows zero value
	MinBidAmount		sdk.Int		// the minimum bid amount; it allows zero value
	FeeRate				sdk.Dec		// default value is 12 hours
}
```

## RewardsAuction

```go
// AuctionStatus enumerates the valid status of an auction.
type AuctionStatus int32

const (
	AuctionStatusNil			AuctionStatus = 0
	AuctionStatusStarted		AuctionStatus = 1
	AuctionStatusFinished		AuctionStatus = 2
	AuctionStatusSkipped		AuctionStatus = 3
)

// RewardsAuction defines rewards auction information.
type RewardsAuction struct {
	Id                   uint64        // rewards auction id
	PoolId               uint64        // corresponding pool id of the target liquid farm
	BiddingCoinDenom     string        // corresponding pool coin denom
	PayingReserveAddress string        // the paying reserve address that collects bidding coin placed by bidders
	StartTime            time.Time     // the auction start time
	EndTime              time.Time     // the auction end time
	Status               AuctionStatus // the auction status
	Winner               string        // the bidder who won the auction
	WinningAmount        sdk.Coin      // the winning amount placed by the winner
	Rewards              sdk.Coins     // the farming rewards for are accumulated every block
	Fees                 sdk.Coins     // the fees for the rewards by the fee rate
	FeeRate              sdk.Dec       // the fee rate for the liquid farm
}
```

## CompoundingRewards

```go
// CompoundingRewards records the amount of farming rewards
type CompoundingRewards struct {
	Amount sdk.Int
}
```

## Bid

```go
// Bid defines a standard bid for an auction.
type Bid struct {
	PoolId      uint64
	Bidder      string
	Amount sdk.Coin
}
```

## Parameter

- ModuleName: `liquidfarming`
- RouterKey: `liquidfarming`
- StoreKey: `liquidfarming`
- QuerierRoute: `liquidfarming`

## Store

- LastRewardsAuctionEndTimeKey: `[]byte{0xe1} -> Timestamp(time.Time)`
- LastRewardsAuctionIdKey: `[]byte{0xe2} | PoolId -> Uint64Value(uint64)`
- LiquidFarmKey: `[]byte{0xe3} | PoolId -> ProtocolBuffer(LiquidFarm)`
- CompoundingRewardsKey: `[]byte{0xe4} | PoolId -> ProtocolBuffer(CompoundingRewards)`
- RewardsAuctionKey: `[]byte{0xe5} | AuctionId | PoolId -> ProtocolBuffer(RewardsAuction)`
- BidKey: `[]byte{0xe6} | PoolId | BidderAddressLen (1 byte) | BidderAddress -> ProtocolBuffer(Bid)`
- WinningBidKey: `[]byte{0xe7} | AuctionId | PoolId -> ProtocolBuffer(Bid)`
