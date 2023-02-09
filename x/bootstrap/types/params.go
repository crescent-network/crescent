package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	farmingtypes "github.com/crescent-network/crescent/v4/x/farming/types"
)

const (
	AddressType = farmingtypes.AddressType32Bytes
	//SampleReserveAccName string = "SampleReserveAcc"
)

// Parameter store keys
var (
	KeyBootstrapCreationFeeRate    = []byte("BootstrapCreationFeeRate")
	KeyBootstrapFeeRate            = []byte("BootstrapFeeRate")
	KeyRequiredAmountReductionRate = []byte("RequiredAmountReductionRate")
	KeyInitialStageRequiredAmount  = []byte("InitialStageRequiredAmount")
	KeyFeeCollectorAddress         = []byte("FeeCollectorAddress")
	KeyTickPrecision               = []byte("TickPrecision")
	KeyOrderExtraGas               = []byte("OrderExtraGas")
	KeyVestingPeriods              = []byte("VestingPeriods")

	// TODO: need to fix default value
	DefaultBootstrapCreationFeeRate    = sdk.MustNewDecFromStr("0.05")
	DefaultBootstrapFeeRate            = sdk.MustNewDecFromStr("0.05")
	DefaultInitialStageRequiredAmount  = sdk.NewInt(10_000_000_000)
	DefaultRequiredAmountReductionRate = sdk.MustNewDecFromStr("0.5")
	DefaultTickPrecision               = uint32(3)

	// TODO: TBD
	//DefaultBootstrapCreationFee = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))

	DefaultFeeCollectorAddress = farmingtypes.DeriveAddress(AddressType, ModuleName, "FeeCollector")

	DefaultVestingPeriods = []int64{2592000, 2592000, 2592000}

	// TODO: TBD
	DefaultOrderExtraGas = sdk.Gas(37000)

	//SampleReserveAcc = farmingtypes.DeriveAddress(AddressType, ModuleName, SampleReserveAccName)
	//DepositReserveAcc            = sdk.AccAddress(crypto.AddressHash([]byte(ModuleName)))
)

var _ paramstypes.ParamSet = (*Params)(nil)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the default bootstrap module parameters.
func DefaultParams() Params {
	return Params{
		BootstrapCreationFeeRate:    DefaultBootstrapCreationFeeRate,
		BootstrapFeeRate:            DefaultBootstrapFeeRate,
		RequiredAmountReductionRate: DefaultRequiredAmountReductionRate,
		InitialStageRequiredAmount:  DefaultInitialStageRequiredAmount,
		FeeCollectorAddress:         DefaultFeeCollectorAddress.String(),
		TickPrecision:               DefaultTickPrecision,
		OrderExtraGas:               DefaultOrderExtraGas,
		VestingPeriods:              DefaultVestingPeriods,
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyBootstrapCreationFeeRate, &p.BootstrapFeeRate, validateBootstrapFeeRate),
		paramstypes.NewParamSetPair(KeyBootstrapFeeRate, &p.BootstrapFeeRate, validateBootstrapFeeRate),
		paramstypes.NewParamSetPair(KeyRequiredAmountReductionRate, &p.RequiredAmountReductionRate, validateRequiredAmountReductionRate),
		paramstypes.NewParamSetPair(KeyInitialStageRequiredAmount, &p.InitialStageRequiredAmount, validateInitialStageRequiredAmount),
		paramstypes.NewParamSetPair(KeyFeeCollectorAddress, &p.FeeCollectorAddress, validateFeeCollectorAddress),
		paramstypes.NewParamSetPair(KeyTickPrecision, &p.TickPrecision, validateTickPrecision),
		paramstypes.NewParamSetPair(KeyOrderExtraGas, &p.OrderExtraGas, validateExtraGas),
		paramstypes.NewParamSetPair(KeyVestingPeriods, &p.VestingPeriods, validateVestingPeriods),
	}
}

// Validate validates parameters.
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.BootstrapCreationFeeRate, validateBootstrapFeeRate},
		{p.BootstrapFeeRate, validateBootstrapFeeRate},
		{p.RequiredAmountReductionRate, validateRequiredAmountReductionRate},
		{p.InitialStageRequiredAmount, validateInitialStageRequiredAmount},
		{p.FeeCollectorAddress, validateFeeCollectorAddress},
		{p.TickPrecision, validateTickPrecision},
		{p.OrderExtraGas, validateExtraGas},
		{p.VestingPeriods, validateVestingPeriods},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

func validateBootstrapFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("bootstrap fee rate must not be negative: %s", v)
	}

	return nil
}

func validateRequiredAmountReductionRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("required amount reduction fee rate must not be negative: %s", v)
	}

	return nil
}

func validateInitialStageRequiredAmount(i interface{}) error {
	_, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// TODO: sdk.Int or sdk.Coin
	//if err := v.Validate(); err != nil {
	//	return err
	//}

	return nil
}

func validateTickPrecision(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateFeeCollectorAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == "" {
		return fmt.Errorf("fee collector address must not be empty")
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid account address: %v", v)
	}

	return nil
}

func validateExtraGas(i interface{}) error {
	_, ok := i.(sdk.Gas)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateVestingPeriods(i interface{}) error {
	v, ok := i.([]int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, p := range v {
		if 60 > p {
			return fmt.Errorf("vesting period length should be over 60: %d", p)
		}
	}

	return nil
}

//func validateBootstrapCreationFee(i interface{}) error {
//	v, ok := i.(sdk.Coins)
//	if !ok {
//		return fmt.Errorf("invalid parameter type: %T", i)
//	}
//
//	if err := v.Validate(); err != nil {
//		return fmt.Errorf("invalid bootstrap creation fee: %w", err)
//	}
//
//	return nil
//}
