package types

import (
	"time"
)

// NewGenesisState returns a new GenesisState.
func NewGenesisState(
	params Params, lastPlanId uint64, plans []Plan, farms []FarmRecord, positions []Position,
	hists []HistoricalRewardsRecord, lastBlockTime *time.Time,
) *GenesisState {
	return &GenesisState{
		Params:            params,
		LastPlanId:        lastPlanId,
		Plans:             plans,
		Farms:             farms,
		Positions:         positions,
		HistoricalRewards: hists,
		LastBlockTime:     lastBlockTime,
	}
}

// DefaultGenesis returns the default genesis state for the module.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, nil, nil, nil, nil, nil)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (genState GenesisState) Validate() error {
	// TODO: implement
	return nil
}
