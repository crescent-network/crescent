package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func HandleLiquidFarmCreateProposal(ctx sdk.Context, k Keeper, p *types.LiquidFarmCreateProposal) error {
	if _, err := k.CreateLiquidFarm(ctx, p.PoolId, p.LowerPrice, p.UpperPrice, p.MinBidAmount, p.FeeRate); err != nil {
		return err
	}
	return nil
}
