package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	farmingtypes "github.com/crescent-network/crescent/v4/x/farming/types"
)

const (
	AddressType = farmingtypes.AddressType32Bytes
	//ClaimableIncentiveReserveAccName string = "ClaimableIncentiveReserveAcc"
)

// Parameter store keys
var (
	KeyBootstrapFeeRate            = []byte("BootstrapFeeRate")
	KeyInitialStageRequiredAmount  = []byte("InitialStageRequiredAmount")
	KeyRequiredAmountReductionRate = []byte("RequiredAmountReductionRate")
	KeyTickPrecision               = []byte("TickPrecision")
	KeyFeeCollectorAddress         = []byte("FeeCollectorAddress")
	KeyDustCollectorAddress        = []byte("DustCollectorAddress")
	KeyBootstrapCreationFee        = []byte("BootstrapCreationFee")
	KeyOrderExtraGas               = []byte("OrderExtraGas")

	DefaultBootstrapFeeRate            = sdk.MustNewDecFromStr("0.05")
	DefaultInitialStageRequiredAmount  = sdk.NewInt(10_000_000_000)
	DefaultRequiredAmountReductionRate = sdk.MustNewDecFromStr("0.5")
	DefaultTickPrecision               = uint32(3)

	// TODO: TBD
	DefaultBootstrapCreationFee = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))

	DefaultFeeCollectorAddress  = farmingtypes.DeriveAddress(AddressType, ModuleName, "FeeCollector")
	DefaultDustCollectorAddress = farmingtypes.DeriveAddress(AddressType, ModuleName, "DustCollector")

	// TODO: TBD
	DefaultOrderExtraGas = sdk.Gas(37000)

	//DefaultIncentiveBudgetAddress = farmingtypes.DeriveAddress(AddressType, farmingtypes.ModuleName, "ecosystem_incentive_mm")
	//ClaimableIncentiveReserveAcc = farmingtypes.DeriveAddress(AddressType, ModuleName, ClaimableIncentiveReserveAccName)
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
		BootstrapFeeRate:            DefaultBootstrapFeeRate,
		InitialStageRequiredAmount:  DefaultInitialStageRequiredAmount,
		RequiredAmountReductionRate: DefaultRequiredAmountReductionRate,
		TickPrecision:               DefaultTickPrecision,
		FeeCollectorAddress:         DefaultFeeCollectorAddress.String(),
		DustCollectorAddress:        DefaultDustCollectorAddress.String(),
		BootstrapCreationFee:        DefaultBootstrapCreationFee,
		OrderExtraGas:               DefaultOrderExtraGas,
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyBootstrapFeeRate, &p.BootstrapFeeRate, validateBootstrapFeeRate),
		paramstypes.NewParamSetPair(KeyInitialStageRequiredAmount, &p.InitialStageRequiredAmount, validateInitialStageRequiredAmount),
		paramstypes.NewParamSetPair(KeyRequiredAmountReductionRate, &p.RequiredAmountReductionRate, validateRequiredAmountReductionRate),
		paramstypes.NewParamSetPair(KeyTickPrecision, &p.TickPrecision, validateTickPrecision),
		paramstypes.NewParamSetPair(KeyFeeCollectorAddress, &p.FeeCollectorAddress, validateFeeCollectorAddress),
		paramstypes.NewParamSetPair(KeyDustCollectorAddress, &p.DustCollectorAddress, validateDustCollectorAddress),
		paramstypes.NewParamSetPair(KeyBootstrapCreationFee, &p.BootstrapCreationFee, validateBootstrapCreationFee),
		paramstypes.NewParamSetPair(KeyOrderExtraGas, &p.OrderExtraGas, validateExtraGas),
	}
}

//func (p Params) IncentiveBudgetAcc() sdk.AccAddress {
//	acc, _ := sdk.AccAddressFromBech32(p.IncentiveBudgetAddress)
//	return acc
//}

//func (p Params) IncentivePairsMap() map[uint64]IncentivePair {
//	iMap := make(map[uint64]IncentivePair)
//	for _, pair := range p.IncentivePairs {
//		iMap[pair.PairId] = pair
//	}
//	return iMap
//}

// String returns a human-readable string representation of the parameters.
//func (p Params) String() string {
//	out, _ := yaml.Marshal(p)
//	return string(out)
//}

// Validate validates parameters.
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.BootstrapFeeRate, validateBootstrapFeeRate},
		{p.InitialStageRequiredAmount, validateInitialStageRequiredAmount},
		{p.RequiredAmountReductionRate, validateRequiredAmountReductionRate},
		{p.TickPrecision, validateTickPrecision},
		{p.FeeCollectorAddress, validateFeeCollectorAddress},
		{p.DustCollectorAddress, validateDustCollectorAddress},
		{p.BootstrapCreationFee, validateBootstrapCreationFee},
		{p.OrderExtraGas, validateExtraGas},
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
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := v.Validate(); err != nil {
		return err
	}

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

func validateDustCollectorAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == "" {
		return fmt.Errorf("dust collector address must not be empty")
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid account address: %v", v)
	}

	return nil
}

func validateBootstrapCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid bootstrap creation fee: %w", err)
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
