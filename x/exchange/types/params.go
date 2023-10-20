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
	KeyMarketCreationFee          = []byte("MarketCreationFee")
	KeyDefaultFees                = []byte("DefaultFees")
	KeyMaxOrderLifespan           = []byte("MaxOrderLifespan")
	KeyMaxOrderPriceRatio         = []byte("MaxOrderPriceRatio")
	KeyDefaultOrderQuantityLimits = []byte("DefaultOrderQuantityLimits")
	KeyDefaultOrderQuoteLimits    = []byte("DefaultOrderQuoteLimits")
	KeyMaxSwapRoutesLen           = []byte("MaxSwapRoutesLen")
	KeyMaxNumMMOrders             = []byte("MaxNumMMOrders")
)

var (
	DefaultMarketCreationFee = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
	DefaultFees              = Fees{
		MakerFeeRate:        sdk.NewDecWithPrec(15, 4), // 0.15%
		TakerFeeRate:        sdk.NewDecWithPrec(3, 3),  // 0.3%
		OrderSourceFeeRatio: sdk.NewDecWithPrec(5, 1),  // 50%
	}
	DefaultMaxOrderLifespan           = 7 * 24 * time.Hour
	DefaultMaxOrderPriceRatio         = sdk.NewDecWithPrec(1, 1) // 10%
	DefaultOrderQuantityLimits        = NewAmountLimits(sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30))
	DefaultOrderQuoteLimits           = NewAmountLimits(sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30))
	DefaultMaxSwapRoutesLen    uint32 = 3
	DefaultMaxNumMMOrders      uint32 = 15

	MinPrice = sdk.NewDecWithPrec(1, 14)                       // 10^-14
	MaxPrice = sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 24)) // 10^24
)

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	marketCreationFee sdk.Coins, defaultFees Fees, maxOrderLifespan time.Duration,
	maxOrderPriceRatio sdk.Dec, defaultOrderQtyLimits, defaultOrderQuoteLimits AmountLimits,
	maxSwapRoutesLen, maxNumMMOrders uint32) Params {
	return Params{
		MarketCreationFee:          marketCreationFee,
		DefaultFees:                defaultFees,
		MaxOrderLifespan:           maxOrderLifespan,
		MaxOrderPriceRatio:         maxOrderPriceRatio,
		DefaultOrderQuantityLimits: defaultOrderQtyLimits,
		DefaultOrderQuoteLimits:    defaultOrderQuoteLimits,
		MaxSwapRoutesLen:           maxSwapRoutesLen,
		MaxNumMMOrders:             maxNumMMOrders,
	}
}

// DefaultParams returns a default params for the module.
func DefaultParams() Params {
	return NewParams(
		DefaultMarketCreationFee, DefaultFees, DefaultMaxOrderLifespan, DefaultMaxOrderPriceRatio,
		DefaultOrderQuantityLimits, DefaultOrderQuoteLimits,
		DefaultMaxSwapRoutesLen, DefaultMaxNumMMOrders)
}

// ParamSetPairs implements ParamSet.
func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyMarketCreationFee, &params.MarketCreationFee, validateMarketCreationFee),
		paramstypes.NewParamSetPair(KeyDefaultFees, &params.DefaultFees, validateDefaultFees),
		paramstypes.NewParamSetPair(KeyMaxOrderLifespan, &params.MaxOrderLifespan, validateMaxOrderLifespan),
		paramstypes.NewParamSetPair(KeyMaxOrderPriceRatio, &params.MaxOrderPriceRatio, validateMaxOrderPriceRatio),
		paramstypes.NewParamSetPair(KeyDefaultOrderQuantityLimits, &params.DefaultOrderQuantityLimits, validateDefaultOrderQuantityLimits),
		paramstypes.NewParamSetPair(KeyDefaultOrderQuoteLimits, &params.DefaultOrderQuoteLimits, validateDefaultOrderQuoteLimits),
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
		{params.DefaultFees, validateDefaultFees},
		{params.MaxOrderLifespan, validateMaxOrderLifespan},
		{params.MaxOrderPriceRatio, validateMaxOrderPriceRatio},
		{params.DefaultOrderQuantityLimits, validateDefaultOrderQuantityLimits},
		{params.DefaultOrderQuoteLimits, validateDefaultOrderQuoteLimits},
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

func validateDefaultFees(i interface{}) error {
	v, ok := i.(Fees)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid default fees: %w", err)
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

func validateDefaultOrderQuantityLimits(i interface{}) error {
	v, ok := i.(AmountLimits)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid default order quantity limits: %w", err)
	}
	return nil
}

func validateDefaultOrderQuoteLimits(i interface{}) error {
	v, ok := i.(AmountLimits)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid default order quote limits: %w", err)
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

func NewFees(
	makerFeeRate, takerFeeRate, orderSourceFeeRatio sdk.Dec) Fees {
	return Fees{
		MakerFeeRate:        makerFeeRate,
		TakerFeeRate:        takerFeeRate,
		OrderSourceFeeRatio: orderSourceFeeRatio,
	}
}

// Validate validates Fees.
// Validate returns an error if any of fee params is out of range [0, 1].
func (fees Fees) Validate() error {
	if fees.MakerFeeRate.GT(utils.OneDec) || fees.MakerFeeRate.IsNegative() {
		return fmt.Errorf("maker fee rate must be in range [0, 1]: %s", fees.MakerFeeRate)
	}
	if fees.TakerFeeRate.GT(utils.OneDec) || fees.TakerFeeRate.IsNegative() {
		return fmt.Errorf("taker fee rate must be in range [0, 1]: %s", fees.TakerFeeRate)
	}
	if fees.OrderSourceFeeRatio.GT(utils.OneDec) || fees.OrderSourceFeeRatio.IsNegative() {
		return fmt.Errorf("order source fee ratio must be in range [0, 1]: %s", fees.OrderSourceFeeRatio)
	}
	return nil
}

func NewAmountLimits(min, max sdk.Int) AmountLimits {
	return AmountLimits{Min: min, Max: max}
}

func (limits AmountLimits) Validate() error {
	if !limits.Min.IsPositive() {
		return fmt.Errorf("the minimum value must be positive: %s", limits.Min)
	}
	if limits.Min.GT(limits.Max) {
		return fmt.Errorf(
			"the minimum value is greater than the maximum value: %s > %s", limits.Min, limits.Max)
	}
	return nil
}
