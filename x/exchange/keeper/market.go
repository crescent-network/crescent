package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) CreateMarket(
	ctx sdk.Context, creatorAddr sdk.AccAddress, baseDenom, quoteDenom string) (market types.Market, err error) {
	if !k.bankKeeper.HasSupply(ctx, baseDenom) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "base denom %s has no supply", baseDenom)
		return
	}
	if !k.bankKeeper.HasSupply(ctx, quoteDenom) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quote denom %s has no supply", quoteDenom)
		return
	}
	if marketId, found := k.GetMarketIdByDenoms(ctx, baseDenom, quoteDenom); found {
		return market, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "market already exists: %d", marketId)
	}

	fees := k.GetFees(ctx)
	if err = k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, creatorAddr, types.ModuleName, fees.MarketCreationFee); err != nil {
		return
	}

	marketId := k.GetNextMarketIdWithUpdate(ctx)
	market = types.NewMarket(
		marketId, baseDenom, quoteDenom, fees.DefaultMakerFeeRate, fees.DefaultTakerFeeRate)
	k.SetMarket(ctx, market)
	k.SetMarketByDenomsIndex(ctx, market)
	k.SetMarketState(ctx, market.Id, types.NewMarketState(nil))

	if err = ctx.EventManager().EmitTypedEvent(&types.EventCreateMarket{
		Creator:  creatorAddr.String(),
		MarketId: marketId,
	}); err != nil {
		return
	}

	return market, nil
}

func (k Keeper) EscrowCoin(
	ctx sdk.Context, market types.Market, addr sdk.AccAddress, amt sdk.Coin) error {
	return k.EscrowCoins(ctx, market, addr, sdk.NewCoins(amt))
}

func (k Keeper) EscrowCoins(
	ctx sdk.Context, market types.Market, addr sdk.AccAddress, amt sdk.Coins) error {
	if amt.IsAllPositive() {
		return k.bankKeeper.SendCoins(ctx, addr, market.MustGetEscrowAddress(), amt)
	}
	return nil
}

func (k Keeper) ReleaseCoin(
	ctx sdk.Context, market types.Market, addr sdk.AccAddress, amt sdk.Coin) error {
	return k.ReleaseCoins(ctx, market, addr, sdk.NewCoins(amt))
}

func (k Keeper) ReleaseCoins(
	ctx sdk.Context, market types.Market, addr sdk.AccAddress, amt sdk.Coins) error {
	if amt.IsAllPositive() {
		return k.bankKeeper.SendCoins(ctx, market.MustGetEscrowAddress(), addr, amt)
	}
	return nil
}
