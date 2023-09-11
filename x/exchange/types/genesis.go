package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewGenesisState(
	params Params, lastMarketId, lastOrderId uint64,
	marketRecords []MarketRecord, orders []Order, numMMOrdersRecords []NumMMOrdersRecord) *GenesisState {
	return &GenesisState{
		Params:             params,
		LastMarketId:       lastMarketId,
		LastOrderId:        lastOrderId,
		MarketRecords:      marketRecords,
		Orders:             orders,
		NumMMOrdersRecords: numMMOrdersRecords,
	}
}

// DefaultGenesis returns the default genesis state for the module.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, 0, nil, nil, nil)
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
	for _, numMMOrdersRecord := range genState.NumMMOrdersRecords {
		if err := numMMOrdersRecord.Validate(); err != nil {
			return fmt.Errorf("invalid num mm orders record: %w", err)
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

func (record NumMMOrdersRecord) Validate() error {
	if _, err := sdk.AccAddressFromBech32(record.Orderer); err != nil {
		return fmt.Errorf("invalid orderer: %w", err)
	}
	if record.MarketId == 0 {
		return fmt.Errorf("market id must not be 0")
	}
	if record.NumMMOrders == 0 {
		return fmt.Errorf("num mm orders must not be 0")
	}
	return nil
}
