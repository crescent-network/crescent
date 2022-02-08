package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type WhitelistedValMap map[string]WhitelistedValidator

func (whitelistedValMap WhitelistedValMap) IsListed(operatorAddr string) bool {
	if _, ok := whitelistedValMap[operatorAddr]; ok {
		return true
	} else {
		return false
	}
}

func GetWhitelistedValMap(whitelistedValidators []WhitelistedValidator) WhitelistedValMap {
	whitelistedValMap := make(WhitelistedValMap)
	for _, wv := range whitelistedValidators {
		whitelistedValMap[wv.ValidatorAddress] = wv
	}
	return whitelistedValMap
}

// Validate validates LiquidValidator.
func (v LiquidValidator) Validate() error {
	_, valErr := sdk.ValAddressFromBech32(v.OperatorAddress)
	if valErr != nil {
		return valErr
	}
	// TODO: add state level validate
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

// TODO: could be deprecated or reduce duplicated sk
func (v LiquidValidator) GetDelShares(ctx sdk.Context, sk StakingKeeper) sdk.Dec {
	del, found := sk.GetDelegation(ctx, LiquidStakingProxyAcc, v.GetOperator())
	if !found {
		return sdk.ZeroDec()
	}
	return del.GetShares()
}

func (v LiquidValidator) GetLiquidTokens(ctx sdk.Context, sk StakingKeeper) sdk.Dec {
	delShares := v.GetDelShares(ctx, sk)
	if !delShares.IsPositive() {
		return sdk.ZeroDec()
	}
	val := sk.Validator(ctx, v.GetOperator())
	return val.TokensFromSharesTruncated(delShares)
}

func (v LiquidValidator) GetWeight(whitelistedValMap WhitelistedValMap, active bool) sdk.Int {
	if wv, ok := whitelistedValMap[v.OperatorAddress]; ok && active {
		return wv.TargetWeight
	} else {
		return sdk.ZeroInt()
	}
}

// TODO: unused receiver refactor
func (v LiquidValidator) GetStatus(activeCondition bool) ValidatorStatus {
	if activeCondition {
		return ValidatorStatusActive
	} else {
		return ValidatorStatusInActive
	}
}

// ActiveCondition checks the liquid validator could be active by below cases
// active conditions
//- included on whitelist
//- existed valid validator on staking module ( existed, not nil del shares and tokens, valid exchange rate, not tombstoned)
func ActiveCondition(validator stakingtypes.Validator, whitelisted bool, tombstoned bool) bool {
	return whitelisted &&
		!tombstoned &&
		// TODO: consider !validator.IsUnbonded(), explicit state checking not Unspecified
		validator.GetStatus() != stakingtypes.Unspecified &&
		!validator.GetTokens().IsNil() &&
		!validator.GetDelegatorShares().IsNil() &&
		!validator.InvalidExRate()
}

// LiquidValidators is a collection of LiquidValidator
type LiquidValidators []LiquidValidator
type ActiveLiquidValidators LiquidValidators

// MinMaxGap Return the list of LiquidValidator with the maximum gap and minimum gap from the target weight of LiquidValidators, respectively.
func (vs LiquidValidators) MinMaxGap(ctx sdk.Context, sk StakingKeeper, targetMap map[string]sdk.Int, threshold sdk.Int) (minGapVal LiquidValidator, maxGapVal LiquidValidator, amountNeeded sdk.Int) {
	maxGap := sdk.ZeroInt()
	minGap := sdk.ZeroInt()

	for _, val := range vs {
		// TODO: liquidTokens or DelShares
		// TODO: 차이가 우선순위가아니라 0로 만드는게 우선순위로
		gap := val.GetLiquidTokens(ctx, sk).TruncateInt().Sub(targetMap[val.OperatorAddress])
		//gap := target.Sub(val.GetLiquidTokens(ctx, sk).TruncateInt())
		if gap.GT(maxGap) {
			maxGap = gap
			maxGapVal = val
		}
		if gap.LT(minGap) {
			minGap = gap
			minGapVal = val
		}
	}
	// TODO: consider when equal
	fmt.Println("-------")
	amountNeeded = sdk.MinInt(maxGap, minGap.Abs())
	lastRedelegation := amountNeeded.IsPositive() && maxGap.Sub(minGap.Abs()).LT(threshold) && !targetMap[maxGapVal.OperatorAddress].IsPositive()
	if lastRedelegation {
		// TODO: verify edge case
		fmt.Println("[@@@@ lastRedelegation]", threshold, maxGap.Sub(minGap.Abs()).LT(threshold) && !targetMap[maxGapVal.OperatorAddress].IsPositive(), maxGap)
		amountNeeded = maxGap
	}

	//if amountNeeded.IsPositive() {
	//	fmt.Println("[MINMAX EXCEPT]", maxGap.Sub(minGap.Abs()), maxGap.Sub(minGap.Abs()).LT(minGap.Abs()), maxGap)
	//	fmt.Println("[MINMAX EXCEPT2]", maxGap.Sub(minGap.Abs()).LT(minGap.Abs()) && targetMap[maxGapVal.OperatorAddress].IsZero(), maxGap)
	//	fmt.Println("[MINMAX EXCEPT3]", maxGap.Sub(minGap.Abs()).LT(threshold) && !targetMap[maxGapVal.OperatorAddress].IsPositive(), maxGap)
	//}

	//origin := sdk.MinInt(maxGap, minGap.Abs())
	//except := sdk.MinInt(maxGap, minGap).Abs()
	//if !sdk.MinInt(maxGap, minGap.Abs()).Equal(sdk.MinInt(maxGap, minGap).Abs()) {
	//	//amountNeeded = sdk.MinInt(maxGap, minGap)
	//	amountNeeded = maxGap
	//	fmt.Println("[MINMAX EXCEPT]", origin, except, amountNeeded)
	//}
	//amountNeeded = sdk.MaxInt(maxGap, minGap).Abs()
	fmt.Println("[MinMaxGap]", amountNeeded, minGapVal.OperatorAddress, minGap, minGapVal.GetLiquidTokens(ctx, sk), maxGapVal.OperatorAddress, maxGap, maxGapVal.GetLiquidTokens(ctx, sk))
	return minGapVal, maxGapVal, amountNeeded
}

func (vs LiquidValidators) Len() int {
	return len(vs)
}

func (vs LiquidValidators) TotalLiquidTokens(ctx sdk.Context, sk StakingKeeper) sdk.Dec {
	totalLiquidTokens := sdk.ZeroDec()
	for _, lv := range vs {
		liquidTokens := lv.GetLiquidTokens(ctx, sk)
		totalLiquidTokens = totalLiquidTokens.Add(liquidTokens)
	}
	return totalLiquidTokens
}

func (vs LiquidValidators) Map() map[string]bool {
	valMap := make(map[string]bool)
	for _, val := range vs {
		valMap[val.OperatorAddress] = true
	}
	return valMap
}

func (avs ActiveLiquidValidators) Len() int {
	return LiquidValidators(avs).Len()
}

func (avs ActiveLiquidValidators) TotalLiquidTokens(ctx sdk.Context, sk StakingKeeper) sdk.Dec {
	return LiquidValidators(avs).TotalLiquidTokens(ctx, sk)
}

// TotalWeight for active liquid validator
func (avs ActiveLiquidValidators) TotalWeight(whitelistedValMap WhitelistedValMap) sdk.Int {
	totalWeight := sdk.ZeroInt()
	for _, val := range avs {
		totalWeight = totalWeight.Add(val.GetWeight(whitelistedValMap, true))
	}
	return totalWeight
}

// BTokenToNativeToken returns UnstakeAmount, NetAmount * BTokenAmount/TotalSupply * (1-UnstakeFeeRate)
func BTokenToNativeToken(bTokenAmount, bTokenTotalSupplyAmount sdk.Int, netAmount, feeRate sdk.Dec) (nativeTokenAmount sdk.Dec) {
	return netAmount.MulTruncate(bTokenAmount.ToDec().QuoTruncate(bTokenTotalSupplyAmount.ToDec())).MulTruncate(sdk.OneDec().Sub(feeRate)).TruncateDec()
}

// mint btoken, MintAmount = TotalSupply * StakeAmount/NetAmount
func NativeTokenToBToken(nativeTokenAmount, bTokenTotalSupplyAmount sdk.Int, netAmount sdk.Dec) (bTokenAmount sdk.Int) {
	return bTokenTotalSupplyAmount.ToDec().MulTruncate(nativeTokenAmount.ToDec()).QuoTruncate(netAmount.TruncateDec()).TruncateInt()
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
