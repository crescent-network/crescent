package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState returns new GenesisState.
func NewGenesisState(
	params Params, globalPlanId uint64, plans []PlanRecord,
	stakings []StakingRecord, queuedStakings []QueuedStakingRecord, totalStakings []TotalStakingsRecord,
	historicalRewards []HistoricalRewardsRecord, outstandingRewards []OutstandingRewardsRecord,
	unharvestedRewards []UnharvestedRewardsRecord, currentEpochs []CurrentEpochRecord,
	rewardPoolCoins sdk.Coins, lastEpochTime *time.Time, currentEpochDays uint32,
) *GenesisState {
	return &GenesisState{
		Params:                    params,
		GlobalPlanId:              globalPlanId,
		PlanRecords:               plans,
		StakingRecords:            stakings,
		QueuedStakingRecords:      queuedStakings,
		TotalStakingsRecords:      totalStakings,
		HistoricalRewardsRecords:  historicalRewards,
		OutstandingRewardsRecords: outstandingRewards,
		UnharvestedRewardsRecords: unharvestedRewards,
		CurrentEpochRecords:       currentEpochs,
		RewardPoolCoins:           rewardPoolCoins,
		LastEpochTime:             lastEpochTime,
		CurrentEpochDays:          currentEpochDays,
	}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		0,
		[]PlanRecord{},
		[]StakingRecord{},
		[]QueuedStakingRecord{},
		[]TotalStakingsRecord{},
		[]HistoricalRewardsRecord{},
		[]OutstandingRewardsRecord{},
		[]UnharvestedRewardsRecord{},
		[]CurrentEpochRecord{},
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

	var plans []PlanI
	for _, record := range data.PlanRecords {
		if err := record.Validate(); err != nil {
			return err
		}
		plan, _ := UnpackPlan(&record.Plan)
		if plan.GetId() > data.GlobalPlanId {
			return fmt.Errorf("plan id is greater than the global last plan id")
		}
		plans = append(plans, plan)
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

	for _, record := range data.TotalStakingsRecords {
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

	for _, record := range data.UnharvestedRewardsRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}

	for _, record := range data.CurrentEpochRecords {
		if err := record.Validate(); err != nil {
			return err
		}
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

// Validate validates StakingRecord.
func (record TotalStakingsRecord) Validate() error {
	if err := sdk.ValidateDenom(record.StakingCoinDenom); err != nil {
		return err
	}
	if !record.Amount.IsPositive() {
		return fmt.Errorf("total staking amount must be positive: %s", record.Amount)
	}
	if err := record.StakingReserveCoins.Validate(); err != nil {
		return err
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

// Validate validates UnharvestedRewardsRecord.
func (record UnharvestedRewardsRecord) Validate() error {
	if _, err := sdk.AccAddressFromBech32(record.Farmer); err != nil {
		return err
	}
	if err := sdk.ValidateDenom(record.StakingCoinDenom); err != nil {
		return err
	}
	if err := record.UnharvestedRewards.Rewards.Validate(); err != nil {
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
