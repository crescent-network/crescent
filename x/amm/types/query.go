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
		TickSpacing:                pool.TickSpacing,
		ReserveAddress:             pool.ReserveAddress,
		CurrentTick:                poolState.CurrentTick,
		CurrentPrice:               poolState.CurrentPrice,
		CurrentLiquidity:           poolState.CurrentLiquidity,
		FeeGrowthGlobal:            poolState.FeeGrowthGlobal,
		FarmingRewardsGrowthGlobal: poolState.FarmingRewardsGrowthGlobal,
	}
}

func NewPositionResponse(position Position, pool Pool) PositionResponse {
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
