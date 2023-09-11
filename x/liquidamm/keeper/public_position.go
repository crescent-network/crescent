package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func (k Keeper) CreatePublicPosition(
	ctx sdk.Context, poolId uint64, lowerPrice, upperPrice sdk.Dec,
	minBidAmt sdk.Int, feeRate sdk.Dec) (publicPosition types.PublicPosition, err error) {
	pool, found := k.ammKeeper.GetPool(ctx, poolId)
	if !found {
		return publicPosition, sdkerrors.Wrap(sdkerrors.ErrNotFound, "pool not found")
	}

	lowerTick := exchangetypes.TickAtPrice(lowerPrice)
	upperTick := exchangetypes.TickAtPrice(upperPrice)

	if found := k.LookupPublicPositionByParams(ctx, poolId, lowerTick, upperTick); found {
		return publicPosition, types.ErrPublicPositionExists
	}

	publicPositionId := k.GetNextPublicPositionIdWithUpdate(ctx)
	publicPosition = types.NewPublicPosition(
		publicPositionId, pool.Id, lowerTick, upperTick, minBidAmt, feeRate)
	k.SetPublicPosition(ctx, publicPosition)
	k.SetPublicPositionsByPoolIndex(ctx, publicPosition)
	k.SetPublicPositionByParamsIndex(ctx, publicPosition)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventPublicPositionCreated{
		PublicPositionId: publicPosition.Id,
		PoolId:           publicPosition.PoolId,
		LowerTick:        publicPosition.LowerTick,
		UpperTick:        publicPosition.UpperTick,
		MinBidAmount:     publicPosition.MinBidAmount,
		FeeRate:          publicPosition.FeeRate,
	}); err != nil {
		return publicPosition, err
	}
	return publicPosition, nil
}

func (k Keeper) MintShare(
	ctx sdk.Context, senderAddr sdk.AccAddress, publicPositionId uint64,
	desiredAmt sdk.Coins) (mintedShare sdk.Coin, position ammtypes.Position, liquidity sdk.Int, amt sdk.Coins, err error) {
	publicPosition, found := k.GetPublicPosition(ctx, publicPositionId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "public position not found")
		return
	}

	lowerPrice := exchangetypes.PriceAtTick(publicPosition.LowerTick)
	upperPrice := exchangetypes.PriceAtTick(publicPosition.UpperTick)
	position, liquidity, amt, err = k.ammKeeper.AddLiquidity(
		ctx, k.GetModuleAddress(), senderAddr, publicPosition.PoolId, lowerPrice, upperPrice, desiredAmt)
	if err != nil {
		return
	}

	shareDenom := types.ShareDenom(publicPositionId)
	shareSupply := k.bankKeeper.GetSupply(ctx, shareDenom).Amount
	mintedShareAmt := types.CalculateMintedShareAmount(
		liquidity, position.Liquidity.Sub(liquidity), shareSupply)
	mintedShare = sdk.NewCoin(shareDenom, mintedShareAmt)
	if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(mintedShare)); err != nil {
		return
	}
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, sdk.NewCoins(mintedShare)); err != nil {
		return
	}

	if err = ctx.EventManager().EmitTypedEvent(&types.EventMintShare{
		Minter:           senderAddr.String(),
		PublicPositionId: publicPositionId,
		MintedShare:      mintedShare,
		Liquidity:        liquidity,
		Amount:           amt,
	}); err != nil {
		return
	}
	return mintedShare, position, liquidity, amt, nil
}

// BurnShare handles types.MsgBurnShare to burn public position share.
func (k Keeper) BurnShare(
	ctx sdk.Context, senderAddr sdk.AccAddress, publicPositionId uint64,
	share sdk.Coin) (removedLiquidity sdk.Int, position ammtypes.Position, amt sdk.Coins, err error) {
	publicPosition, found := k.GetPublicPosition(ctx, publicPositionId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "public position not found")
		return
	}

	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(share)); err != nil {
		return
	}

	position = k.MustGetAMMPosition(ctx, publicPosition)

	shareSupply := k.bankKeeper.GetSupply(ctx, share.Denom).Amount
	var prevWinningBidShareAmt sdk.Int
	auction, found := k.GetPreviousRewardsAuction(ctx, publicPosition)
	if found && auction.WinningBid != nil {
		prevWinningBidShareAmt = auction.WinningBid.Share.Amount
	} else {
		prevWinningBidShareAmt = utils.ZeroInt
	}
	removedLiquidity = types.CalculateRemovedLiquidity(
		share.Amount, shareSupply, position.Liquidity, prevWinningBidShareAmt)

	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(share)); err != nil {
		return
	}
	_, amt, err = k.ammKeeper.RemoveLiquidity(
		ctx, k.GetModuleAddress(), senderAddr, position.Id, removedLiquidity)
	if err != nil {
		return
	}

	if err = ctx.EventManager().EmitTypedEvent(&types.EventBurnShare{
		Burner:           senderAddr.String(),
		PublicPositionId: publicPositionId,
		Share:            share,
		RemovedLiquidity: removedLiquidity,
		Amount:           amt,
	}); err != nil {
		return
	}
	return removedLiquidity, position, amt, nil
}

func (k Keeper) GetAMMPosition(ctx sdk.Context, publicPosition types.PublicPosition) (position ammtypes.Position, found bool) {
	return k.ammKeeper.GetPositionByParams(
		ctx, k.GetModuleAddress(), publicPosition.PoolId, publicPosition.LowerTick, publicPosition.UpperTick)
}

func (k Keeper) MustGetAMMPosition(ctx sdk.Context, publicPosition types.PublicPosition) ammtypes.Position {
	position, found := k.GetAMMPosition(ctx, publicPosition)
	if !found {
		panic("position not found")
	}
	return position
}
