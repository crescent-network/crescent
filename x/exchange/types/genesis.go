package types

import (
	"fmt"
)

func NewGenesisState(
	params Params, lastMarketId, lastOrderId uint64,
	marketRecords []MarketRecord, orders []Order) *GenesisState {
	return &GenesisState{
		Params:        params,
		LastMarketId:  lastMarketId,
		LastOrderId:   lastOrderId,
		MarketRecords: marketRecords,
		Orders:        orders,
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
	for _, marketRecord := range genState.MarketRecords {
		if err := marketRecord.Validate(); err != nil {
			return fmt.Errorf("invalid market record: %w", err)
		}
	}
	for _, order := range genState.Orders {
		if err := order.Validate(); err != nil {
			return fmt.Errorf("invalid order: %w", err)
		}
	}
	return nil
}

func (record MarketRecord) Validate() error {
	if err := record.Market.Validate(); err != nil {
		return fmt.Errorf("invalid market: %w", err)
	}
	if err := record.State.Validate(); err != nil {
		return fmt.Errorf("invalid market state: %w", err)
	}
	return nil
}
