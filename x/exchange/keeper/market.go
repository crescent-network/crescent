package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) CreateMarket(ctx sdk.Context, creatorAddr sdk.AccAddress, baseDenom, quoteDenom string) (market types.Market, err error) {
	if !k.bankKeeper.HasSupply(ctx, baseDenom) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "base denom %s has no supply", baseDenom)
		return
	}
	if !k.bankKeeper.HasSupply(ctx, quoteDenom) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quote denom %s has no supply", quoteDenom)
		return
	}
	if _, found := k.GetMarketByDenoms(ctx, baseDenom, quoteDenom); found {
		return market, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, " market already exists")
	}

	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, k.GetMarketCreationFee(ctx)); err != nil {
		return
	}

	marketId := k.GetNextMarketIdWithUpdate(ctx)
	defaultMakerFeeRate := k.GetDefaultMakerFeeRate(ctx)
	defaultTakerFeeRate := k.GetDefaultTakerFeeRate(ctx)
	market = types.NewMarket(
		marketId, baseDenom, quoteDenom, defaultMakerFeeRate, defaultTakerFeeRate)
	k.SetMarket(ctx, market)
	k.SetMarketByDenomsIndex(ctx, market)
	k.SetMarketState(ctx, market.Id, types.NewMarketState(nil))

	return market, nil
}

func (k Keeper) EscrowCoin(ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress, amt sdk.Coin) error {
	escrowAddr := sdk.MustAccAddressFromBech32(market.EscrowAddress)
	return k.bankKeeper.SendCoins(ctx, ordererAddr, escrowAddr, sdk.NewCoins(amt))
}

func (k Keeper) ReleaseCoin(ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress, amt sdk.Coin) error {
	escrowAddr := sdk.MustAccAddressFromBech32(market.EscrowAddress)
	return k.bankKeeper.SendCoins(ctx, escrowAddr, ordererAddr, sdk.NewCoins(amt))
}
