package types

import (
	"time"
)

func NewGenesisState(params Params, lastBlockTime *time.Time) *GenesisState {
	return &GenesisState{
		Params:        params,
		LastBlockTime: lastBlockTime,
	}
}

// DefaultGenesis returns the default genesis state for the module.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), nil)
}

// Validate validates GenesisState.
func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return err
	}
	return nil
}
