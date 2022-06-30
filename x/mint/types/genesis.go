package types

import (
	"fmt"
	"time"

	utils "github.com/crescent-network/crescent/v2/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, lastBlockTime *time.Time) *GenesisState {
	return &GenesisState{
		LastBlockTime: lastBlockTime,
		Params:        params,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), nil)
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	if data.LastBlockTime != nil && data.LastBlockTime.Before(utils.ParseTime("0001-01-01T00:00:00Z")) {
		return fmt.Errorf("invalid last block time")
	}
	return data.Params.Validate()
}
