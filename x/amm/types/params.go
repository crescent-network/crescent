package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramstypes.ParamSet = (*Params)(nil)

var (
	KeyPoolCreationFee    = []byte("SpotMarketCreationFee")
	KeyDefaultTickSpacing = []byte("DefaultTickSpacing")
)

var (
	DefaultPoolCreationFee    = sdk.NewCoins()
	DefaultDefaultTickSpacing = uint32(50)
)

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default params for the module.
func DefaultParams() Params {
	return Params{
		PoolCreationFee:    DefaultPoolCreationFee,
		DefaultTickSpacing: DefaultDefaultTickSpacing,
	}
}

// ParamSetPairs implements ParamSet.
func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyPoolCreationFee, &params.PoolCreationFee, validatePoolCreationFee),
		paramstypes.NewParamSetPair(KeyDefaultTickSpacing, &params.DefaultTickSpacing, validateDefaultTickSpacing),
	}
}

// Validate validates Params.
func (params Params) Validate() error {
	for _, field := range []struct {
		val          interface{}
		validateFunc func(i interface{}) error
	}{
		{params.PoolCreationFee, validatePoolCreationFee},
		{params.DefaultTickSpacing, validateDefaultTickSpacing},
	} {
		if err := field.validateFunc(field.val); err != nil {
			return err
		}
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

func validateDefaultTickSpacing(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("tick spacing must be positive")
	}
	return nil
}
