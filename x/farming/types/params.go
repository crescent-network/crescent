package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// PrivatePlanMaxNumDenoms is the maximum number of denoms in a private plan's
	// staking coin weights and epoch amount.
	PrivatePlanMaxNumDenoms = 50
	// PublicPlanMaxNumDenoms is the maximum number of denoms in a public plan's
	// staking coin weights and epoch amount.
	PublicPlanMaxNumDenoms = 500

	RewardReserveAccName             string = "RewardsReserveAcc"
	UnharvestedRewardsReserveAccName string = "UnharvestedRewardsReserveAcc"
)

// Parameter store keys
var (
	KeyPrivatePlanCreationFee = []byte("PrivatePlanCreationFee")
	KeyNextEpochDays          = []byte("NextEpochDays")
	KeyFarmingFeeCollector    = []byte("FarmingFeeCollector")
	KeyDelayedStakingGasFee   = []byte("DelayedStakingGasFee")
	KeyMaxNumPrivatePlans     = []byte("MaxNumPrivatePlans")

	DefaultPrivatePlanCreationFee = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1_000_000_000)))
	DefaultCurrentEpochDays       = uint32(1)
	DefaultNextEpochDays          = uint32(1)
	DefaultFarmingFeeCollector    = sdk.AccAddress(address.Module(ModuleName, []byte("FarmingFeeCollectorAcc")))
	DefaultDelayedStakingGasFee   = sdk.Gas(60000) // See https://github.com/tendermint/farming/issues/102 for details.
	DefaultMaxNumPrivatePlans     = uint32(10000)

	// ReserveAddressType is an address type of reserve accounts for staking or rewards.
	// The module uses the address type of 32 bytes length, but it can be changed depending on Cosmos SDK's direction.
	// The discussion around this issue can be found in this link.
	// https://github.com/tendermint/farming/issues/200
	ReserveAddressType           = AddressType32Bytes
	RewardsReserveAcc            = DeriveAddress(ReserveAddressType, ModuleName, RewardReserveAccName)
	UnharvestedRewardsReserveAcc = DeriveAddress(ReserveAddressType, ModuleName, UnharvestedRewardsReserveAccName)
)

var _ paramstypes.ParamSet = (*Params)(nil)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the default farming module parameters.
func DefaultParams() Params {
	return Params{
		PrivatePlanCreationFee: DefaultPrivatePlanCreationFee,
		NextEpochDays:          DefaultNextEpochDays,
		FarmingFeeCollector:    DefaultFarmingFeeCollector.String(),
		DelayedStakingGasFee:   DefaultDelayedStakingGasFee,
		MaxNumPrivatePlans:     DefaultMaxNumPrivatePlans,
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyPrivatePlanCreationFee, &p.PrivatePlanCreationFee, validatePrivatePlanCreationFee),
		paramstypes.NewParamSetPair(KeyNextEpochDays, &p.NextEpochDays, validateNextEpochDays),
		paramstypes.NewParamSetPair(KeyFarmingFeeCollector, &p.FarmingFeeCollector, validateFarmingFeeCollector),
		paramstypes.NewParamSetPair(KeyDelayedStakingGasFee, &p.DelayedStakingGasFee, validateDelayedStakingGas),
		paramstypes.NewParamSetPair(KeyMaxNumPrivatePlans, &p.MaxNumPrivatePlans, validateMaxNumPrivatePlans),
	}
}

// String returns a human-readable string representation of the parameters.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates parameters.
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.PrivatePlanCreationFee, validatePrivatePlanCreationFee},
		{p.NextEpochDays, validateNextEpochDays},
		{p.FarmingFeeCollector, validateFarmingFeeCollector},
		{p.DelayedStakingGasFee, validateDelayedStakingGas},
		{p.MaxNumPrivatePlans, validateMaxNumPrivatePlans},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

func validatePrivatePlanCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := v.Validate(); err != nil {
		return err
	}

	return nil
}

func validateNextEpochDays(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("next epoch days must be positive: %d", v)
	}

	return nil
}

func validateFarmingFeeCollector(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == "" {
		return fmt.Errorf("farming fee collector address must not be empty")
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid account address: %v", v)
	}

	return nil
}

func validateDelayedStakingGas(i interface{}) error {
	_, ok := i.(sdk.Gas)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateMaxNumPrivatePlans(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Allow zero MaxNumPrivatePlans
	return nil
}
