<!-- order: 4 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network.
The Cosmos SDK wraps and unwraps `liquidfarming` module messages from transactions.

## MsgLiquidFarm

Farm coin to liquid farm. Farming coins are the pool coin that starts with pool prefix, which is a pool coin of a corresponding pool.
It is important to note that a farmer is not receiving a synthetic version of the farming coins right away.
It are expected to receive the synthetic version of the farming coins after one epoch at the current mint rate.
A synthetic version of the farming coin is called as LFCoin in the module and the terminology is used throughout the documentation and codebase.

```go
type MsgLiquidFarm struct {
	PoolId      uint64   // target pool id
	Farmer      string   // the bech32-encoded address that farms coin
	FarmingCoin sdk.Coin // farming amount of pool coin
}
```

Validity checks are performed for `MsgLiquidFarm` message. The transaction that is triggered with the `MsgLiquidFarm` message fails if:

- The target liquid farm with the pool id does not exist
- The amount of farming coin is not positive
- The amount of farming coin is less than `MinimumFarmAmount`
- The farming coin denom is not the same as the pool coin denom of the pool with `PoolId`
- The farmer has insufficient spendable balances for the farming coin amount

## MsgLiquidUnfarm

Unfarm LFCoin to liquid unfarm.
The module burns LFCoin amounts and releases the corresponding amount of pool coins to a farmer at the current burn rate.

```go
type MsgLiquidUnfarm struct {
	PoolId        uint64   // target deposit request id
	Farmer        string   // the bech32-encoded address that unfarms liquid farm coin
	UnfarmingCoin sdk.Coin // withdrawing amount of LF coin
}
```

Validity checks are performed for `MsgLiquidUnfarm` message. The transaction that is triggered with the `MsgLiquidUnfarm` message fails if:

- The target liquid farm with the pool id does not exist
- The amount of LF coins is not positive
- The unfarming coin denom is not the same as the pool coin denom of the pool with `PoolId`
- The farmer has insufficient spendable balances for the unfarming amount

## MsgLiquidUnfarmAndWithdraw

Unfarm LFCoin to liquid unfarm and withdraw the pool coin from the pool.
The module burns LFCoin amounts at the current burn rate, withdraw the corresponding amount of pool coins from the pool, and then releases the withdrawn coins to a farmer.

```go
type MsgLiquidUnfarmAndWithdraw struct {
	PoolId        uint64   // target pool id
	Farmer        string   // the bech32-encoded address that unfarms liquid farm coin and withdraws
	UnfarmingCoin sdk.Coin // withdrawing amount of LF coin
}
```

Validity checks are performed for `MsgLiquidUnfarmAndWithdraw` message. The transaction that is triggered with the `MsgLiquidUnfarmAndWithdraw` message fails if:

- The target liquid farm with the pool id does not exist
- The amount of LF coins is not positive
- The unfarming coin denom is not the same as the pool coin denom of the pool with `PoolId`
- The farmer has insufficient spendable balances for the unfarming amount

## MsgPlaceBid

Place a bid for a rewards auction.
Anyone can place a bid for an auction where the bidder placing with the highest bid amount takes all the rewards.

```go
type MsgPlaceBid struct {
	AuctionId   uint64   // auction id
	PoolId      uint64   // target pool id
	Bidder      string   // the bech32-encoded address that places a bid
	BiddingCoin sdk.Coin // bidding amount of pool coin
}
```

Validity checks are performed for `MsgPlaceBid` message. The transaction that is triggered with the `MsgPlaceBid` message fails if:

- The target liquid farm with the pool id does not exist
- The target auction status is in invalid status

## MsgRefundBid

Refund the bid that is not winning for the auction.

```go
type MsgRefundBid struct {
	AuctionId uint64 // auction id
	PoolId    uint64 // target pool id
	Bidder    string // the bech32-encoded address that refunds a bid
}
```

Validity checks are performed for `MsgRefundBid` message. The transaction that is triggered with the `MsgRefundBid` message fails if:

- The target liquid farm with the pool id does not exist
- The target auction status is in invalid status
- The bid by the bidder in the auction of the liquid farm with the pool id does not exist

## MsgAdvanceAuction

***This message is disabled by default, you have to build the binary with `make install-testing` to activate this message.***

For testing purposes only, this custom message is used to advance auction.

When you send the `MsgAdvanceAuction` message to the network, finish the ongoing rewards auction and create the next one.

```go
type MsgAdvanceAuction struct {
	Requester string // requester defines the bech32-encoded address of the requester
}
```