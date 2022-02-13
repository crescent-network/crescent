package types

import (
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

func (v LiquidValidator) IsTombstoned(ctx sdk.Context, stakingKeeper StakingKeeper, slashingKeeper SlashingKeeper) bool {
	val, found := stakingKeeper.GetValidator(ctx, v.GetOperator())
	if !found {
		return false
	}
	consPk, err := val.ConsPubKey()
	if err != nil {
		return false
	}
	return slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(consPk.Address()))
}

func (v LiquidValidator) GetDelShares(ctx sdk.Context, sk StakingKeeper) sdk.Dec {
	del, found := sk.GetDelegation(ctx, LiquidStakingProxyAcc, v.GetOperator())
	if !found {
		return sdk.ZeroDec()
	}
	return del.GetShares()
}

func (v LiquidValidator) GetLiquidTokens(ctx sdk.Context, sk StakingKeeper) sdk.Int {
	delShares := v.GetDelShares(ctx, sk)
	if !delShares.IsPositive() {
		return sdk.ZeroInt()
	}
	val := sk.Validator(ctx, v.GetOperator())
	return val.TokensFromSharesTruncated(delShares).TruncateInt()
}

func (v LiquidValidator) GetWeight(whitelistedValMap WhitelistedValMap, active bool) sdk.Int {
	if wv, ok := whitelistedValMap[v.OperatorAddress]; ok && active {
		return wv.TargetWeight
	} else {
		return sdk.ZeroInt()
	}
}

// TODO: refactor unused receiver
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

// TODO: add test code type level
// MinMaxGap Return the list of LiquidValidator with the maximum gap and minimum gap from the target weight of LiquidValidators, respectively.
func (vs LiquidValidators) MinMaxGap(targetMap, liquidTokenMap map[string]sdk.Int, threshold sdk.Int) (minGapVal LiquidValidator, maxGapVal LiquidValidator, amountNeeded sdk.Int, lastRedelegation bool) {
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

func (vs LiquidValidators) TotalLiquidTokens(ctx sdk.Context, sk StakingKeeper) (sdk.Int, map[string]sdk.Int) {
	totalLiquidTokens := sdk.ZeroInt()
	liquidTokenMap := make(map[string]sdk.Int)
	for _, lv := range vs {
		liquidTokens := lv.GetLiquidTokens(ctx, sk)
		liquidTokenMap[lv.OperatorAddress] = liquidTokens
		totalLiquidTokens = totalLiquidTokens.Add(liquidTokens)
	}
	return totalLiquidTokens, liquidTokenMap
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

// TODO: assert ActiveLiquidValidators == TotalLiquidTokens, need to handle no liquid tokens of inactive liquid validator
func (avs ActiveLiquidValidators) TotalActiveLiquidTokens(ctx sdk.Context, sk StakingKeeper) (sdk.Int, map[string]sdk.Int) {
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

// NativeTokenToBToken returns nativeTokenAmount * bTokenTotalSupply / netAmount
func NativeTokenToBToken(nativeTokenAmount, bTokenTotalSupplyAmount sdk.Int, netAmount sdk.Dec) (bTokenAmount sdk.Int) {
	return bTokenTotalSupplyAmount.ToDec().MulTruncate(nativeTokenAmount.ToDec()).QuoTruncate(netAmount.TruncateDec()).TruncateInt()
}

// BTokenToNativeToken returns bTokenAmount * netAmount / bTokenTotalSupply with truncations
func BTokenToNativeToken(bTokenAmount, bTokenTotalSupplyAmount sdk.Int, netAmount sdk.Dec) (nativeTokenAmount sdk.Dec) {
	return bTokenAmount.ToDec().MulTruncate(netAmount).QuoTruncate(bTokenTotalSupplyAmount.ToDec()).TruncateDec()
}

// DeductFeeRate returns Input * (1-FeeRate) with truncations
func DeductFeeRate(input sdk.Dec, feeRate sdk.Dec) (feeDeductedOutput sdk.Dec) {
	return input.MulTruncate(sdk.OneDec().Sub(feeRate)).TruncateDec()
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
