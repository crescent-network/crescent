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
	KeyMarketCreationFee       = []byte("MarketCreationFee")
	KeyFees                    = []byte("Fees")
	KeyMaxOrderLifespan        = []byte("MaxOrderLifespan")
	KeyMaxOrderPriceRatio      = []byte("MaxOrderPriceRatio")
	KeyDefaultMinOrderQuantity = []byte("DefaultMinOrderQuantity")
	KeyDefaultMinOrderQuote    = []byte("DefaultMinOrderQuote")
	KeyDefaultMaxOrderQuantity = []byte("DefaultMaxOrderQuantity")
	KeyDefaultMaxOrderQuote    = []byte("DefaultMaxOrderQuote")
	KeyMaxSwapRoutesLen        = []byte("MaxSwapRoutesLen")
	KeyMaxNumMMOrders          = []byte("MaxNumMMOrders")
)

var (
	DefaultMarketCreationFee = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
	DefaultFees              = Fees{
		DefaultMakerFeeRate:        sdk.NewDecWithPrec(15, 4), // 0.15%
		DefaultTakerFeeRate:        sdk.NewDecWithPrec(3, 3),  // 0.3%
		DefaultOrderSourceFeeRatio: sdk.NewDecWithPrec(5, 1),  // 50%
	}
	DefaultMaxOrderLifespan               = 7 * 24 * time.Hour
	DefaultMaxOrderPriceRatio             = sdk.NewDecWithPrec(1, 1) // 10%
	DefaultDefaultMinOrderQuantity        = sdk.NewInt(1)
	DefaultDefaultMinOrderQuote           = sdk.NewInt(1)
	DefaultDefaultMaxOrderQuantity        = sdk.NewIntWithDecimal(1, 40)
	DefaultDefaultMaxOrderQuote           = sdk.NewIntWithDecimal(1, 40)
	DefaultMaxSwapRoutesLen        uint32 = 3
	DefaultMaxNumMMOrders          uint32 = 15

	MinPrice = sdk.NewDecWithPrec(1, 14)                       // 10^-14
	MaxPrice = sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 24)) // 10^24
)

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	marketCreationFee sdk.Coins, fees Fees, maxOrderLifespan time.Duration, maxOrderPriceRatio sdk.Dec,
	defaultMinOrderQty, defaultMinOrderQuote, defaultMaxOrderQty, defaultMaxOrderQuote sdk.Int,
	maxSwapRoutesLen, maxNumMMOrders uint32) Params {
	return Params{
		MarketCreationFee:       marketCreationFee,
		Fees:                    fees,
		MaxOrderLifespan:        maxOrderLifespan,
		MaxOrderPriceRatio:      maxOrderPriceRatio,
		DefaultMinOrderQuantity: defaultMinOrderQty,
		DefaultMinOrderQuote:    defaultMinOrderQuote,
		DefaultMaxOrderQuantity: defaultMaxOrderQty,
		DefaultMaxOrderQuote:    defaultMaxOrderQuote,
		MaxSwapRoutesLen:        maxSwapRoutesLen,
		MaxNumMMOrders:          maxNumMMOrders,
	}
}

// DefaultParams returns a default params for the module.
func DefaultParams() Params {
	return NewParams(
		DefaultMarketCreationFee, DefaultFees, DefaultMaxOrderLifespan, DefaultMaxOrderPriceRatio,
		DefaultDefaultMinOrderQuantity, DefaultDefaultMinOrderQuote,
		DefaultDefaultMaxOrderQuantity, DefaultDefaultMaxOrderQuote,
		DefaultMaxSwapRoutesLen, DefaultMaxNumMMOrders)
}

// ParamSetPairs implements ParamSet.
func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyMarketCreationFee, &params.MarketCreationFee, validateMarketCreationFee),
		paramstypes.NewParamSetPair(KeyFees, &params.Fees, validateFees),
		paramstypes.NewParamSetPair(KeyMaxOrderLifespan, &params.MaxOrderLifespan, validateMaxOrderLifespan),
		paramstypes.NewParamSetPair(KeyMaxOrderPriceRatio, &params.MaxOrderPriceRatio, validateMaxOrderPriceRatio),
		paramstypes.NewParamSetPair(KeyDefaultMinOrderQuantity, &params.DefaultMinOrderQuantity, validateDefaultMinOrderQuantity),
		paramstypes.NewParamSetPair(KeyDefaultMinOrderQuote, &params.DefaultMinOrderQuote, validateDefaultMinOrderQuote),
		paramstypes.NewParamSetPair(KeyDefaultMaxOrderQuantity, &params.DefaultMaxOrderQuantity, validateDefaultMaxOrderQuantity),
		paramstypes.NewParamSetPair(KeyDefaultMaxOrderQuote, &params.DefaultMaxOrderQuote, validateDefaultMaxOrderQuote),
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
		{params.DefaultMinOrderQuantity, validateDefaultMinOrderQuantity},
		{params.DefaultMinOrderQuote, validateDefaultMinOrderQuote},
		{params.DefaultMaxOrderQuantity, validateDefaultMaxOrderQuantity},
		{params.DefaultMaxOrderQuote, validateDefaultMaxOrderQuote},
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

func validateDefaultMinOrderQuantity(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("default min order quantity must not be negative: %s", v)
	}
	return nil
}

func validateDefaultMinOrderQuote(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("default min order quote must not be negative: %s", v)
	}
	return nil
}

func validateDefaultMaxOrderQuantity(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("default max order quantity must not be negative: %s", v)
	}
	return nil
}

func validateDefaultMaxOrderQuote(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("default max order quote must not be negative: %s", v)
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
