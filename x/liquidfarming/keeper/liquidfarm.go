package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func (k Keeper) CreateLiquidFarm(
	ctx sdk.Context, poolId uint64, lowerPrice, upperPrice sdk.Dec,
	minBidAmt sdk.Int, feeRate sdk.Dec) (liquidFarm types.LiquidFarm, err error) {
	pool, found := k.ammKeeper.GetPool(ctx, poolId)
	if !found {
		return liquidFarm, sdkerrors.Wrap(sdkerrors.ErrNotFound, "pool not found")
	}

	lowerTick := exchangetypes.TickAtPrice(lowerPrice)
	upperTick := exchangetypes.TickAtPrice(upperPrice)

	liquidFarmId := k.GetNextLiquidFarmIdWithUpdate(ctx)
	liquidFarm = types.NewLiquidFarm(
		liquidFarmId, pool.Id, lowerTick, upperTick, minBidAmt, feeRate)
	k.SetLiquidFarm(ctx, liquidFarm)
	return liquidFarm, nil
}

func (k Keeper) MintShare(
	ctx sdk.Context, senderAddr sdk.AccAddress, liquidFarmId uint64,
	desiredAmt sdk.Coins) (mintedShare sdk.Coin, position ammtypes.Position, liquidity sdk.Int, amt sdk.Coins, err error) {
	liquidFarm, found := k.GetLiquidFarm(ctx, liquidFarmId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "liquid farm not found")
		return
	}

	moduleAccAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	lowerPrice := exchangetypes.PriceAtTick(liquidFarm.LowerTick)
	upperPrice := exchangetypes.PriceAtTick(liquidFarm.UpperTick)
	position, liquidity, amt, err = k.ammKeeper.AddLiquidity(
		ctx, moduleAccAddr, senderAddr, liquidFarm.PoolId, lowerPrice, upperPrice, desiredAmt)
	if err != nil {
		return
	}

	shareDenom := types.ShareDenom(liquidFarmId)
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

	// TODO: emit typed event
	//ctx.EventManager().EmitEvents(sdk.Events{
	//	sdk.NewEvent(
	//		types.EventTypeLiquidFarm,
	//		sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
	//		sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
	//		sdk.NewAttribute(types.AttributeKeyFarmingCoin, farmingCoin.String()),
	//		sdk.NewAttribute(types.AttributeKeyMintedCoin, mintedShare.String()),
	//		sdk.NewAttribute(types.AttributeKeyLiquidFarmReserveAddress, reserveAddr.String()),
	//	),
	//})

	return mintedShare, position, liquidity, amt, nil
}

// BurnShare handles types.MsgBurnShare to burn liquid farm share.
func (k Keeper) BurnShare(
	ctx sdk.Context, senderAddr sdk.AccAddress, liquidFarmId uint64,
	share sdk.Coin) (removedLiquidity sdk.Int, position ammtypes.Position, amt sdk.Coins, err error) {
	liquidFarm, found := k.GetLiquidFarm(ctx, liquidFarmId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "liquid farm not found")
		return
	}

	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(share)); err != nil {
		return
	}

	position = k.MustGetLiquidFarmPosition(ctx, liquidFarm)

	shareSupply := k.bankKeeper.GetSupply(ctx, share.Denom).Amount
	var prevWinningBidShareAmt sdk.Int
	auction, found := k.GetPreviousRewardsAuction(ctx, liquidFarm)
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
	moduleAccAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	_, amt, err = k.ammKeeper.RemoveLiquidity(
		ctx, moduleAccAddr, senderAddr, position.Id, removedLiquidity)
	if err != nil {
		return
	}

	// TODO: emit typed event
	//ctx.EventManager().EmitEvents(sdk.Events{
	//	sdk.NewEvent(
	//		types.EventTypeLiquidUnfarm,
	//		sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
	//		sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
	//		sdk.NewAttribute(types.AttributeKeyUnfarmingCoin, unfarmingCoin.String()),
	//		sdk.NewAttribute(types.AttributeKeyUnfarmedCoin, unfarmedCoin.String()),
	//	),
	//})

	return removedLiquidity, position, amt, nil
}

func (k Keeper) GetLiquidFarmPosition(ctx sdk.Context, liquidFarm types.LiquidFarm) (position ammtypes.Position, found bool) {
	moduleAccAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	return k.ammKeeper.GetPositionByParams(
		ctx, moduleAccAddr, liquidFarm.PoolId, liquidFarm.LowerTick, liquidFarm.UpperTick)
}

func (k Keeper) MustGetLiquidFarmPosition(ctx sdk.Context, liquidFarm types.LiquidFarm) ammtypes.Position {
	position, found := k.GetLiquidFarmPosition(ctx, liquidFarm)
	if !found {
		panic("position not found")
	}
	return position
}
