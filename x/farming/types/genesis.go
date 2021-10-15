package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState returns new GenesisState.
func NewGenesisState(
	params Params, plans []PlanRecord, stakings []StakingRecord, queuedStakings []QueuedStakingRecord,
	historicalRewards []HistoricalRewardsRecord, outstandingRewards []OutstandingRewardsRecord,
	currentEpochs []CurrentEpochRecord, stakingReserveCoins, rewardPoolCoins sdk.Coins,
	lastEpochTime *time.Time, currentEpochDays uint32,
) *GenesisState {
	return &GenesisState{
		Params:                    params,
		PlanRecords:               plans,
		StakingRecords:            stakings,
		QueuedStakingRecords:      queuedStakings,
		HistoricalRewardsRecords:  historicalRewards,
		OutstandingRewardsRecords: outstandingRewards,
		CurrentEpochRecords:       currentEpochs,
		StakingReserveCoins:       stakingReserveCoins,
		RewardPoolCoins:           rewardPoolCoins,
		LastEpochTime:             lastEpochTime,
		CurrentEpochDays:          currentEpochDays,
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
		[]OutstandingRewardsRecord{},
		[]CurrentEpochRecord{},
		sdk.Coins{},
		sdk.Coins{},
		nil,
		DefaultCurrentEpochDays,
	)
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
		plan, _ := UnpackPlan(&record.Plan)
		if plan.GetId() < id {
			return fmt.Errorf("pool records must be sorted")
		}
		plans = append(plans, plan)
		id = plan.GetId() + 1
	}

	if err := ValidateTotalEpochRatio(plans); err != nil {
		return err
	}

	for _, record := range data.StakingRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	for _, record := range data.QueuedStakingRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	for _, record := range data.HistoricalRewardsRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	for _, record := range data.OutstandingRewardsRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	for _, record := range data.CurrentEpochRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	if err := data.StakingReserveCoins.Validate(); err != nil {
		return err
	}
	if err := data.RewardPoolCoins.Validate(); err != nil {
		return err
	}

	if data.CurrentEpochDays == 0 {
		return fmt.Errorf("current epoch days must be positive")
	}

	return nil
}

// Validate validates PlanRecord.
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

// Validate validates StakingRecord.
func (record StakingRecord) Validate() error {
	if _, err := sdk.AccAddressFromBech32(record.Farmer); err != nil {
		return err
	}
	if err := sdk.ValidateDenom(record.StakingCoinDenom); err != nil {
		return err
	}
	if !record.Staking.Amount.IsPositive() {
		return fmt.Errorf("staking amount must be positive: %s", record.Staking.Amount)
	}
	return nil
}

// Validate validates QueuedStakingRecord.
func (record QueuedStakingRecord) Validate() error {
	if _, err := sdk.AccAddressFromBech32(record.Farmer); err != nil {
		return err
	}
	if err := sdk.ValidateDenom(record.StakingCoinDenom); err != nil {
		return err
	}
	if !record.QueuedStaking.Amount.IsPositive() {
		return fmt.Errorf("queued staking amount must be positive: %s", record.QueuedStaking.Amount)
	}
	return nil
}

// Validate validates HistoricalRewardsRecord.
func (record HistoricalRewardsRecord) Validate() error {
	if err := sdk.ValidateDenom(record.StakingCoinDenom); err != nil {
		return err
	}
	if err := record.HistoricalRewards.CumulativeUnitRewards.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates OutstandingRewardsRecord.
func (record OutstandingRewardsRecord) Validate() error {
	if err := sdk.ValidateDenom(record.StakingCoinDenom); err != nil {
		return err
	}
	if err := record.OutstandingRewards.Rewards.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates CurrentEpochRecord.
func (record CurrentEpochRecord) Validate() error {
	if err := sdk.ValidateDenom(record.StakingCoinDenom); err != nil {
		return err
	}
	return nil
}
