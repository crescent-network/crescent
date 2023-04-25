package types

import (
	"fmt"
)

func NewGenesisState(
	params Params, lastSpotMarketId, lastSpotOrderId uint64,
	marketRecords []SpotMarketRecord, orders []SpotOrder) *GenesisState {
	return &GenesisState{
		Params:            params,
		LastSpotMarketId:  lastSpotMarketId,
		LastSpotOrderId:   lastSpotOrderId,
		SpotMarketRecords: marketRecords,
		SpotOrders:        orders,
	}
}

// DefaultGenesis returns the default genesis state for the module.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, 0, nil, nil)
}

func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return err
	}
	for _, marketRecord := range genState.SpotMarketRecords {
		if err := marketRecord.Validate(); err != nil {
			return fmt.Errorf("invalid spot market record: %w", err)
		}
	}
	for _, order := range genState.SpotOrders {
		if err := order.Validate(); err != nil {
			return fmt.Errorf("invalid spot order: %w", err)
		}
	}
	return nil
}

func (record SpotMarketRecord) Validate() error {
	if err := record.Market.Validate(); err != nil {
		return fmt.Errorf("invalid market: %w", err)
	}
	if err := record.State.Validate(); err != nil {
		return fmt.Errorf("invalid market state: %w", err)
	}
	return nil
}
