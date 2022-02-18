package keeper // noalias

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// RandomValidator returns a random validator given access to the keeper and ctx
func RandomValidator(r *rand.Rand, keeper types.StakingKeeper, ctx sdk.Context) (val stakingtypes.Validator, ok bool) {
	vals := keeper.GetAllValidators(ctx)
	if len(vals) == 0 {
		return stakingtypes.Validator{}, false
	}

	i := r.Intn(len(vals))

	return vals[i], true
}

// RandomValidator returns a random validator given access to the keeper and ctx
func RandomActiveLiquidValidator(r *rand.Rand, ctx sdk.Context, k Keeper, sk types.StakingKeeper) (val stakingtypes.Validator, ok bool) {
	avs := k.GetActiveLiquidValidators(ctx, k.GetParams(ctx).WhitelistedValMap())
	if len(avs) == 0 {
		return stakingtypes.Validator{}, false
	}
	i := r.Intn(len(avs))

	val, found := sk.GetValidator(ctx, avs[i].GetOperator())
	return val, found
}
