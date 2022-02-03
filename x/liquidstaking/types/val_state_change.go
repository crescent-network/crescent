package types

import (
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO: Deprecated Status management, Liquid validator Status should be calculated directly

//// UpdateStatus updates the location of the shares within a liquid validator
//// to reflect the new status
//func (v LiquidValidator) UpdateStatus(newStatus ValidatorStatus) LiquidValidator {
//	v.Status = newStatus
//	return v
//}
//
//// ActiveToDelisting change active liquid validator status to delisting on below cases
//// active -> delisting conditions
////- excluded from the whitelist
////- with low power, kicked out of the Top MaxValidators list.
////- jailed(unbonding, unbonded)
////- downtime slashing
////- double signing slashing ( tombstoned, infinite jail )
////- self-delegation condition failed (SelfDelegationBelowMinSelfDelegation)
//func (activeLiquidValidators LiquidValidators) ActiveToDelisting(valsMap map[string]stakingtypes.Validator, whitelistedValMap map[string]WhitelistedValidator) {
//	for _, lv := range activeLiquidValidators {
//		valStr := lv.GetOperator().String()
//		if lv.Status != ValidatorStatusActive {
//			continue
//		}
//		_, whitelisted := whitelistedValMap[lv.OperatorAddress]
//		// not whitelisted or jailed, unbonding, unbonded due to downtime, double-sign slashing, SelfDelegationBelowMinSelfDelegation
//		if !lv.ActiveCondition(valsMap[valStr], whitelisted) {
//			lv.UpdateStatus(ValidatorStatusDelisting)
//			fmt.Println("[delisting liquid validator]", valStr)
//		}
//	}
//}
//
//// TODO: check delisting -> delisted for mature redelegation queue
//func (vs LiquidValidators) DelistingToDelisted(valsMap map[string]stakingtypes.Validator) {
//	for _, lv := range vs {
//		valStr := lv.GetOperator().String()
//		// TODO: only check for unbonding
//		if lv.Status == ValidatorStatusDelisting && valsMap[valStr].IsUnbonded() {
//			lv.UpdateStatus(ValidatorStatusDelisted)
//			// TODO: consider conditions and set immediately
//			fmt.Println("[delisted liquid validator]", valStr)
//		}
//	}
//}

// ActiveCondition checks the liquid validator could be active by below cases
// active conditions
//- included on whitelist
//- existed valid validator on staking module ( existed, not nil del shares and tokens, valid exchange rate)
// TODO: add unit test case, consider refactoring to IsActive
func (lv LiquidValidator) ActiveCondition(validator stakingtypes.Validator, whitelisted bool) bool {
	return whitelisted &&
		// TODO: consider !validator.IsUnbonded(), explicit state checking not Unspecified
		validator.GetStatus() != stakingtypes.Unspecified &&
		!validator.GetTokens().IsNil() &&
		!validator.GetDelegatorShares().IsNil() &&
		!validator.InvalidExRate()
}
