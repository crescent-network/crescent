package types

import (
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

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
