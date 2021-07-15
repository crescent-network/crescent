package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGenesisState returns new GenesisState.
func NewGenesisState(params Params, planRecords []PlanRecord, stakings []Staking, rewards []Reward) *GenesisState {
	return &GenesisState{
		Params:      params,
		PlanRecords: planRecords,
		Stakings:    stakings,
		Rewards:     rewards,
	}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]PlanRecord{},
		[]Staking{},
		[]Reward{},
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
	for _, reward := range data.Rewards {
		if err := reward.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates PlanRecord.
func (r PlanRecord) Validate() error {
	_, err := UnpackPlan(&r.Plan)
	if err != nil {
		return err
	}
	if err := r.FarmingPoolCoins.Validate(); err != nil {
		return err
	}
	if err := r.RewardPoolCoins.Validate(); err != nil {
		return err
	}
	if err := r.StakingReserveCoins.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates Staking.
func (s Staking) Validate() error {
	if _, err := sdk.AccAddressFromBech32(s.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", s.Farmer, err)
	}
	if err := s.StakedCoins.Validate(); err != nil {
		return err
	}
	if err := s.QueuedCoins.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates Reward.
func (r Reward) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", r.Farmer, err)
	}
	if err := r.RewardCoins.Validate(); err != nil {
		return err
	}
	return nil
}
