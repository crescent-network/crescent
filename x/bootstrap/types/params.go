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
	KeyCreationFeeRate             = []byte("CreationFeeRate")
	KeyProposerTradingFeeRate      = []byte("ProposerTradingFeeRate")
	KeyNonProposerTradingFeeRate   = []byte("NonProposerTradingFeeRate")
	KeyFeeFarmingRate              = []byte("FeeFarmingRate")
	KeyRequiredAmountReductionRate = []byte("RequiredAmountReductionRate")
	KeyInitialStageRequiredAmount  = []byte("InitialStageRequiredAmount")
	KeyTickPrecision               = []byte("TickPrecision")
	KeyOrderExtraGas               = []byte("OrderExtraGas")
	KeyVestingPeriods              = []byte("VestingPeriods")
	KeyQuoteCoinWhitelist          = []byte("QuoteCoinWhitelist")
	KeyFeeCollectorAddress         = []byte("FeeCollectorAddress")

	// TODO: need to fix default value
	DefaultCreationFeeRate             = sdk.MustNewDecFromStr("0.05")
	DefaultProposerTradingFeeRate      = sdk.MustNewDecFromStr("0.05")
	DefaultNonProposerTradingFeeRate   = sdk.MustNewDecFromStr("0.05")
	DefaultFeeFarmingRate              = sdk.MustNewDecFromStr("0.05")
	DefaultInitialStageRequiredAmount  = sdk.NewInt(10_000_000_000)
	DefaultRequiredAmountReductionRate = sdk.MustNewDecFromStr("0.5")
	DefaultTickPrecision               = uint32(3)
	DefaultFeeCollectorAddress         = farmingtypes.DeriveAddress(AddressType, ModuleName, "FeeCollector")
	DefaultVestingPeriods              = []int64{2592000, 2592000, 2592000}
	DefaultQuoteCoinWhitelist          = []string{sdk.DefaultBondDenom}
	DefaultOrderExtraGas               = sdk.Gas(37000)

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
		CreationFeeRate:             DefaultCreationFeeRate,
		ProposerTradingFeeRate:      DefaultProposerTradingFeeRate,
		NonProposerTradingFeeRate:   DefaultNonProposerTradingFeeRate,
		FeeFarmingRate:              DefaultFeeFarmingRate,
		RequiredAmountReductionRate: DefaultRequiredAmountReductionRate,
		InitialStageRequiredAmount:  DefaultInitialStageRequiredAmount,
		FeeCollectorAddress:         DefaultFeeCollectorAddress.String(),
		TickPrecision:               DefaultTickPrecision,
		OrderExtraGas:               DefaultOrderExtraGas,
		QuoteCoinWhitelist:          DefaultQuoteCoinWhitelist,
		VestingPeriods:              DefaultVestingPeriods,
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyCreationFeeRate, &p.CreationFeeRate, validateFeeRate),
		paramstypes.NewParamSetPair(KeyProposerTradingFeeRate, &p.ProposerTradingFeeRate, validateFeeRate),
		paramstypes.NewParamSetPair(KeyNonProposerTradingFeeRate, &p.NonProposerTradingFeeRate, validateFeeRate),
		paramstypes.NewParamSetPair(KeyFeeFarmingRate, &p.FeeFarmingRate, validateFeeRate),
		paramstypes.NewParamSetPair(KeyRequiredAmountReductionRate, &p.RequiredAmountReductionRate, validateRequiredAmountReductionRate),
		paramstypes.NewParamSetPair(KeyInitialStageRequiredAmount, &p.InitialStageRequiredAmount, validateInitialStageRequiredAmount),
		paramstypes.NewParamSetPair(KeyTickPrecision, &p.TickPrecision, validateTickPrecision),
		paramstypes.NewParamSetPair(KeyOrderExtraGas, &p.OrderExtraGas, validateExtraGas),
		paramstypes.NewParamSetPair(KeyVestingPeriods, &p.VestingPeriods, validateVestingPeriods),
		paramstypes.NewParamSetPair(KeyQuoteCoinWhitelist, &p.QuoteCoinWhitelist, validateQuoteCoinWhitelist),
		paramstypes.NewParamSetPair(KeyFeeCollectorAddress, &p.FeeCollectorAddress, validateFeeCollectorAddress),
	}
}

// Validate validates parameters.
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.CreationFeeRate, validateFeeRate},
		{p.ProposerTradingFeeRate, validateFeeRate},
		{p.NonProposerTradingFeeRate, validateFeeRate},
		{p.FeeFarmingRate, validateFeeRate},
		{p.RequiredAmountReductionRate, validateRequiredAmountReductionRate},
		{p.InitialStageRequiredAmount, validateInitialStageRequiredAmount},
		{p.TickPrecision, validateTickPrecision},
		{p.OrderExtraGas, validateExtraGas},
		{p.VestingPeriods, validateVestingPeriods},
		{p.QuoteCoinWhitelist, validateQuoteCoinWhitelist},
		{p.FeeCollectorAddress, validateFeeCollectorAddress},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

func validateFeeRate(i interface{}) error {
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

func validateQuoteCoinWhitelist(i interface{}) error {
	v, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, s := range v {
		if err := sdk.ValidateDenom(s); err != nil {
			return err
		}
	}

	return nil
}
