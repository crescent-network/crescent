package types

import (
	"fmt"
	"strings"

	farmingtypes "github.com/tendermint/farming/x/farming/types"
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyLiquidBondDenom       = []byte("LiquidBondDenom")
	KeyWhitelistedValidators = []byte("WhitelistedValidators")
	KeyUnstakeFeeRate        = []byte("UnstakeFeeRate")

	DefaultLiquidBondDenom = "bstake"
	// DefaultUnstakeFeeRate is the default Unstake Fee Rate.
	DefaultUnstakeFeeRate = sdk.NewDecWithPrec(1, 3) // "0.001000000000000000"
	// TODO: DefaultRebalancingThreshold

	// Const variables

	MinimumStakingAmount = sdk.NewInt(1000000)

	//LiquidStakingProxyAcc = farmingtypes.DeriveAddress(farmingtypes.AddressType20Bytes, ModuleName, "LiquidStakingProxyAcc")
	LiquidStakingProxyAcc = farmingtypes.DeriveAddress(farmingtypes.AddressType32Bytes, ModuleName, "LiquidStakingProxyAcc")
)

var _ paramstypes.ParamSet = (*Params)(nil)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the default liquidstaking module parameters.
func DefaultParams() Params {
	return Params{
		// TODO: btoken denom immutable
		LiquidBondDenom:       DefaultLiquidBondDenom,
		WhitelistedValidators: []WhitelistedValidator{},
		UnstakeFeeRate:        DefaultUnstakeFeeRate,
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyLiquidBondDenom, &p.LiquidBondDenom, ValidateLiquidBondDenom),
		paramstypes.NewParamSetPair(KeyWhitelistedValidators, &p.WhitelistedValidators, ValidateWhitelistedValidators),
		paramstypes.NewParamSetPair(KeyUnstakeFeeRate, &p.UnstakeFeeRate, validateUnstakeFeeRate),
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
		{p.LiquidBondDenom, ValidateLiquidBondDenom},
		{p.WhitelistedValidators, ValidateWhitelistedValidators},
		{p.UnstakeFeeRate, validateUnstakeFeeRate},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

// ValidateWhitelistedValidators validates liquidstaking validator and total weight.
func ValidateWhitelistedValidators(i interface{}) error {
	wvs, ok := i.([]WhitelistedValidator)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for _, wv := range wvs {
		_, valErr := sdk.ValAddressFromBech32(wv.ValidatorAddress)
		if valErr != nil {
			return valErr
		}

		if wv.Weight.IsNil() {
			return fmt.Errorf("liquidstaking validator weight must not be nil")
		}

		if wv.Weight.IsNegative() {
			return fmt.Errorf("liquidstaking validator weight must not be negative: %s", wv.Weight)
		}
	}
	return nil
}

func validateUnstakeFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("unstake fee rate must not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("unstake fee rate must not be negative: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("unstake fee rate too large: %s", v)
	}

	return nil
}

func ValidateLiquidBondDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("bond denom cannot be blank")
	}

	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}
	return nil
}
