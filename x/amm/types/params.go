package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramstypes.ParamSet = (*Params)(nil)

const (
	MinTick = -1260000 // 0.000000000000010000
	MaxTick = 3600000  // 10000000000000000000000000000000000000000
)

var (
	KeyPoolCreationFee               = []byte("PoolCreationFee")
	KeyDefaultTickSpacing            = []byte("DefaultTickSpacing")
	KeyPrivateFarmingPlanCreationFee = []byte("PrivateFarmingPlanCreationFee")
	KeyMaxNumPrivateFarmingPlans     = []byte("MaxNumPrivateFarmingPlans")
	KeyMaxFarmingBlockTime           = []byte("MaxFarmingBlockTime")
)

var (
	DefaultPoolCreationFee               = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
	DefaultDefaultTickSpacing            = uint32(50)
	DefaultPrivateFarmingPlanCreationFee = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
	DefaultMaxNumPrivateFarmingPlans     = uint32(50)
	DefaultMaxFarmingBlockTime           = 10 * time.Second

	AllowedTickSpacings = []uint32{1, 5, 10, 50}
	// DecMulFactor is multiplied to fee and farming rewards growth variables
	// so that small amount of rewards can be handled correctly.
	DecMulFactor = sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 12))
)

func IsAllowedTickSpacing(tickSpacing uint32) bool {
	for _, ts := range AllowedTickSpacings {
		if tickSpacing == ts {
			return true
		}
	}
	return false
}

func ValidateTickSpacing(prevTickSpacing, tickSpacing uint32) error {
	if !IsAllowedTickSpacing(tickSpacing) {
		return fmt.Errorf("tick spacing %d is not allowed", tickSpacing)
	}
	if prevTickSpacing%tickSpacing != 0 {
		return fmt.Errorf("tick spacing must be a divisor of previous tick spacing %d", prevTickSpacing)
	}
	return nil
}

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default params for the module.
func DefaultParams() Params {
	return Params{
		PoolCreationFee:               DefaultPoolCreationFee,
		DefaultTickSpacing:            DefaultDefaultTickSpacing,
		PrivateFarmingPlanCreationFee: DefaultPrivateFarmingPlanCreationFee,
		MaxNumPrivateFarmingPlans:     DefaultMaxNumPrivateFarmingPlans,
		MaxFarmingBlockTime:           DefaultMaxFarmingBlockTime,
	}
}

// ParamSetPairs implements ParamSet.
func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyPoolCreationFee, &params.PoolCreationFee, validatePoolCreationFee),
		paramstypes.NewParamSetPair(KeyDefaultTickSpacing, &params.DefaultTickSpacing, validateDefaultTickSpacing),
		paramstypes.NewParamSetPair(KeyPrivateFarmingPlanCreationFee, &params.PrivateFarmingPlanCreationFee, validatePrivateFarmingPlanCreationFee),
		paramstypes.NewParamSetPair(KeyMaxNumPrivateFarmingPlans, &params.MaxNumPrivateFarmingPlans, validateMaxNumPrivateFarmingPlans),
		paramstypes.NewParamSetPair(KeyMaxFarmingBlockTime, &params.MaxFarmingBlockTime, validateMaxFarmingBlockTime),
	}
}

// Validate validates Params.
func (params Params) Validate() error {
	for _, field := range []struct {
		val          interface{}
		validateFunc func(i interface{}) error
	}{
		{params.PoolCreationFee, validatePoolCreationFee},
		{params.DefaultTickSpacing, validateDefaultTickSpacing},
		{params.PrivateFarmingPlanCreationFee, validatePrivateFarmingPlanCreationFee},
		{params.MaxNumPrivateFarmingPlans, validateMaxNumPrivateFarmingPlans},
		{params.MaxFarmingBlockTime, validateMaxFarmingBlockTime},
	} {
		if err := field.validateFunc(field.val); err != nil {
			return err
		}
	}
	return nil
}

func validatePoolCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid pool creation fee: %w", err)
	}
	return nil
}

func validateDefaultTickSpacing(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !IsAllowedTickSpacing(v) {
		return fmt.Errorf("tick spacing %d is not allowed", v)
	}
	return nil
}

func validatePrivateFarmingPlanCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid private farming plan creation fee: %w", err)
	}
	return nil
}

func validateMaxNumPrivateFarmingPlans(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateMaxFarmingBlockTime(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("max farming block time must be positive: %v", v)
	}
	return nil
}
