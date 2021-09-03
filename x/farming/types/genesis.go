package types

import (
	"fmt"
)

// NewGenesisState returns new GenesisState.
func NewGenesisState(params Params, planRecords []PlanRecord, stakings []Staking) *GenesisState {
	return &GenesisState{
		Params:      params,
		PlanRecords: planRecords,
		Stakings:    stakings,
	}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]PlanRecord{},
		[]Staking{},
	)
}

// ValidateGenesis validates GenesisState.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	for _, record := range data.PlanRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, staking := range data.Stakings {
		if err := staking.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates PlanRecord.
func (r PlanRecord) Validate() error {
	if err := r.FarmingPoolCoins.Validate(); err != nil {
		return err
	}
	if err := r.StakingReserveCoins.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates Staking.
func (s Staking) Validate() error {
	if !s.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}
