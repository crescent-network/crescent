package types

import (
	"fmt"
)

// NewGenesisState returns new GenesisState.
func NewGenesisState(
	params Params, marketMakers []MarketMaker, incentives []Incentive, depositRecords []DepositRecord,
) *GenesisState {
	return &GenesisState{
		Params:         params,
		MarketMakers:   marketMakers,
		Incentives:     incentives,
		DepositRecords: depositRecords,
	}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]MarketMaker{},
		[]Incentive{},
		[]DepositRecord{},
	)
}

// ValidateGenesis validates GenesisState.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	for _, record := range data.MarketMakers {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	for _, record := range data.Incentives {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	for _, record := range data.DepositRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	if err := ValidateDepositRecords(data.MarketMakers, data.DepositRecords); err != nil {
		return err
	}
	return nil
}

func ValidateDepositRecords(mms []MarketMaker, DepositRecords []DepositRecord) error {
	// not eligible market maker must have deposit record
	for _, mm := range mms {
		if !mm.Eligible {
			found := false
			for _, record := range DepositRecords {
				if record.PairId == mm.PairId && record.Address == mm.Address {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("deposit invariant failed, not eligible market maker must have deposit record")
			}
		}
	}

	// deposit record's market maker must not be eligible
	for _, record := range DepositRecords {
		found := false
		for _, mm := range mms {
			if !mm.Eligible && record.PairId == mm.PairId && record.Address == mm.Address {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("deposit invariant failed, deposit record's market maker must not be eligible")
		}
	}
	return nil
}
