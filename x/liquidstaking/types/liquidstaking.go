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
	return val.TokensFromShares(delShares)
}

// TODO: add status dependency
func (v LiquidValidator) GetWeight(whitelistedValMap WhitelistedValMap) sdk.Int {
	if wv, ok := whitelistedValMap[v.OperatorAddress]; ok {
		return wv.TargetWeight
	} else {
		return sdk.ZeroInt()
	}
}

func (v LiquidValidator) GetStatus(ctx sdk.Context, sk StakingKeeper, whitelisted bool) ValidatorStatus {
	validator, found := sk.GetValidator(ctx, v.GetOperator())
	if !found {
		return ValidatorStatusInActive
	}
	active := ActiveCondition(validator, whitelisted)
	if v.OperatorAddress == validator.OperatorAddress && active {
		return ValidatorStatusActive
	} else {
		return ValidatorStatusInActive
	}
}

// ActiveCondition checks the liquid validator could be active by below cases
// active conditions
//- included on whitelist
//- existed valid validator on staking module ( existed, not nil del shares and tokens, valid exchange rate)
func ActiveCondition(validator stakingtypes.Validator, whitelisted bool) bool {
	return whitelisted &&
		// TODO: consider !validator.IsUnbonded(), explicit state checking not Unspecified
		validator.GetStatus() != stakingtypes.Unspecified &&
		!validator.GetTokens().IsNil() &&
		!validator.GetDelegatorShares().IsNil() &&
		!validator.InvalidExRate()
}

// LiquidValidators is a collection of LiquidValidator
type LiquidValidators []LiquidValidator

// MinMaxGap Return the list of LiquidValidator with the maximum gap and minimum gap from the target weight of LiquidValidators, respectively.
func (vs LiquidValidators) MinMaxGap(ctx sdk.Context, sk StakingKeeper, targetMap map[string]sdk.Int) (minGapVal LiquidValidator, maxGapVal LiquidValidator, amountNeeded sdk.Int) {
	maxGap := sdk.ZeroDec()
	minGap := sdk.ZeroDec()

	for _, val := range vs {
		target := targetMap[val.OperatorAddress]
		if val.GetDelShares(ctx, sk).Sub(target.ToDec()).GT(maxGap) {
			maxGap = val.GetDelShares(ctx, sk).Sub(target.ToDec())
			maxGapVal = val
		}
		if val.GetDelShares(ctx, sk).Sub(target.ToDec()).LT(minGap) {
			minGap = val.GetDelShares(ctx, sk).Sub(target.ToDec())
			minGapVal = val
		}
	}
	amountNeeded = sdk.MinInt(maxGap.TruncateInt(), minGap.TruncateInt().Abs())

	return minGapVal, maxGapVal, amountNeeded
}

func (vs LiquidValidators) Len() int {
	return len(vs)
}

func (vs LiquidValidators) TotalWeight(whitelistedValMap WhitelistedValMap) sdk.Int {
	totalWeight := sdk.ZeroInt()
	for _, val := range vs {
		totalWeight = totalWeight.Add(val.GetWeight(whitelistedValMap))
	}
	return totalWeight
}

func (vs LiquidValidators) TotalDelSharesAndLiquidTokens(ctx sdk.Context, sk StakingKeeper) (sdk.Dec, sdk.Dec) {
	totalDelShares := sdk.ZeroDec()
	totalLiquidTokens := sdk.ZeroDec()
	for _, lv := range vs {
		delShares := lv.GetDelShares(ctx, sk)
		totalDelShares = totalDelShares.Add(delShares)
		// TODO: kv optimizing
		liquidTokens := lv.GetLiquidTokens(ctx, sk)
		totalLiquidTokens = totalLiquidTokens.Add(liquidTokens)
	}
	return totalDelShares, totalLiquidTokens
}

// TODO: refactor to bool Map
func (vs LiquidValidators) Map() map[string]*LiquidValidator {
	valMap := make(map[string]*LiquidValidator)
	for _, val := range vs {
		valMap[val.OperatorAddress] = &val
	}
	return valMap
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
