package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func NewPoolResponse(pool Pool, poolState PoolState, balances sdk.Coins) PoolResponse {
	return PoolResponse{
		Id:                         pool.Id,
		MarketId:                   pool.MarketId,
		Balance0:                   sdk.NewCoin(pool.Denom0, balances.AmountOf(pool.Denom0)),
		Balance1:                   sdk.NewCoin(pool.Denom1, balances.AmountOf(pool.Denom1)),
		ReserveAddress:             pool.ReserveAddress,
		RewardsPool:                pool.RewardsPool,
		TickSpacing:                pool.TickSpacing,
		MinOrderQuantity:           pool.MinOrderQuantity,
		MinOrderQuote:              pool.MinOrderQuote,
		CurrentTick:                poolState.CurrentTick,
		CurrentSqrtPrice:           poolState.CurrentSqrtPrice,
		CurrentLiquidity:           poolState.CurrentLiquidity,
		TotalLiquidity:             poolState.TotalLiquidity,
		FeeGrowthGlobal:            poolState.FeeGrowthGlobal,
		FarmingRewardsGrowthGlobal: poolState.FarmingRewardsGrowthGlobal,
	}
}

func NewPositionResponse(position Position) PositionResponse {
	return PositionResponse{
		Id:                             position.Id,
		PoolId:                         position.PoolId,
		Owner:                          position.Owner,
		LowerPrice:                     exchangetypes.PriceAtTick(position.LowerTick),
		UpperPrice:                     exchangetypes.PriceAtTick(position.UpperTick),
		Liquidity:                      position.Liquidity,
		LastFeeGrowthInside:            position.LastFeeGrowthInside,
		OwedFee:                        position.OwedFee,
		LastFarmingRewardsGrowthInside: position.LastFarmingRewardsGrowthInside,
		OwedFarmingRewards:             position.OwedFarmingRewards,
	}
}

func NewTickInfoResponse(tick int32, tickInfo TickInfo) TickInfoResponse {
	return TickInfoResponse{
		Tick:                        tick,
		GrossLiquidity:              tickInfo.GrossLiquidity,
		NetLiquidity:                tickInfo.NetLiquidity,
		FeeGrowthOutside:            tickInfo.FeeGrowthOutside,
		FarmingRewardsGrowthOutside: tickInfo.FarmingRewardsGrowthOutside,
	}
}
