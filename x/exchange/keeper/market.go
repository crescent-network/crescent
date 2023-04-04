package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) CreateSpotMarket(ctx sdk.Context, creatorAddr sdk.AccAddress, baseDenom, quoteDenom string) (market types.SpotMarket, err error) {
	marketId := types.DeriveMarketId(baseDenom, quoteDenom)
	if _, found := k.GetSpotMarket(ctx, marketId); found {
		return market, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "spot market already exists")
	}

	// TODO: charge creation fees

	market = types.NewSpotMarket(baseDenom, quoteDenom)
	k.SetSpotMarket(ctx, market)

	return market, nil
}
