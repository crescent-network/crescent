package types

import "time"

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, lastBlockTime *time.Time) *GenesisState {
	return &GenesisState{
		LastBlockTime: lastBlockTime,
		Params:        params,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		LastBlockTime: nil,
		Params:        DefaultParams(),
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	return data.Params.Validate()
}
