package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type WhitelistedValsMap map[string]WhitelistedValidator

func (whitelistedValsMap WhitelistedValsMap) IsListed(operatorAddr string) bool {
	if _, ok := whitelistedValsMap[operatorAddr]; ok {
		return true
	} else {
		return false
	}
}

func GetWhitelistedValsMap(whitelistedValidators []WhitelistedValidator) WhitelistedValsMap {
	whitelistedValsMap := make(WhitelistedValsMap)
	for _, wv := range whitelistedValidators {
		whitelistedValsMap[wv.ValidatorAddress] = wv
	}
	return whitelistedValsMap
}

// Validate validates LiquidValidator.
func (v LiquidValidator) Validate() error {
	_, valErr := sdk.ValAddressFromBech32(v.OperatorAddress)
	if valErr != nil {
		return valErr
	}
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

func (v LiquidValidator) GetDelShares(ctx sdk.Context, sk StakingKeeper) sdk.Dec {
	del, found := sk.GetDelegation(ctx, LiquidStakingProxyAcc, v.GetOperator())
	if !found {
		return sdk.ZeroDec()
	}
	return del.GetShares()
}

func (v LiquidValidator) GetLiquidTokens(ctx sdk.Context, sk StakingKeeper, onlyBonded bool) sdk.Int {
	delShares := v.GetDelShares(ctx, sk)
	if !delShares.IsPositive() {
		return sdk.ZeroInt()
	}
	val := sk.Validator(ctx, v.GetOperator())
	if onlyBonded && !val.IsBonded() {
		return sdk.ZeroInt()
	}
	return val.TokensFromSharesTruncated(delShares).TruncateInt()
}

func (v LiquidValidator) GetWeight(whitelistedValsMap WhitelistedValsMap, active bool) sdk.Int {
	if wv, ok := whitelistedValsMap[v.OperatorAddress]; ok && active {
		return wv.TargetWeight
	} else {
		return sdk.ZeroInt()
	}
}

func (v LiquidValidator) GetStatus(activeCondition bool) ValidatorStatus {
	if activeCondition {
		return ValidatorStatusActive
	} else {
		return ValidatorStatusInactive
	}
}

// ActiveCondition checks the liquid validator could be active by below cases
//- included on whitelist
//- existed valid validator on staking module ( existed, not nil del shares and tokens, valid exchange rate)
//- not tombstoned
func ActiveCondition(validator stakingtypes.Validator, whitelisted bool, tombstoned bool) bool {
	return whitelisted &&
		!tombstoned &&
		// !Unspecified ==> Bonded, Unbonding, Unbonded
		validator.GetStatus() != stakingtypes.Unspecified &&
		!validator.GetTokens().IsNil() &&
		!validator.GetDelegatorShares().IsNil() &&
		!validator.InvalidExRate()
}

// LiquidValidators is a collection of LiquidValidator
type LiquidValidators []LiquidValidator
type ActiveLiquidValidators LiquidValidators

// MinMaxGap Return the list of LiquidValidator with the maximum gap and minimum gap from the target weight of LiquidValidators, respectively.
func (vs LiquidValidators) MinMaxGap(targetMap, liquidTokenMap map[string]sdk.Int) (minGapVal LiquidValidator, maxGapVal LiquidValidator, amountNeeded sdk.Int, lastRedelegation bool) {
	maxGap := sdk.ZeroInt()
	minGap := sdk.ZeroInt()

	for _, val := range vs {
		gap := liquidTokenMap[val.OperatorAddress].Sub(targetMap[val.OperatorAddress])
		if gap.GT(maxGap) {
			maxGap = gap
			maxGapVal = val
		}
		if gap.LT(minGap) {
			minGap = gap
			minGapVal = val
		}
	}
	amountNeeded = sdk.MinInt(maxGap, minGap.Abs())
	// lastRedelegation when maxGap validator's liquid token == amountNeeded for redelegation all delShares
	lastRedelegation = amountNeeded.IsPositive() &&
		!targetMap[maxGapVal.OperatorAddress].IsPositive() &&
		liquidTokenMap[maxGapVal.OperatorAddress].Equal(amountNeeded)

	return minGapVal, maxGapVal, amountNeeded, lastRedelegation
}

func (vs LiquidValidators) Len() int {
	return len(vs)
}

func (vs LiquidValidators) TotalLiquidTokens(ctx sdk.Context, sk StakingKeeper, onlyBonded bool) (sdk.Int, map[string]sdk.Int) {
	totalLiquidTokens := sdk.ZeroInt()
	liquidTokenMap := map[string]sdk.Int{}
	for _, lv := range vs {
		liquidTokens := lv.GetLiquidTokens(ctx, sk, onlyBonded)
		liquidTokenMap[lv.OperatorAddress] = liquidTokens
		totalLiquidTokens = totalLiquidTokens.Add(liquidTokens)
	}
	return totalLiquidTokens, liquidTokenMap
}

func (vs LiquidValidators) Map() map[string]struct{} {
	valsMap := map[string]struct{}{}
	for _, val := range vs {
		valsMap[val.OperatorAddress] = struct{}{}
	}
	return valsMap
}

func (avs ActiveLiquidValidators) Len() int {
	return LiquidValidators(avs).Len()
}

func (avs ActiveLiquidValidators) TotalActiveLiquidTokens(ctx sdk.Context, sk StakingKeeper, onlyBonded bool) (sdk.Int, map[string]sdk.Int) {
	return LiquidValidators(avs).TotalLiquidTokens(ctx, sk, onlyBonded)
}

// TotalWeight for active liquid validator
func (avs ActiveLiquidValidators) TotalWeight(whitelistedValsMap WhitelistedValsMap) sdk.Int {
	totalWeight := sdk.ZeroInt()
	for _, val := range avs {
		totalWeight = totalWeight.Add(val.GetWeight(whitelistedValsMap, true))
	}
	return totalWeight
}

// NativeTokenToBToken returns bTokenTotalSupply * nativeTokenAmount / netAmount
func NativeTokenToBToken(nativeTokenAmount, bTokenTotalSupplyAmount sdk.Int, netAmount sdk.Dec) (bTokenAmount sdk.Int) {
	return bTokenTotalSupplyAmount.ToDec().MulTruncate(nativeTokenAmount.ToDec()).QuoTruncate(netAmount.TruncateDec()).TruncateInt()
}

// BTokenToNativeToken returns bTokenAmount * netAmount / bTokenTotalSupply with truncations
func BTokenToNativeToken(bTokenAmount, bTokenTotalSupplyAmount sdk.Int, netAmount sdk.Dec) (nativeTokenAmount sdk.Dec) {
	return bTokenAmount.ToDec().MulTruncate(netAmount).Quo(bTokenTotalSupplyAmount.ToDec()).TruncateDec()
}

// DeductFeeRate returns Input * (1-FeeRate) with truncations
func DeductFeeRate(input sdk.Dec, feeRate sdk.Dec) (feeDeductedOutput sdk.Dec) {
	return input.MulTruncate(sdk.OneDec().Sub(feeRate)).TruncateDec()
}

func (nas NetAmountState) CalcNetAmount() sdk.Dec {
	return nas.ProxyAccBalance.Add(nas.TotalLiquidTokens).Add(nas.TotalUnbondingBalance).ToDec().Add(nas.TotalRemainingRewards)
}

func (nas NetAmountState) CalcMintRate() sdk.Dec {
	if nas.NetAmount.IsNil() || !nas.NetAmount.IsPositive() {
		return sdk.ZeroDec()
	}
	return nas.BtokenTotalSupply.ToDec().QuoTruncate(nas.NetAmount)
}

type LiquidValidatorStates []LiquidValidatorState

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
