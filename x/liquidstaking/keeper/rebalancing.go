package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

func (k Keeper) GetProxyAccBalance(ctx sdk.Context, proxyAcc sdk.AccAddress) (balance sdk.Int) {
	return k.bankKeeper.GetBalance(ctx, proxyAcc, k.stakingKeeper.BondDenom(ctx)).Amount
}

func (k Keeper) TryRedelegations(ctx sdk.Context, proxyAcc sdk.AccAddress, redelegations []types.Redelegation) (completionTime time.Time, err error) {
	cachedCtx, writeCache := ctx.CacheContext()
	for _, re := range redelegations {
		// TODO: ValidateUnbondAmount check
		shares, err := k.stakingKeeper.ValidateUnbondAmount(
			cachedCtx, re.Delegator, re.SrcValidator, re.Amount,
		)
		if err != nil {
			return time.Time{}, err
		}
		completionTime, err = k.stakingKeeper.BeginRedelegation(cachedCtx, re.Delegator, re.SrcValidator, re.DstValidator, shares)
		if err != nil {
			return time.Time{}, err
		}
	}
	writeCache()
	// TODO: bug on liquidValsMap pointer set, need to optimize UpdateLiquidTokens
	k.UpdateLiquidTokens(ctx, proxyAcc)
	return completionTime, nil
}
