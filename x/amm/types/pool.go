package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v5/types"
)

func DerivePoolReserveAddress(poolId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("PoolReserveAddress/%d", poolId)))
}

func DerivePoolRewardsPoolAddress(poolId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("PoolRewardsPool/%d", poolId)))
}

func NewPool(id uint64, marketId uint64, denom0, denom1 string, tickSpacing uint32) Pool {
	return Pool{
		Id:               id,
		MarketId:         marketId,
		Denom0:           denom0,
		Denom1:           denom1,
		ReserveAddress:   DerivePoolReserveAddress(id).String(),
		RewardsPool:      DerivePoolRewardsPoolAddress(id).String(),
		TickSpacing:      tickSpacing,
		MinOrderQuantity: DefaultMinOrderQuantity,
	}
}

func (pool Pool) MustGetReserveAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(pool.ReserveAddress)
}

func (pool Pool) MustGetRewardsPoolAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(pool.RewardsPool)
}

func (pool Pool) DenomIn(isBuy bool) string {
	if isBuy {
		return pool.Denom0
	}
	return pool.Denom1
}

func (pool Pool) DenomOut(isBuy bool) string {
	if isBuy {
		return pool.Denom1
	}
	return pool.Denom0
}

func (pool Pool) Validate() error {
	if pool.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if pool.MarketId == 0 {
		return fmt.Errorf("market id must not be 0")
	}
	if err := sdk.ValidateDenom(pool.Denom0); err != nil {
		return fmt.Errorf("invalid denom 0: %w", err)
	}
	if err := sdk.ValidateDenom(pool.Denom1); err != nil {
		return fmt.Errorf("invalid denom 1: %w", err)
	}
	if pool.Denom0 == pool.Denom1 {
		return fmt.Errorf("denom 0 and denom 1 must not be same: %s", pool.Denom0)
	}
	if _, err := sdk.AccAddressFromBech32(pool.ReserveAddress); err != nil {
		return fmt.Errorf("invalid reserve address: %w", err)
	}
	if _, err := sdk.AccAddressFromBech32(pool.RewardsPool); err != nil {
		return fmt.Errorf("invalid rewards pool: %w", err)
	}
	if !IsAllowedTickSpacing(pool.TickSpacing) {
		return fmt.Errorf("tick spacing %d is not allowed", pool.TickSpacing)
	}
	if pool.MinOrderQuantity.IsNegative() {
		return fmt.Errorf("min order quantity must not be negative: %s", pool.MinOrderQuantity)
	}
	return nil
}

func NewPoolState(tick int32, price sdk.Dec) PoolState {
	return PoolState{
		CurrentTick:                tick,
		CurrentPrice:               price,
		CurrentLiquidity:           utils.ZeroInt,
		TotalLiquidity:             utils.ZeroInt,
		FeeGrowthGlobal:            sdk.DecCoins{},
		FarmingRewardsGrowthGlobal: sdk.DecCoins{},
	}
}

func (poolState PoolState) Validate() error {
	if !poolState.CurrentPrice.IsPositive() {
		return fmt.Errorf("current price must be positive: %s", poolState.CurrentPrice)
	}
	if poolState.CurrentLiquidity.IsNegative() {
		return fmt.Errorf("current liquidity must not be negative: %s", poolState.CurrentLiquidity)
	}
	if poolState.TotalLiquidity.IsNegative() {
		return fmt.Errorf("total liquidity must not be negative: %s", poolState.TotalLiquidity)
	}
	if err := poolState.FeeGrowthGlobal.Validate(); err != nil {
		return fmt.Errorf("invalid fee growth global: %w", err)
	}
	if len(poolState.FeeGrowthGlobal) > 2 {
		return fmt.Errorf(
			"number of coins in fee growth global must not be higher than 2: %d", len(poolState.FeeGrowthGlobal))
	}
	if err := poolState.FarmingRewardsGrowthGlobal.Validate(); err != nil {
		return fmt.Errorf("invalid farming rewards growth global: %w", err)
	}
	return nil
}
