package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

func (k Keeper) GetProxyAccBalance(ctx sdk.Context, proxyAcc sdk.AccAddress) (balance sdk.Int) {
	return k.bankKeeper.GetBalance(ctx, proxyAcc, k.stakingKeeper.BondDenom(ctx)).Amount
}

func (k Keeper) TryRedelegations(ctx sdk.Context, redelegations []types.Redelegation, liquidValsMap map[string]*types.LiquidValidator) (completionTime time.Time, err error) {
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
		liquidValsMap[re.SrcValidator.String()].LiquidTokens = liquidValsMap[re.SrcValidator.String()].LiquidTokens.Sub(shares.TruncateInt())
		liquidValsMap[re.DstValidator.String()].LiquidTokens = liquidValsMap[re.DstValidator.String()].LiquidTokens.Add(shares.TruncateInt())
		// set liquid token, with also changed status
		k.SetLiquidValidator(ctx, *liquidValsMap[re.SrcValidator.String()])
		k.SetLiquidValidator(ctx, *liquidValsMap[re.DstValidator.String()])
	}
	writeCache()
	return completionTime, nil
}
