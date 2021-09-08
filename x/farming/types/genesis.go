package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState returns new GenesisState.
func NewGenesisState(params Params, planRecords []PlanRecord, stakings []Staking, stakingReserveCoins, rewardPoolCoins sdk.Coins, globalLastEpochTime time.Time) *GenesisState {
	return &GenesisState{
		Params:              params,
		PlanRecords:         planRecords,
		Stakings:            stakings,
		StakingReserveCoins: stakingReserveCoins,
		RewardPoolCoins:     rewardPoolCoins,
		GlobalLastEpochTime: globalLastEpochTime,
		// TODO: add queued staking and struct for f1
	}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]PlanRecord{},
		[]Staking{},
		sdk.Coins{},
		sdk.Coins{},
		time.Time{},)
}

// ValidateGenesis validates GenesisState.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	id := uint64(0)
	var plans []PlanI
	for _, record := range data.PlanRecords {
		if err := record.Validate(); err != nil {
			return err
		}
		plan, err := UnpackPlan(&record.Plan)
		if err != nil {
			return err
		}
		if plan.GetId() < id {
			return fmt.Errorf("pool records must be sorted")
		}
		plans = append(plans, plan)
	}
	err := ValidateRatioPlans(plans)
	if err != nil {
		return err
	}

	for _, staking := range data.Stakings {
		if err := staking.Validate(); err != nil {
			return err
		}
	}
	if err := data.RewardPoolCoins.Validate(); err != nil {
		return err
	}
	if err := data.StakingReserveCoins.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates Staking.
func (s Staking) Validate() error {
	// TODO: fix to f1 struct
	//if _, err := sdk.AccAddressFromBech32(s.Farmer); err != nil {
	//	return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", s.Farmer, err)
	//}
	//if err := s.StakedCoins.Validate(); err != nil {
	//	return err
	//}
	//if err := s.QueuedCoins.Validate(); err != nil {
	//	return err
	//}
	return nil
}

// Validate validates Reward.
func (record PlanRecord) Validate() error {
	plan, err := UnpackPlan(&record.Plan)
	if err != nil {
		return err
	}
	if err := plan.Validate(); err != nil {
		return err
	}
	if err := record.FarmingPoolCoins.Validate(); err != nil {
		return err
	}
	return nil
}
