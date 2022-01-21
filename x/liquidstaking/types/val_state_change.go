package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// UpdateStatus updates the location of the shares within a liquid validator
// to reflect the new status
// TODO: refactor to liquidvalidator.go or val_status_change.go, etc.
func (v LiquidValidator) UpdateStatus(newStatus ValidatorStatus) LiquidValidator {
	v.Status = newStatus
	return v
}

// ActiveToDelisting change active liquid validator status to delisting on below cases
// active -> delisting conditions
//- excluded from the whitelist
//- commission rate unmatched
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
		if !whitelisted ||
			valsMap[valStr].IsJailed() ||
			valsMap[valStr].IsUnbonding() ||
			valsMap[valStr].IsUnbonded() || // TODO: already unbonded case
			// TODO: whether to allow only the exact value or the lower value.
			// commission rate unmatched
			valsMap[valStr].Commission.Rate.GT(commissionRate) {
			lv.UpdateStatus(ValidatorStatusDelisting)
			fmt.Println("[delisting liquid validator]", valStr)
		}
		// TODO: consider add params.MinSelfDelegation condition
	}
}
