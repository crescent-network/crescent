package types

import (
	fmt "fmt"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                     DefaultParams(),
		LastRewardsAuctionIdRecord: []LastRewardsAuctionIdRecord{},
		LiquidFarms:                []LiquidFarm{},
		RewardsAuctions:            []RewardsAuction{},
		Bids:                       []Bid{},
		WinningBidRecords:          []WinningBidRecord{},
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	for _, liquidFarm := range gs.LiquidFarms {
		if err := liquidFarm.Validate(); err != nil {
			return fmt.Errorf("invalid liquid farm %w", err)
		}
	}

	for _, auction := range gs.RewardsAuctions {
		if err := auction.Validate(); err != nil {
			return err
		}
	}

	for _, bid := range gs.Bids {
		if err := bid.Validate(); err != nil {
			return err
		}
	}

	winningBidMap := map[uint64]Bid{} // AuctionId => Bid
	for _, record := range gs.WinningBidRecords {
		if record.AuctionId == 0 {
			return fmt.Errorf("auction id must not be 0")
		}
		if err := record.WinningBid.Validate(); err != nil {
			return fmt.Errorf("invalid winning bid: %w", err)
		}
		if _, ok := winningBidMap[record.AuctionId]; ok {
			return fmt.Errorf("multiple winning bids at auction %d", record.AuctionId)
		}
		winningBidMap[record.AuctionId] = record.WinningBid
	}

	return nil
}
