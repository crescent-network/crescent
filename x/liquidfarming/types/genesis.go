package types

import (
	"fmt"
	"time"
)

func NewGenesisState(
	params Params, lastLiquidFarmId uint64, liquidFarms []LiquidFarm,
	auctions []RewardsAuction, bids []Bid, nextAuctionEndTime *time.Time) *GenesisState {
	return &GenesisState{
		Params:                    params,
		LastLiquidFarmId:          lastLiquidFarmId,
		LiquidFarms:               liquidFarms,
		RewardsAuctions:           auctions,
		Bids:                      bids,
		NextRewardsAuctionEndTime: nextAuctionEndTime,
	}
}

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, nil, nil, nil, nil)
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	for _, liquidFarm := range genState.LiquidFarms {
		if err := liquidFarm.Validate(); err != nil {
			return fmt.Errorf("invalid liquid farm: %w", err)
		}
	}
	for _, auction := range genState.RewardsAuctions {
		if err := auction.Validate(); err != nil {
			return fmt.Errorf("invalid rewards auction: %w", err)
		}
	}
	for _, bid := range genState.Bids {
		if err := bid.Validate(); err != nil {
			return fmt.Errorf("invalid bid: %w", err)
		}
	}
	return nil
}
