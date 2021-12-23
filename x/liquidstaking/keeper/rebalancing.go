package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/farming/x/liquidstaking/types"
)

func (k Keeper) UpdateLiquidValidators(ctx sdk.Context) {
	//liquidVals := k.GetAllLiquidValidators()
	// TODO: GET, SET, GETALL, Iterate LiquidValidators, indexing
	//for _, val := range liquidVals {
	//	if val.
	//}
	//k.stakingKeeper.GetLastTotalPower()
}

// activeVals containing ValidatorStatusWhiteListed which is containing just added on whitelist(power 0) and ValidatorStatusDelisting
func (k Keeper) Rebalancing(ctx sdk.Context, moduleAcc sdk.AccAddress, activeVals types.LiquidValidators, delistingVals types.LiquidValidators, threshold sdk.Dec) (rebalancedLiquidVals types.LiquidValidators) {
	if len(delistingVals) != 0 {
		// delisting logics
	}

	//maxVals, minVals := activeVals.MinMax()
	//for _, val := range activeVals {
	//	//TODO: add rebalancing logic
	//
	//
	//	//val.Weight
	//	//val.OperatorAddress
	//	//val.LiquidTokens
	//	//val.Status
	//
	//
	//	// //example of rebalancing
	//	// k.stakingKeeper.BeginRedelegation(ctx, moduleAcc, val.GetOperator(), liquidVals[1].GetOperator(), sdk.NewDec(1000))
	//
	//}
	return
}
