package types

import (
	fmt "fmt"
	time "time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// NewRewardsAuction creates a new RewardsAuction.
func NewRewardsAuction(
	id uint64,
	poolId uint64,
	startTime time.Time,
	endTime time.Time,
) RewardsAuction {
	return RewardsAuction{
		Id:                   id,
		PoolId:               poolId,
		BiddingCoinDenom:     liquiditytypes.PoolCoinDenom(poolId),
		PayingReserveAddress: PayingReserveAddress(poolId).String(),
		StartTime:            startTime,
		EndTime:              endTime,
		Status:               AuctionStatusStarted,
		Winner:               "",            // the value is determined when the auction is finished
		WinningAmount:        sdk.Coin{},    // the value is determined when the auction is finished
		Rewards:              sdk.Coins{},   // the value is determined when the auction is finished
		Fees:                 sdk.Coins{},   // the value is determined when the auction is finished
		FeeRate:              sdk.ZeroDec(), // the value is determined when the auction is finished
	}
}

// Validate validates RewardsAuction.
func (a *RewardsAuction) Validate() error {
	if a.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if a.BiddingCoinDenom == "" {
		return fmt.Errorf("denom must not be empty")
	}
	if err := sdk.ValidateDenom(a.BiddingCoinDenom); err != nil {
		return fmt.Errorf("invalid coin denom")
	}
	if _, err := sdk.AccAddressFromBech32(a.PayingReserveAddress); err != nil {
		return fmt.Errorf("invalid paying reserve address %w", err)
	}
	if !a.EndTime.After(a.StartTime) {
		return fmt.Errorf("end time must be set after the start time")
	}
	if a.Status != AuctionStatusStarted && a.Status != AuctionStatusFinished {
		return fmt.Errorf("invalid auction status")
	}
	return nil
}

// SetStatus sets rewards auction status.
func (a *RewardsAuction) SetStatus(status AuctionStatus) {
	a.Status = status
}

// SetWinner sets winner address.
func (a *RewardsAuction) SetWinner(winner string) {
	a.Winner = winner
}

// SetWinningAmount sets the winning amount.
func (a *RewardsAuction) SetWinningAmount(winningAmount sdk.Coin) {
	a.WinningAmount = winningAmount
}

// SetRewards sets auction rewards.
func (a *RewardsAuction) SetRewards(rewards sdk.Coins) {
	a.Rewards = rewards
}

// SetFees sets auction fees.
func (a *RewardsAuction) SetFees(fees sdk.Coins) {
	a.Fees = fees
}

// SetFeeRate sets auction fee rate.
func (a *RewardsAuction) SetFeeRate(feeRate sdk.Dec) {
	a.FeeRate = feeRate
}

// GetPayingReserveAddress returns the paying reserve address in the form of sdk.AccAddress.
func (a RewardsAuction) GetPayingReserveAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(a.PayingReserveAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewBid creates a new Bid.
func NewBid(poolId uint64, bidder string, amount sdk.Coin) Bid {
	return Bid{
		PoolId: poolId,
		Bidder: bidder,
		Amount: amount,
	}
}

// GetBidder returns the bidder address in the form of sdk.AccAddress.
func (b Bid) GetBidder() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(b.Bidder)
	if err != nil {
		panic(err)
	}
	return addr
}

// Validate validates Bid.
func (b Bid) Validate() error {
	if b.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(b.Bidder); err != nil {
		return fmt.Errorf("invalid bidder address %w", err)
	}
	if !b.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive value")
	}
	if err := b.Amount.Validate(); err != nil {
		return fmt.Errorf("invalid bid amount %w", err)
	}
	return nil
}

// MustMarshalRewardsAuction marshals RewardsAuction and
// it panics upon failure.
func MustMarshalRewardsAuction(cdc codec.BinaryCodec, auction RewardsAuction) []byte {
	return cdc.MustMarshal(&auction)
}

// MustUnmarshalRewardsAuction unmarshals RewardsAuction and
// it panics upon failure.
func MustUnmarshalRewardsAuction(cdc codec.BinaryCodec, value []byte) RewardsAuction {
	pair, err := UnmarshalRewardsAuction(cdc, value)
	if err != nil {
		panic(err)
	}
	return pair
}

// UnmarshalRewardsAuction unmarshals RewardsAuction.
func UnmarshalRewardsAuction(cdc codec.BinaryCodec, value []byte) (auction RewardsAuction, err error) {
	err = cdc.Unmarshal(value, &auction)
	return auction, err
}

// UnmarshalBid unmarshals bid from a store value.
func UnmarshalBid(cdc codec.BinaryCodec, value []byte) (bid Bid, err error) {
	err = cdc.Unmarshal(value, &bid)
	return bid, err
}
