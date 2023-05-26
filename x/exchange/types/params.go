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
	KeyMarketCreationFee   = []byte("MarketCreationFee")
	KeyDefaultMakerFeeRate = []byte("DefaultMakerFeeRate")
	KeyDefaultTakerFeeRate = []byte("DefaultTakerFeeRate")
	KeyMaxOrderLifespan    = []byte("MaxOrderLifespan")
	KeyMaxOrderPriceRatio  = []byte("MaxOrderPriceRatio")
	KeyMaxSwapRoutesLen    = []byte("MaxSwapRoutesLen")
	KeyMaxNumMMOrders      = []byte("MaxNumMMOrders")
)

var (
	DefaultMarketCreationFee          = sdk.NewCoins()
	DefaultDefaultMakerFeeRate        = sdk.NewDecWithPrec(-15, 4) // -0.15%
	DefaultDefaultTakerFeeRate        = sdk.NewDecWithPrec(3, 3)   // 0.3%
	DefaultMaxOrderLifespan           = 24 * time.Hour
	DefaultMaxOrderPriceRatio         = sdk.NewDecWithPrec(1, 1) // 10%
	DefaultMaxSwapRoutesLen    uint32 = 3
	DefaultMaxNumMMOrders      uint32 = 15

	MinPrice = sdk.NewDecWithPrec(1, 14)
	MaxPrice = sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 40))
)

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default params for the module.
func DefaultParams() Params {
	return Params{
		MarketCreationFee:   DefaultMarketCreationFee,
		DefaultMakerFeeRate: DefaultDefaultMakerFeeRate,
		DefaultTakerFeeRate: DefaultDefaultTakerFeeRate,
		MaxOrderLifespan:    DefaultMaxOrderLifespan,
		MaxOrderPriceRatio:  DefaultMaxOrderPriceRatio,
		MaxSwapRoutesLen:    DefaultMaxSwapRoutesLen,
		MaxNumMMOrders:      DefaultMaxNumMMOrders,
	}
}

// ParamSetPairs implements ParamSet.
func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyMarketCreationFee, &params.MarketCreationFee, validateMarketCreationFee),
		paramstypes.NewParamSetPair(KeyDefaultMakerFeeRate, &params.DefaultMakerFeeRate, validateDefaultMakerFeeRate),
		paramstypes.NewParamSetPair(KeyDefaultTakerFeeRate, &params.DefaultTakerFeeRate, validateDefaultTakerFeeRate),
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
		{params.DefaultMakerFeeRate, validateDefaultMakerFeeRate},
		{params.DefaultTakerFeeRate, validateDefaultTakerFeeRate},
		{params.MaxOrderLifespan, validateMaxOrderLifespan},
		{params.MaxOrderPriceRatio, validateMaxOrderPriceRatio},
		{params.MaxSwapRoutesLen, validateMaxSwapRoutesLen},
		{params.MaxNumMMOrders, validateMaxNumMMOrders},
	} {
		if err := field.validateFunc(field.val); err != nil {
			return err
		}
	}
	if params.DefaultMakerFeeRate.IsNegative() && params.DefaultMakerFeeRate.Neg().GT(params.DefaultTakerFeeRate) {
		return fmt.Errorf("minus default maker fee rate must not be greater than default taker fee rate")
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

func validateDefaultMakerFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.GT(utils.OneDec) {
		return fmt.Errorf("default maker fee rate must not be greater than 1.0: %s", v)
	}
	if v.LT(utils.OneDec.Neg()) {
		return fmt.Errorf("default maker fee rate must not be less than -1.0: %s", v)
	}
	return nil
}

func validateDefaultTakerFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.GT(utils.OneDec) {
		return fmt.Errorf("default taker fee rate must not be greater than 1.0: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("default taker fee rate must not be negative: %s", v)
	}
	return nil
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
