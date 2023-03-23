package types

import (
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramstypes.ParamSet = (*Params)(nil)

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default params for the module.
func DefaultParams() Params {
	return Params{}
}

// ParamSetPairs implements ParamSet.
func (params *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{}
}

// Validate validates Params.
func (params Params) Validate() error {
	for _, field := range []struct {
		val          interface{}
		validateFunc func(i interface{}) error
	}{} {
		if err := field.validateFunc(field.val); err != nil {
			return err
		}
	}
	return nil
}
