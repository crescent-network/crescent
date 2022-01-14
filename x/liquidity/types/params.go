package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	farmingtypes "github.com/crescent-network/crescent/x/farming/types"
)

const (
	PoolReserveAccPrefix   = "PoolReserveAcc"
	PairEscrowAddrPrefix   = "PairEscrowAddr"
	ModuleAddrNameSplitter = "|"
	AddressType            = farmingtypes.AddressType32Bytes
)

// TODO: sort keys
var (
	KeyInitialPoolCoinSupply   = []byte("InitialPoolCoinSupply")
	KeyBatchSize               = []byte("BatchSize")
	KeyTickPrecision           = []byte("TickPrecision")
	KeyMinInitialDepositAmount = []byte("MinInitialDepositAmount")
	KeyPoolCreationFee         = []byte("PoolCreationFee")
	KeyFeeCollectorAddress     = []byte("FeeCollectorAddress")
	KeyMaxPriceLimitRatio      = []byte("MaxPriceLimitRatio")
	KeySwapFeeRate             = []byte("SwapFeeRate")
	KeyWithdrawFeeRate         = []byte("WithdrawFeeRate")
	KeyMaxOrderLifespan        = []byte("MaxOrderLifespan")
)

var (
	DefaultInitialPoolCoinSupply          = sdk.NewInt(1_000_000_000_000)
	DefaultBatchSize               uint32 = 1
	DefaultTickPrecision           uint32 = 3
	DefaultMinInitialDepositAmount        = sdk.NewInt(1000000)
	DefaultPoolCreationFee                = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)))
	DefaultFeeCollectorAddress            = farmingtypes.DeriveAddress(AddressType, ModuleName, "FeeCollector")
	DefaultMaxPriceLimitRatio             = sdk.NewDecWithPrec(1, 1) // 10%
	DefaultSwapFeeRate                    = sdk.ZeroDec()
	DefaultWithdrawFeeRate                = sdk.ZeroDec()
	DefaultMaxOrderLifespan               = 24 * time.Hour
)

var (
	MinOfferCoinAmount = sdk.NewInt(100) // This value can be modified in the future

	GlobalEscrowAddr = farmingtypes.DeriveAddress(AddressType, ModuleName, "GlobalEscrow")
)

var _ paramstypes.ParamSet = (*Params)(nil)

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func DefaultParams() Params {
	return Params{
		InitialPoolCoinSupply:   DefaultInitialPoolCoinSupply,
		BatchSize:               DefaultBatchSize,
		TickPrecision:           DefaultTickPrecision,
		MinInitialDepositAmount: DefaultMinInitialDepositAmount,
		PoolCreationFee:         DefaultPoolCreationFee,
		FeeCollectorAddress:     DefaultFeeCollectorAddress.String(),
		MaxPriceLimitRatio:      DefaultMaxPriceLimitRatio,
		SwapFeeRate:             DefaultSwapFeeRate,
		WithdrawFeeRate:         DefaultWithdrawFeeRate,
		MaxOrderLifespan:        DefaultMaxOrderLifespan,
	}
}

func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyInitialPoolCoinSupply, &params.InitialPoolCoinSupply, validateInitialPoolCoinSupply),
		paramstypes.NewParamSetPair(KeyBatchSize, &params.BatchSize, validateBatchSize),
		paramstypes.NewParamSetPair(KeyTickPrecision, &params.TickPrecision, validateTickPrecision),
		paramstypes.NewParamSetPair(KeyMinInitialDepositAmount, &params.MinInitialDepositAmount, validateMinInitialDepositAmount),
		paramstypes.NewParamSetPair(KeyPoolCreationFee, &params.PoolCreationFee, validatePoolCreationFee),
		paramstypes.NewParamSetPair(KeyFeeCollectorAddress, &params.FeeCollectorAddress, validateFeeCollectorAddress),
		paramstypes.NewParamSetPair(KeyMaxPriceLimitRatio, &params.MaxPriceLimitRatio, validateMaxPriceLimitRatio),
		paramstypes.NewParamSetPair(KeySwapFeeRate, &params.SwapFeeRate, validateSwapFeeRate),
		paramstypes.NewParamSetPair(KeyWithdrawFeeRate, &params.WithdrawFeeRate, validateWithdrawFeeRate),
		paramstypes.NewParamSetPair(KeyMaxOrderLifespan, &params.MaxOrderLifespan, validateMaxOrderLifespan),
	}
}

func (params Params) Validate() error {
	for _, field := range []struct {
		val          interface{}
		validateFunc func(i interface{}) error
	}{
		{params.InitialPoolCoinSupply, validateInitialPoolCoinSupply},
		{params.BatchSize, validateBatchSize},
		{params.TickPrecision, validateTickPrecision},
		{params.MinInitialDepositAmount, validateMinInitialDepositAmount},
		{params.PoolCreationFee, validatePoolCreationFee},
		{params.FeeCollectorAddress, validateFeeCollectorAddress},
		{params.MaxPriceLimitRatio, validateMaxPriceLimitRatio},
		{params.SwapFeeRate, validateSwapFeeRate},
		{params.WithdrawFeeRate, validateWithdrawFeeRate},
		{params.MaxOrderLifespan, validateMaxOrderLifespan},
	} {
		if err := field.validateFunc(field.val); err != nil {
			return err
		}
	}
	return nil
}

func validateInitialPoolCoinSupply(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("initial pool coin supply must not be nil")
	}

	if !v.IsPositive() {
		return fmt.Errorf("initial pool coin supply must be positive: %s", v)
	}

	return nil
}

func validateBatchSize(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("batch size must be positive: %d", v)
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

func validateMinInitialDepositAmount(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("minimum initial deposit amount must not be negative: %s", v)
	}

	return nil
}

func validatePoolCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid pool creation fee: %w", err)
	}

	return nil
}

func validateFeeCollectorAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if _, err := sdk.AccAddressFromBech32(v); err != nil {
		return fmt.Errorf("invalid fee collector address: %w", err)
	}

	return nil
}

func validateMaxPriceLimitRatio(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max price limit ratio must not be negative: %s", v)
	}

	return nil
}

func validateSwapFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("swap fee rate must not be negative: %s", v)
	}

	return nil
}

func validateWithdrawFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("withdraw fee rate must not be negative: %s", v)
	}

	return nil
}

func validateMaxOrderLifespan(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 0 {
		return fmt.Errorf("max order lifespan must not be negative: %s", v)
	}

	return nil
}
