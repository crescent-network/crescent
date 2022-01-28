package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// UpdateStatus updates the location of the shares within a liquid validator
// to reflect the new status
func (v LiquidValidator) UpdateStatus(newStatus ValidatorStatus) LiquidValidator {
	v.Status = newStatus
	return v
}

// ActiveToDelisting change active liquid validator status to delisting on below cases
// active -> delisting conditions
//- excluded from the whitelist
//- with low power, kicked out of the Top MaxValidators list.
//- jailed(unbonding, unbonded)
//- downtime slashing
//- double signing slashing ( tombstoned, infinite jail )
//- self-delegation condition failed (SelfDelegationBelowMinSelfDelegation)
func (activeLiquidValidators LiquidValidators) ActiveToDelisting(valsMap map[string]stakingtypes.Validator, whitelistedValMap map[string]WhitelistedValidator, commissionRate sdk.Dec) {
	for _, lv := range activeLiquidValidators {
		valStr := lv.GetOperator().String()
		if lv.Status != ValidatorStatusActive {
			continue
		}
		_, whitelisted := whitelistedValMap[lv.OperatorAddress]
		// not whitelisted or jailed, unbonding, unbonded due to downtime, double-sign slashing, SelfDelegationBelowMinSelfDelegation
		if !lv.ActiveCondition(valsMap[valStr], whitelisted, commissionRate) {
			lv.UpdateStatus(ValidatorStatusDelisting)
			fmt.Println("[delisting liquid validator]", valStr)
		}
	}
}

// TODO: check delisting -> delisted for mature redelegation queue
func (vs LiquidValidators) DelistingToDelisted(valsMap map[string]stakingtypes.Validator) {
	for _, lv := range vs {
		valStr := lv.GetOperator().String()
		if lv.Status == ValidatorStatusDelisting && valsMap[valStr].IsUnbonded() {
			lv.UpdateStatus(ValidatorStatusDelisted)
			// TODO: consider conditions and set immediately
			fmt.Println("[delisted liquid validator]", valStr)
		}
	}
}

// ActiveCondition checks the liquid validator could be active by below cases
// active conditions
//- included on whitelist
//- included on the Top MaxValidators list.
//- not jailed(unbonding, unbonded)
//- not downtime slashing
//- not double signing slashing ( tombstoned, infinite jail )
//- not self-delegation condition failed (SelfDelegationBelowMinSelfDelegation)
func (lv LiquidValidator) ActiveCondition(validator stakingtypes.Validator, whitelisted bool, commissionRate sdk.Dec) bool {
	// whitelisted and not jailed, not unbonding, not unbonded due to downtime, double-sign slashing, match commissionRate
	return whitelisted &&
		!validator.IsJailed() &&
		!validator.IsUnbonding() &&
		!validator.IsUnbonded() // TODO: already unbonded case
}
