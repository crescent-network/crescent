package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyPrivatePlanCreationFee = []byte("PrivatePlanCreationFee")
)

var (
	DefaultPrivatePlanCreationFee = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100_000_000)))
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
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyPrivatePlanCreationFee, &p.PrivatePlanCreationFee, validatePrivatePlanCreationFee),
	}
}

// String returns a human readable string representation of the parameters.
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

	if v.Empty() {
		return fmt.Errorf("plan creation fee must not be empty")
	}

	return nil
}
