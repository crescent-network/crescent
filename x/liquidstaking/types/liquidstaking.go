package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate validates LiquidValidator.
func (lv LiquidValidator) Validate() error {
	_, valErr := sdk.ValAddressFromBech32(lv.OperatorAddress)
	if valErr != nil {
		return valErr
	}

	if lv.Weight.IsNil() {
		return fmt.Errorf("liquidstaking validator weight must not be nil")
	}

	if lv.Weight.IsNegative() {
		return fmt.Errorf("liquidstaking validator weight must not be negative: %s", lv.Weight)
	}

	// TODO: add validation for LiquidTokens, Status
	return nil
}
