package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func NewPosition(id, poolId uint64, ownerAddr sdk.AccAddress, lowerTick, upperTick int32) Position {
	return Position{
		Id:                             id,
		PoolId:                         poolId,
		Owner:                          ownerAddr.String(),
		LowerTick:                      lowerTick,
		UpperTick:                      upperTick,
		Liquidity:                      utils.ZeroInt,
		LastFeeGrowthInside0:           utils.ZeroDec,
		LastFeeGrowthInside1:           utils.ZeroDec,
		OwedToken0:                     utils.ZeroInt,
		OwedToken1:                     utils.ZeroInt,
		LastFarmingRewardsGrowthInside: sdk.DecCoins{},
		OwedFarmingRewards:             sdk.Coins{},
	}
}

func (position Position) Validate() error {
	if position.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if position.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(position.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %w", err)
	}
	if position.Liquidity.IsNegative() {
		return fmt.Errorf("liquidity must not be negative: %s", position.Liquidity)
	}
	if position.LastFeeGrowthInside0.IsNegative() {
		return fmt.Errorf("last fee growth inside 0 must not be negative: %s", position.LastFeeGrowthInside0)
	}
	if position.LastFeeGrowthInside1.IsNegative() {
		return fmt.Errorf("last fee growth inside 1 must not be negative: %s", position.LastFeeGrowthInside1)
	}
	if position.OwedToken0.IsNegative() {
		return fmt.Errorf("owed token 0 must not be negative: %s", position.OwedToken0)
	}
	if position.OwedToken1.IsNegative() {
		return fmt.Errorf("owed token 1 must not be negative: %s", position.OwedToken1)
	}
	if err := position.LastFarmingRewardsGrowthInside.Validate(); err != nil {
		return fmt.Errorf("invalid last farming rewards growth inside: %w", err)
	}
	if err := position.OwedFarmingRewards.Validate(); err != nil {
		return fmt.Errorf("invalid owed farming rewards: %w", err)
	}
	return nil
}
