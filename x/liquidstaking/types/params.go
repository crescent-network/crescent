package types

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	farmingtypes "github.com/crescent-network/crescent/v2/x/farming/types"
)

// Parameter store keys
var (
	KeyLiquidBondDenom        = []byte("LiquidBondDenom")
	KeyWhitelistedValidators  = []byte("WhitelistedValidators")
	KeyUnstakeFeeRate         = []byte("UnstakeFeeRate")
	KeyMinLiquidStakingAmount = []byte("MinLiquidStakingAmount")

	DefaultLiquidBondDenom = "bstake"

	// DefaultUnstakeFeeRate is the default Unstake Fee Rate.
	DefaultUnstakeFeeRate = sdk.NewDecWithPrec(1, 3) // "0.001000000000000000"

	// DefaultMinLiquidStakingAmount is the default minimum liquid staking amount.
	DefaultMinLiquidStakingAmount = sdk.NewInt(1000000)

	// Const variables

	// RebalancingTrigger if the maximum difference and needed each redelegation amount exceeds it, asset rebalacing will be executed.
	RebalancingTrigger = sdk.NewDecWithPrec(1, 3) // "0.001000000000000000"

	// RewardTrigger If the sum of balance and the upcoming rewards of LiquidStakingProxyAcc exceeds it, the reward is automatically withdrawn and re-stake according to the weights.
	RewardTrigger = sdk.NewDecWithPrec(1, 3) // "0.001000000000000000"

	// LiquidStakingProxyAcc is a proxy reserve account for delegation and undelegation.
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
		WhitelistedValidators:  []WhitelistedValidator{},
		LiquidBondDenom:        DefaultLiquidBondDenom,
		UnstakeFeeRate:         DefaultUnstakeFeeRate,
		MinLiquidStakingAmount: DefaultMinLiquidStakingAmount,
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyLiquidBondDenom, &p.LiquidBondDenom, validateLiquidBondDenom),
		paramstypes.NewParamSetPair(KeyWhitelistedValidators, &p.WhitelistedValidators, validateWhitelistedValidators),
		paramstypes.NewParamSetPair(KeyUnstakeFeeRate, &p.UnstakeFeeRate, validateUnstakeFeeRate),
		paramstypes.NewParamSetPair(KeyMinLiquidStakingAmount, &p.MinLiquidStakingAmount, validateMinLiquidStakingAmount),
	}
}

// String returns a human-readable string representation of the parameters.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func (p Params) WhitelistedValsMap() WhitelistedValsMap {
	return GetWhitelistedValsMap(p.WhitelistedValidators)
}

// Validate validates parameters.
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.LiquidBondDenom, validateLiquidBondDenom},
		{p.WhitelistedValidators, validateWhitelistedValidators},
		{p.UnstakeFeeRate, validateUnstakeFeeRate},
		{p.MinLiquidStakingAmount, validateMinLiquidStakingAmount},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

func validateLiquidBondDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("liquid bond denom cannot be blank")
	}

	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}
	return nil
}

// validateWhitelistedValidators validates liquidstaking validator and total weight.
func validateWhitelistedValidators(i interface{}) error {
	wvs, ok := i.([]WhitelistedValidator)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	valsMap := map[string]struct{}{}
	for _, wv := range wvs {
		_, valErr := sdk.ValAddressFromBech32(wv.ValidatorAddress)
		if valErr != nil {
			return valErr
		}

		if wv.TargetWeight.IsNil() {
			return fmt.Errorf("liquidstaking validator target weight must not be nil")
		}

		if !wv.TargetWeight.IsPositive() {
			return fmt.Errorf("liquidstaking validator target weight must be positive: %s", wv.TargetWeight)
		}

		if _, ok := valsMap[wv.ValidatorAddress]; ok {
			return fmt.Errorf("liquidstaking validator cannot be duplicated: %s", wv.ValidatorAddress)
		}
		valsMap[wv.ValidatorAddress] = struct{}{}
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

func validateMinLiquidStakingAmount(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("min liquid staking amount must not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("min liquid staking amount must not be negative: %s", v)
	}

	return nil
}
