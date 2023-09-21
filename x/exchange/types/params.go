package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

var _ paramstypes.ParamSet = (*Params)(nil)

var (
	KeyMarketCreationFee  = []byte("MarketCreationFee")
	KeyFees               = []byte("Fees")
	KeyMaxOrderLifespan   = []byte("MaxOrderLifespan")
	KeyMaxOrderPriceRatio = []byte("MaxOrderPriceRatio")
	KeyMaxSwapRoutesLen   = []byte("MaxSwapRoutesLen")
	KeyMaxNumMMOrders     = []byte("MaxNumMMOrders")
)

var (
	DefaultMarketCreationFee = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
	DefaultFees              = Fees{
		DefaultMakerFeeRate:        sdk.NewDecWithPrec(15, 4), // 0.15%
		DefaultTakerFeeRate:        sdk.NewDecWithPrec(3, 3),  // 0.3%
		DefaultOrderSourceFeeRatio: sdk.NewDecWithPrec(5, 1),  // 50%
	}
	DefaultMaxOrderLifespan          = 7 * 24 * time.Hour
	DefaultMaxOrderPriceRatio        = sdk.NewDecWithPrec(1, 1) // 10%
	DefaultMaxSwapRoutesLen   uint32 = 3
	DefaultMaxNumMMOrders     uint32 = 15

	MinPrice = sdk.NewDecWithPrec(1, 14)
	MaxPrice = sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 40))
)

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	marketCreationFee sdk.Coins, fees Fees, maxOrderLifespan time.Duration, maxOrderPriceRatio sdk.Dec,
	maxSwapRoutesLen, maxNumMMOrders uint32) Params {
	return Params{
		MarketCreationFee:  marketCreationFee,
		Fees:               fees,
		MaxOrderLifespan:   maxOrderLifespan,
		MaxOrderPriceRatio: maxOrderPriceRatio,
		MaxSwapRoutesLen:   maxSwapRoutesLen,
		MaxNumMMOrders:     maxNumMMOrders,
	}
}

// DefaultParams returns a default params for the module.
func DefaultParams() Params {
	return NewParams(
		DefaultMarketCreationFee, DefaultFees, DefaultMaxOrderLifespan, DefaultMaxOrderPriceRatio,
		DefaultMaxSwapRoutesLen, DefaultMaxNumMMOrders)
}

// ParamSetPairs implements ParamSet.
func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyMarketCreationFee, &params.MarketCreationFee, validateMarketCreationFee),
		paramstypes.NewParamSetPair(KeyFees, &params.Fees, validateFees),
		paramstypes.NewParamSetPair(KeyMaxOrderLifespan, &params.MaxOrderLifespan, validateMaxOrderLifespan),
		paramstypes.NewParamSetPair(KeyMaxOrderPriceRatio, &params.MaxOrderPriceRatio, validateMaxOrderPriceRatio),
		paramstypes.NewParamSetPair(KeyMaxSwapRoutesLen, &params.MaxSwapRoutesLen, validateMaxSwapRoutesLen),
		paramstypes.NewParamSetPair(KeyMaxNumMMOrders, &params.MaxNumMMOrders, validateMaxNumMMOrders),
	}
}

// Validate validates Params.
func (params Params) Validate() error {
	for _, field := range []struct {
		val          interface{}
		validateFunc func(i interface{}) error
	}{
		{params.MarketCreationFee, validateMarketCreationFee},
		{params.Fees, validateFees},
		{params.MaxOrderLifespan, validateMaxOrderLifespan},
		{params.MaxOrderPriceRatio, validateMaxOrderPriceRatio},
		{params.MaxSwapRoutesLen, validateMaxSwapRoutesLen},
		{params.MaxNumMMOrders, validateMaxNumMMOrders},
	} {
		if err := field.validateFunc(field.val); err != nil {
			return err
		}
	}
	return nil
}

func validateMarketCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid market creation fee: %w", err)
	}
	return nil
}

func validateFees(i interface{}) error {
	v, ok := i.(Fees)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return ValidateFees(v.DefaultMakerFeeRate, v.DefaultTakerFeeRate, v.DefaultOrderSourceFeeRatio)
}

func validateMaxOrderLifespan(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v < 0 {
		return fmt.Errorf("max order lifespan must not be negative: %v", v)
	}
	return nil
}

func validateMaxOrderPriceRatio(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !(v.IsPositive() && v.LT(utils.OneDec)) {
		return fmt.Errorf("max order price ratio must be in range (0.0, 1.0): %s", v)
	}
	return nil
}

func validateMaxSwapRoutesLen(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("max swap routes len must not be 0")
	}
	return nil
}

func validateMaxNumMMOrders(i interface{}) error {
	if _, ok := i.(uint32); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
