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
		return market, sdkerrors.Wrap(err, "insufficient market creation fee")
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

func (k Keeper) EscrowCoins(
	ctx sdk.Context, market types.Market, addr sdk.AccAddress, amt ...sdk.Coin) error {
	coins := sdk.NewCoins(amt...)
	if coins.IsAllPositive() {
		return k.bankKeeper.SendCoins(ctx, addr, market.MustGetEscrowAddress(), coins)
	}
	return nil
}

func (k Keeper) ReleaseCoins(
	ctx sdk.Context, market types.Market, addr sdk.AccAddress, amt ...sdk.Coin) error {
	coins := sdk.NewCoins(amt...)
	if coins.IsAllPositive() {
		return k.bankKeeper.SendCoins(ctx, market.MustGetEscrowAddress(), addr, coins)
	}
	return nil
}
