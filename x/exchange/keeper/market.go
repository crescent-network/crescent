package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) CreateSpotMarket(ctx sdk.Context, creatorAddr sdk.AccAddress, baseDenom, quoteDenom string) (market types.SpotMarket, err error) {
	if !k.bankKeeper.HasSupply(ctx, baseDenom) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "base denom %s has no supply", baseDenom)
		return
	}
	if !k.bankKeeper.HasSupply(ctx, quoteDenom) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quote denom %s has no supply", quoteDenom)
		return
	}
	if _, found := k.GetSpotMarketByDenoms(ctx, baseDenom, quoteDenom); found {
		return market, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "spot market already exists")
	}

	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, k.GetSpotMarketCreationFee(ctx)); err != nil {
		return
	}

	marketId := k.GetNextSpotMarketIdWithUpdate(ctx)
	market = types.NewSpotMarket(marketId, baseDenom, quoteDenom)
	k.SetSpotMarket(ctx, market)
	k.SetSpotMarketByDenomsIndex(ctx, market)
	k.SetSpotMarketState(ctx, market.Id, types.NewSpotMarketState(nil))

	return market, nil
}

func (k Keeper) EscrowCoin(ctx sdk.Context, market types.SpotMarket, ordererAddr sdk.AccAddress, amt sdk.Coin) error {
	escrowAddr := sdk.MustAccAddressFromBech32(market.EscrowAddress)
	return k.bankKeeper.SendCoins(ctx, ordererAddr, escrowAddr, sdk.NewCoins(amt))
}

func (k Keeper) ReleaseCoin(ctx sdk.Context, market types.SpotMarket, ordererAddr sdk.AccAddress, amt sdk.Coin) error {
	escrowAddr := sdk.MustAccAddressFromBech32(market.EscrowAddress)
	return k.bankKeeper.SendCoins(ctx, escrowAddr, ordererAddr, sdk.NewCoins(amt))
}
