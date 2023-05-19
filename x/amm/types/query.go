package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func NewPoolResponse(pool Pool, poolState PoolState, balances sdk.Coins) PoolResponse {
	return PoolResponse{
		Id:               pool.Id,
		MarketId:         pool.MarketId,
		Balance0:         sdk.NewCoin(pool.Denom0, balances.AmountOf(pool.Denom0)),
		Balance1:         sdk.NewCoin(pool.Denom1, balances.AmountOf(pool.Denom1)),
		TickSpacing:      pool.TickSpacing,
		ReserveAddress:   pool.ReserveAddress,
		CurrentTick:      poolState.CurrentTick,
		CurrentPrice:     poolState.CurrentPrice,
		CurrentLiquidity: poolState.CurrentLiquidity,
		FeeGrowthGlobal: sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(pool.Denom0, poolState.FeeGrowthGlobal0),
			sdk.NewDecCoinFromDec(pool.Denom1, poolState.FeeGrowthGlobal1)),
		FarmingRewardsGrowthGlobal: poolState.FarmingRewardsGrowthGlobal,
	}
}

func NewPositionResponse(position Position, pool Pool) PositionResponse {
	return PositionResponse{
		Id:         position.Id,
		PoolId:     position.PoolId,
		Owner:      position.Owner,
		LowerPrice: exchangetypes.PriceAtTick(position.LowerTick),
		UpperPrice: exchangetypes.PriceAtTick(position.UpperTick),
		Liquidity:  position.Liquidity,
		LastFeeGrowthInside: sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(pool.Denom0, position.LastFeeGrowthInside0),
			sdk.NewDecCoinFromDec(pool.Denom1, position.LastFeeGrowthInside1)),
		OwedTokens: sdk.NewCoins(
			sdk.NewCoin(pool.Denom0, position.OwedToken0),
			sdk.NewCoin(pool.Denom1, position.OwedToken1)),
		LastFarmingRewardsGrowthInside: position.LastFarmingRewardsGrowthInside,
		OwedFarmingRewards:             position.OwedFarmingRewards,
	}
}
