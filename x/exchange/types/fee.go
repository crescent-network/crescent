package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

// ValidateMakerTakerFeeRates validates maker fee rate and taker fee rate.
// ValidateMakerTakerFeeRates returns an error if taker fee rate is out of range [0, 1]
// or maker fee rate is out of range [-takerFeeRate, 1].
func ValidateMakerTakerFeeRates(makerFeeRate, takerFeeRate sdk.Dec) error {
	if takerFeeRate.GT(utils.OneDec) || takerFeeRate.IsNegative() {
		return fmt.Errorf("taker fee rate must be in range [0, 1]: %s", takerFeeRate)
	}
	minMakerFeeRate := takerFeeRate.Neg()
	if makerFeeRate.GT(utils.OneDec) || makerFeeRate.LT(minMakerFeeRate) {
		return fmt.Errorf(
			"maker fee rate must be in range [%s, 1]: %s", minMakerFeeRate, makerFeeRate)
	}
	return nil
}

// DeductFee returns coin amount after deducting fee along with the fee.
func DeductFee(amt, feeRate sdk.Dec) (deducted, fee sdk.Dec) {
	fee = feeRate.Mul(amt)
	deducted = amt.Sub(fee)
	return
}
