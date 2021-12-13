package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate validates LiquidValidator.
func (v LiquidValidator) Validate() error {
	_, valErr := sdk.ValAddressFromBech32(v.OperatorAddress)
	if valErr != nil {
		return valErr
	}

	if v.Weight.IsNil() {
		return fmt.Errorf("liquidstaking validator weight must not be nil")
	}

	if v.Weight.IsNegative() {
		return fmt.Errorf("liquidstaking validator weight must not be negative: %s", v.Weight)
	}

	// TODO: add validation for LiquidTokens, Status
	return nil
}

func (v LiquidValidator) GetOperator() sdk.ValAddress {
	if v.OperatorAddress == "" {
		return nil
	}
	addr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// LiquidValidators is a collection of LiquidValidator
type LiquidValidators []LiquidValidator

// TODO: Unimplemented, MinMax Return the list of LiquidValidator with the maximum and minimum values of LiquidTokens, respectively.
func (vs LiquidValidators) MinMax() (minVals LiquidValidators, maxVals LiquidValidators) {

	return
}
