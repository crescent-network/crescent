package types

import (
	"fmt"
	"time"
)

func NewGenesisState(
	params Params, lastPublicPositionId uint64, publicPositions []PublicPosition,
	auctions []RewardsAuction, bids []Bid, nextAuctionEndTime *time.Time) *GenesisState {
	return &GenesisState{
		Params:                    params,
		LastPublicPositionId:      lastPublicPositionId,
		PublicPositions:           publicPositions,
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
	for _, publicPosition := range genState.PublicPositions {
		if err := publicPosition.Validate(); err != nil {
			return fmt.Errorf("invalid public position: %w", err)
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
