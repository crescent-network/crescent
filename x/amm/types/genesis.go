package types

import (
	"fmt"
)

func NewGenesisState(
	params Params, lastPoolId, lastPositionId uint64,
	poolRecords []PoolRecord, positions []Position, tickInfoRecords []TickInfoRecord,
	lastFarmingPlanId uint64, numPrivateFarmingPlans uint32, farmingPlans []FarmingPlan) *GenesisState {
	return &GenesisState{
		Params:                 params,
		LastPoolId:             lastPoolId,
		LastPositionId:         lastPositionId,
		PoolRecords:            poolRecords,
		Positions:              positions,
		TickInfoRecords:        tickInfoRecords,
		LastFarmingPlanId:      lastFarmingPlanId,
		NumPrivateFarmingPlans: numPrivateFarmingPlans,
		FarmingPlans:           farmingPlans,
	}
}

// DefaultGenesis returns the default genesis state for the module.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, 0, nil, nil, nil, 0, 0, nil)
}

func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return err
	}
	for _, poolRecord := range genState.PoolRecords {
		if err := poolRecord.Validate(); err != nil {
			return fmt.Errorf("invalid pool record: %w", err)
		}
	}
	for _, position := range genState.Positions {
		if err := position.Validate(); err != nil {
			return fmt.Errorf("invalid position: %w", err)
		}
	}
	for _, tickInfoRecord := range genState.TickInfoRecords {
		if err := tickInfoRecord.Validate(); err != nil {
			return fmt.Errorf("invalid tick info record: %w", err)
		}
	}
	for _, plan := range genState.FarmingPlans {
		if err := plan.Validate(); err != nil {
			return fmt.Errorf("invalid farming plan: %w", err)
		}
	}
	return nil
}

func (record PoolRecord) Validate() error {
	if err := record.Pool.Validate(); err != nil {
		return fmt.Errorf("invalid pool: %w", err)
	}
	if err := record.State.Validate(); err != nil {
		return fmt.Errorf("invalid pool state: %w", err)
	}
	return nil
}

func (record TickInfoRecord) Validate() error {
	if record.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if err := record.TickInfo.Validate(); err != nil {
		return fmt.Errorf("invalid tick info: %w", err)
	}
	return nil
}
