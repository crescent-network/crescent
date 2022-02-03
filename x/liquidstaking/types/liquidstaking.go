package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
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

// TODO: consider changing to decimal
func (v LiquidValidator) GetDelShares(ctx sdk.Context, sk StakingKeeper) sdk.Int {
	del, found := sk.GetDelegation(ctx, LiquidStakingProxyAcc, v.GetOperator())
	if !found {
		return sdk.ZeroInt()
	}
	return del.GetShares().TruncateInt()
}

// LiquidValidators is a collection of LiquidValidator
type LiquidValidators []LiquidValidator

// MinMaxGap Return the list of LiquidValidator with the maximum gap and minimum gap from the target weight of LiquidValidators, respectively.
func (vs LiquidValidators) MinMaxGap(ctx sdk.Context, sk StakingKeeper, targetMap map[string]sdk.Int) (minGapVal LiquidValidator, maxGapVal LiquidValidator, amountNeeded sdk.Int) {
	maxGap := sdk.ZeroInt()
	minGap := sdk.ZeroInt()

	for _, val := range vs {
		target := targetMap[val.OperatorAddress]
		if val.GetDelShares(ctx, sk).Sub(target).GT(maxGap) {
			maxGap = val.GetDelShares(ctx, sk).Sub(target)
			maxGapVal = val
		}
		if val.GetDelShares(ctx, sk).Sub(target).LT(minGap) {
			minGap = val.GetDelShares(ctx, sk).Sub(target)
			minGapVal = val
		}
	}
	amountNeeded = sdk.MinInt(maxGap, minGap.Abs())

	return minGapVal, maxGapVal, amountNeeded
}

func (vs LiquidValidators) Len() int {
	return len(vs)
}

func (vs LiquidValidators) TotalWeight() sdk.Int {
	totalWeight := sdk.ZeroInt()
	for _, val := range vs {
		totalWeight = totalWeight.Add(val.Weight)
	}
	return totalWeight
}

func (vs LiquidValidators) TotalDelShares(ctx sdk.Context, sk StakingKeeper) sdk.Int {
	totalDelShares := sdk.ZeroInt()
	for _, val := range vs {
		totalDelShares = totalDelShares.Add(val.GetDelShares(ctx, sk))
	}
	return totalDelShares
}

// TODO: pointer map looks uncertainty, need to fix
func (vs LiquidValidators) Map() map[string]*LiquidValidator {
	valsMap := make(map[string]*LiquidValidator)
	for _, val := range vs {
		valsMap[val.OperatorAddress] = &val
	}
	return valsMap
}

// TODO: add testcodes with consider netAmount.TruncateDec() or not
// BTokenToNativeToken returns UnstakeAmount, NetAmount * BTokenAmount/TotalSupply * (1-UnstakeFeeRate)
func BTokenToNativeToken(bTokenAmount, bTokenTotalSupplyAmount sdk.Int, netAmount, feeRate sdk.Dec) (nativeTokenAmount sdk.Dec) {
	return netAmount.TruncateDec().Mul(bTokenAmount.ToDec().QuoTruncate(bTokenTotalSupplyAmount.ToDec())).Mul(sdk.OneDec().Sub(feeRate)).TruncateDec()
}

// mint btoken, MintAmount = TotalSupply * StakeAmount/NetAmount
func NativeTokenToBToken(nativeTokenAmount, bTokenTotalSupplyAmount sdk.Int, netAmount sdk.Dec) (bTokenAmount sdk.Int) {
	return bTokenTotalSupplyAmount.ToDec().Mul(nativeTokenAmount.ToDec()).QuoTruncate(netAmount.TruncateDec()).TruncateInt()
}

func MustMarshalLiquidValidator(cdc codec.BinaryCodec, val *LiquidValidator) []byte {
	return cdc.MustMarshal(val)
}

// must unmarshal a liquid validator from a store value
func MustUnmarshalLiquidValidator(cdc codec.BinaryCodec, value []byte) LiquidValidator {
	validator, err := UnmarshalLiquidValidator(cdc, value)
	if err != nil {
		panic(err)
	}

	return validator
}

// unmarshal a liquid validator from a store value
func UnmarshalLiquidValidator(cdc codec.BinaryCodec, value []byte) (val LiquidValidator, err error) {
	err = cdc.Unmarshal(value, &val)
	return val, err
}
