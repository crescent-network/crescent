package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState returns new GenesisState.
func NewGenesisState(
	params Params, plans []PlanRecord, stakings []StakingRecord, queuedStakings []QueuedStakingRecord,
	historicalRewards []HistoricalRewardsRecord, currentEpochs []CurrentEpochRecord, stakingReserveCoins,
	rewardPoolCoins sdk.Coins, globalLastEpochTime time.Time,
) *GenesisState {
	return &GenesisState{
		Params:                   params,
		PlanRecords:              plans,
		StakingRecords:           stakings,
		QueuedStakingRecords:     queuedStakings,
		HistoricalRewardsRecords: historicalRewards,
		CurrentEpochRecords:      currentEpochs,
		StakingReserveCoins:      stakingReserveCoins,
		RewardPoolCoins:          rewardPoolCoins,
		GlobalLastEpochTime:      globalLastEpochTime,
	}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]PlanRecord{},
		[]StakingRecord{},
		[]QueuedStakingRecord{},
		[]HistoricalRewardsRecord{},
		[]CurrentEpochRecord{},
		sdk.Coins{},
		sdk.Coins{},
		time.Time{})
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
		id = plan.GetId() + 1
	}

	if err := ValidateName(plans); err != nil {
		return err
	}

	if err := ValidateTotalEpochRatio(plans); err != nil {
		return err
	}

	// TODO: validate other fields
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
