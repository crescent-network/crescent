package types

import (
	"fmt"
	"time"

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
	KeyCreationFeeRate                  = []byte("CreationFeeRate")
	KeyProtocolFeeRate                  = []byte("ProtocolFeeRate")
	KeyInitialTradingFeeRate            = []byte("InitialTradingFeeRate")
	KeyTradingFeeRate                   = []byte("TradingFeeRate")
	KeyRequiredVotingPowerReductionRate = []byte("RequiredVotingPowerReductionRate")
	KeyRequiredVotingPower              = []byte("RequiredVotingPower")
	KeyTickPrecision                    = []byte("TickPrecision")
	KeyOrderExtraGas                    = []byte("OrderExtraGas")
	KeyVestingPeriods                   = []byte("VestingPeriods")
	KeyQuoteCoinWhitelist               = []byte("QuoteCoinWhitelist")
	KeyProtocolFeeCollectorAddress      = []byte("ProtocolFeeCollectorAddress")

	// TODO: need to fix default value
	DefaultCreationFeeRate                  = sdk.ZeroDec()
	DefaultProtocolFeeRate                  = sdk.MustNewDecFromStr("0.05")
	DefaultInitialTradingFeeRate            = sdk.MustNewDecFromStr("0.05")
	DefaultTradingFeeRate                   = sdk.MustNewDecFromStr("0.25")
	DefaultRequiredVotingPower              = sdk.NewInt(10_000_000_000)
	DefaultRequiredVotingPowerReductionRate = sdk.MustNewDecFromStr("0.5")
	DefaultTickPrecision                    = uint32(3)
	DefaultProtocolFeeCollectorAddress      = farmingtypes.DeriveAddress(AddressType, ModuleName, "FeeCollector")
	DefaultVestingPeriods                   = []int64{2592000, 2592000, 2592000}
	DefaultQuoteCoinWhitelist               = []string{sdk.DefaultBondDenom}
	DefaultOrderExtraGas                    = sdk.Gas(37000)

	//SampleReserveAcc = farmingtypes.DeriveAddress(AddressType, ModuleName, SampleReserveAccName)
	//DepositReserveAcc            = sdk.AccAddress(crypto.AddressHash([]byte(ModuleName)))

	// TODO: TBD
	MinStageDuration = 2 * time.Hour
	MaxStageDuration = 30 * 24 * time.Hour
	MinNumOfStages   = uint32(1)
	MaxNumOfStages   = uint32(20)
	MaxInitialOrders = 10000
)

var _ paramstypes.ParamSet = (*Params)(nil)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the default bootstrap module parameters.
func DefaultParams() Params {
	return Params{
		CreationFeeRate:                  DefaultCreationFeeRate,
		ProtocolFeeRate:                  DefaultProtocolFeeRate,
		InitialTradingFeeRate:            DefaultInitialTradingFeeRate,
		TradingFeeRate:                   DefaultTradingFeeRate,
		RequiredVotingPowerReductionRate: DefaultRequiredVotingPowerReductionRate,
		RequiredVotingPower:              DefaultRequiredVotingPower,
		ProtocolFeeCollectorAddress:      DefaultProtocolFeeCollectorAddress.String(),
		TickPrecision:                    DefaultTickPrecision,
		OrderExtraGas:                    DefaultOrderExtraGas,
		QuoteCoinWhitelist:               DefaultQuoteCoinWhitelist,
		VestingPeriods:                   DefaultVestingPeriods,
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyCreationFeeRate, &p.CreationFeeRate, validateFeeRate),
		paramstypes.NewParamSetPair(KeyProtocolFeeRate, &p.ProtocolFeeRate, validateFeeRate),
		paramstypes.NewParamSetPair(KeyInitialTradingFeeRate, &p.InitialTradingFeeRate, validateFeeRate),
		paramstypes.NewParamSetPair(KeyTradingFeeRate, &p.TradingFeeRate, validateFeeRate),
		paramstypes.NewParamSetPair(KeyRequiredVotingPowerReductionRate, &p.RequiredVotingPowerReductionRate, validateRequiredVotingPowerReductionRate),
		paramstypes.NewParamSetPair(KeyRequiredVotingPower, &p.RequiredVotingPower, validateRequiredVotingPower),
		paramstypes.NewParamSetPair(KeyTickPrecision, &p.TickPrecision, validateTickPrecision),
		paramstypes.NewParamSetPair(KeyOrderExtraGas, &p.OrderExtraGas, validateExtraGas),
		paramstypes.NewParamSetPair(KeyVestingPeriods, &p.VestingPeriods, validateVestingPeriods),
		paramstypes.NewParamSetPair(KeyQuoteCoinWhitelist, &p.QuoteCoinWhitelist, validateQuoteCoinWhitelist),
		paramstypes.NewParamSetPair(KeyProtocolFeeCollectorAddress, &p.ProtocolFeeCollectorAddress, validateFeeCollectorAddress),
	}
}

// Validate validates parameters.
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.CreationFeeRate, validateFeeRate},
		{p.ProtocolFeeRate, validateFeeRate},
		{p.InitialTradingFeeRate, validateFeeRate},
		{p.TradingFeeRate, validateFeeRate},
		{p.RequiredVotingPowerReductionRate, validateRequiredVotingPowerReductionRate},
		{p.RequiredVotingPower, validateRequiredVotingPower},
		{p.TickPrecision, validateTickPrecision},
		{p.OrderExtraGas, validateExtraGas},
		{p.VestingPeriods, validateVestingPeriods},
		{p.QuoteCoinWhitelist, validateQuoteCoinWhitelist},
		{p.ProtocolFeeCollectorAddress, validateFeeCollectorAddress},
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
		return fmt.Errorf("fee rate must not be negative: %s", v)
	}

	return nil
}

func validateRequiredVotingPowerReductionRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("required voting power reduction rate must not be negative: %s", v)
	}

	return nil
}

func validateRequiredVotingPower(i interface{}) error {
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
