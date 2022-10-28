package types

import (
	"fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyFeeCollector           = []byte("FeeCollector")
	KeyRewardsAuctionDuration = []byte("RewardsAuctionDuration")
	KeyLiquidFarms            = []byte("LiquidFarms")
)

// Default parameters
var (
	DefaultFeeCollector           = sdk.AccAddress(address.Module(ModuleName, []byte("FeeCollector")))
	DefaultRewardsAuctionDuration = time.Hour * 8
	DefaultLiquidFarms            = []LiquidFarm{}
)

var _ paramstypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for the module
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		FeeCollector:           DefaultFeeCollector.String(),
		RewardsAuctionDuration: DefaultRewardsAuctionDuration,
		LiquidFarms:            DefaultLiquidFarms,
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyFeeCollector, &p.FeeCollector, validateFeeCollector),
		paramstypes.NewParamSetPair(KeyRewardsAuctionDuration, &p.RewardsAuctionDuration, validateRewardsAuctionDuration),
		paramstypes.NewParamSetPair(KeyLiquidFarms, &p.LiquidFarms, validateLiquidFarms),
	}
}

// Validate validates the set of parameters
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.FeeCollector, validateFeeCollector},
		{p.RewardsAuctionDuration, validateRewardsAuctionDuration},
		{p.LiquidFarms, validateLiquidFarms},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

func validateFeeCollector(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid fee collector address: %v", v)
	}
	return nil
}

func validateRewardsAuctionDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("rewards auction duration must be positive: %d", v)
	}
	return nil
}

func validateLiquidFarms(i interface{}) error {
	liquidFarms, ok := i.([]LiquidFarm)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for _, l := range liquidFarms {
		if err := l.Validate(); err != nil {
			return fmt.Errorf("invalid liquid farm: %v", err)
		}
	}
	return nil
}
