package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState returns a new GenesisState.
func NewGenesisState(
	params Params, lastBlockTime *time.Time, lastPlanId, numPrivatePlans uint64,
	plans []Plan, farms []FarmRecord, positions []Position, hists []HistoricalRewardsRecord,
) *GenesisState {
	return &GenesisState{
		Params:            params,
		LastBlockTime:     lastBlockTime,
		LastPlanId:        lastPlanId,
		NumPrivatePlans:   numPrivatePlans,
		Plans:             plans,
		Farms:             farms,
		Positions:         positions,
		HistoricalRewards: hists,
	}
}

// DefaultGenesis returns the default genesis state for the module.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), nil, 0, 0, nil, nil, nil, nil)
}

func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return err
	}
	planIdSet := map[uint64]struct{}{}
	for _, plan := range genState.Plans {
		if err := plan.Validate(); err != nil {
			return fmt.Errorf("invalid plan: %w", err)
		}
		if _, ok := planIdSet[plan.Id]; ok {
			return fmt.Errorf("duplicate plan: %d", plan.Id)
		}
		planIdSet[plan.Id] = struct{}{}
	}
	farmDenomSet := map[string]struct{}{}
	for _, farm := range genState.Farms {
		if err := sdk.ValidateDenom(farm.Denom); err != nil {
			return fmt.Errorf("invalid farm denom: %s", err)
		}
		if farm.Farm.TotalFarmingAmount.IsNegative() {
			return fmt.Errorf(
				"total farming amount must not be negative: %s", farm.Farm.TotalFarmingAmount)
		}
		if err := farm.Farm.CurrentRewards.Validate(); err != nil {
			return fmt.Errorf("invalid current rewards: %w", err)
		}
		if err := farm.Farm.OutstandingRewards.Validate(); err != nil {
			return fmt.Errorf("invalid outstanding rewards: %w", err)
		}
		if farm.Farm.Period == 0 {
			return fmt.Errorf("period must be positive")
		}
		if _, ok := farmDenomSet[farm.Denom]; ok {
			return fmt.Errorf("duplicate farm: %s", farm.Denom)
		}
		farmDenomSet[farm.Denom] = struct{}{}
	}
	type positionKey struct {
		farmer, denom string
	}
	positionKeySet := map[positionKey]struct{}{}
	for _, position := range genState.Positions {
		if _, err := sdk.AccAddressFromBech32(position.Farmer); err != nil {
			return fmt.Errorf("invalid farmer address: %w", err)
		}
		if err := sdk.ValidateDenom(position.Denom); err != nil {
			return fmt.Errorf("invalid position denom: %w", err)
		}
		if !position.FarmingAmount.IsPositive() {
			return fmt.Errorf("farming amount must be positive: %s", position.FarmingAmount)
		}
		if position.StartingBlockHeight <= 0 {
			return fmt.Errorf(
				"starting block height must be positive: %d", position.StartingBlockHeight)
		}
		key := positionKey{position.Farmer, position.Denom}
		if _, ok := positionKeySet[key]; ok {
			return fmt.Errorf("duplicate position: %s, %s", position.Farmer, position.Denom)
		}
		positionKeySet[key] = struct{}{}
	}
	type historicalRewardsKey struct {
		denom  string
		period uint64
	}
	histKeySet := map[historicalRewardsKey]struct{}{}
	for _, hist := range genState.HistoricalRewards {
		if err := sdk.ValidateDenom(hist.Denom); err != nil {
			return fmt.Errorf("invalid historical rewards denom: %s", err)
		}
		if err := hist.HistoricalRewards.CumulativeUnitRewards.Validate(); err != nil {
			return fmt.Errorf("invalid cumulative unit rewards: %w", err)
		}
		if hist.HistoricalRewards.ReferenceCount == 0 {
			return fmt.Errorf("reference count must be positive")
		}
		if hist.HistoricalRewards.ReferenceCount > 2 {
			return fmt.Errorf("reference count must not exceed 2")
		}
		key := historicalRewardsKey{hist.Denom, hist.Period}
		if _, ok := histKeySet[key]; ok {
			return fmt.Errorf("duplicate historical rewards: %s, %d", hist.Denom, hist.Period)
		}
		histKeySet[key] = struct{}{}
	}
	return nil
}
