package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramstypes.ParamSet = (*Params)(nil)

var (
	KeyPrivatePlanCreationFee = []byte("PrivatePlanCreationFee")
	KeyFeeCollector           = []byte("FeeCollector")
	KeyMaxNumPrivatePlans     = []byte("MaxNumPrivatePlans")
)

const (
	DefaultMaxNumPrivatePlans = 50
)

var (
	DefaultPrivatePlanCreationFee = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000000))
	DefaultFeeCollector           = sdk.AccAddress(address.Module(ModuleName, []byte("FeeCollector"))).String()
)

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default params for the module.
func DefaultParams() Params {
	return Params{
		PrivatePlanCreationFee: DefaultPrivatePlanCreationFee,
		FeeCollector:           DefaultFeeCollector,
		MaxNumPrivatePlans:     DefaultMaxNumPrivatePlans,
	}
}

// ParamSetPairs implements ParamSet.
func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyPrivatePlanCreationFee, &params.PrivatePlanCreationFee, validatePrivatePlanCreationFee),
		paramstypes.NewParamSetPair(KeyFeeCollector, &params.FeeCollector, validateFeeCollector),
		paramstypes.NewParamSetPair(KeyMaxNumPrivatePlans, &params.MaxNumPrivatePlans, validateMaxNumPrivatePlans),
	}
}

// Validate validates Params.
func (params Params) Validate() error {
	for _, field := range []struct {
		val          interface{}
		validateFunc func(i interface{}) error
	}{
		{params.PrivatePlanCreationFee, validatePrivatePlanCreationFee},
		{params.FeeCollector, validateFeeCollector},
		{params.MaxNumPrivatePlans, validateMaxNumPrivatePlans},
	} {
		if err := field.validateFunc(field.val); err != nil {
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
		return fmt.Errorf("invalid private plan creation fee: %w", err)
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

func validateMaxNumPrivatePlans(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
