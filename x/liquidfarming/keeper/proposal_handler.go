package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func HandleLiquidFarmCreateProposal(ctx sdk.Context, k Keeper, p *types.LiquidFarmCreateProposal) error {
	liquidFarm, err := k.CreateLiquidFarm(ctx, p.PoolId, p.LowerPrice, p.UpperPrice, p.MinBidAmount, p.FeeRate)
	if err != nil {
		return err
	}
	if err := ctx.EventManager().EmitTypedEvent(&types.EventLiquidFarmCreated{
		LiquidFarmId: liquidFarm.Id,
		PoolId:       liquidFarm.PoolId,
		LowerTick:    liquidFarm.LowerTick,
		UpperTick:    liquidFarm.UpperTick,
		MinBidAmount: liquidFarm.MinBidAmount,
		FeeRate:      liquidFarm.FeeRate,
	}); err != nil {
		return err
	}
	return nil
}
