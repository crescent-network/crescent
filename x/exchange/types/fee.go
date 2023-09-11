package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func NewFees(
	defaultMakerFeeRate, defaultTakerFeeRate, defaultOrderSourceFeeRatio sdk.Dec) Fees {
	return Fees{
		DefaultMakerFeeRate:        defaultMakerFeeRate,
		DefaultTakerFeeRate:        defaultTakerFeeRate,
		DefaultOrderSourceFeeRatio: defaultOrderSourceFeeRatio,
	}
}

// ValidateFees validates maker fee rate, taker fee rate and order source fee ratio.
// ValidateFees returns an error if any of fee params is out of range [0, 1].
func ValidateFees(makerFeeRate, takerFeeRate, orderSourceFeeRatio sdk.Dec) error {
	if makerFeeRate.GT(utils.OneDec) || makerFeeRate.IsNegative() {
		return fmt.Errorf("maker fee rate must be in range [0, 1]: %s", makerFeeRate)
	}
	if takerFeeRate.GT(utils.OneDec) || takerFeeRate.IsNegative() {
		return fmt.Errorf("taker fee rate must be in range [0, 1]: %s", takerFeeRate)
	}
	if orderSourceFeeRatio.GT(utils.OneDec) || orderSourceFeeRatio.IsNegative() {
		return fmt.Errorf("order source fee ratio must be in range [0, 1]: %s", orderSourceFeeRatio)
	}
	return nil
}

// DeductFee returns coin amount after deducting fee along with the fee.
func DeductFee(amt, feeRate sdk.Dec) (deducted, fee sdk.Dec) {
	fee = feeRate.Mul(amt)
	deducted = amt.Sub(fee)
	return
}
