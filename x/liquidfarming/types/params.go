package types

import (
	"fmt"
	time "time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyLiquidFarms            = []byte("LiquidFarms")
	KeyRewardsAuctionDuration = []byte("RewardsAuctionDuration")

	DefaultLiquidFarms                          = []LiquidFarm{}
	DefaultRewardsAuctionDuration time.Duration = time.Hour * 12
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		LiquidFarms:            DefaultLiquidFarms,
		RewardsAuctionDuration: DefaultRewardsAuctionDuration,
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyLiquidFarms, &p.LiquidFarms, validateLiquidFarms),
		paramtypes.NewParamSetPair(KeyRewardsAuctionDuration, &p.RewardsAuctionDuration, validateRewardsAuctionDuration),
	}
}

// Validate validates the set of parameters
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.LiquidFarms, validateLiquidFarms},
		{p.RewardsAuctionDuration, validateRewardsAuctionDuration},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

func validateLiquidFarms(i interface{}) error {
	liquidFarms, ok := i.([]LiquidFarm)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, liquidFarm := range liquidFarms {
		if err := liquidFarm.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func validateRewardsAuctionDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("auction period hours must be positive: %d", v)
	}
	return nil
}
