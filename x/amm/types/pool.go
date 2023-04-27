package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v5/types"
)

func NewPool(id uint64, marketId uint64, denom0, denom1 string, tickSpacing uint32) Pool {
	return Pool{
		Id:             id,
		MarketId:       marketId,
		Denom0:         denom0,
		Denom1:         denom1,
		TickSpacing:    tickSpacing,
		ReserveAddress: DerivePoolReserveAddress(id).String(),
	}
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
	if pool.TickSpacing == 0 {
		return fmt.Errorf("tick spacing must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(pool.ReserveAddress); err != nil {
		return fmt.Errorf("invalid reserve address: %w", err)
	}
	return nil
}

func NewPoolState(tick int32, price sdk.Dec) PoolState {
	return PoolState{
		CurrentTick:      tick,
		CurrentPrice:     price,
		CurrentLiquidity: utils.ZeroDec,
		FeeGrowthGlobal0: utils.ZeroDec,
		FeeGrowthGlobal1: utils.ZeroDec,
	}
}

func (poolState PoolState) Validate() error {
	if !poolState.CurrentPrice.IsPositive() {
		return fmt.Errorf("current price must be positive: %s", poolState.CurrentPrice)
	}
	if poolState.CurrentLiquidity.IsNegative() {
		return fmt.Errorf("current liquidity must not be negative: %s", poolState.CurrentLiquidity)
	}
	if poolState.FeeGrowthGlobal0.IsNegative() {
		return fmt.Errorf("fee growth global 0 must not be negative: %s", poolState.FeeGrowthGlobal0)
	}
	if poolState.FeeGrowthGlobal1.IsNegative() {
		return fmt.Errorf("fee growth global 1 must not be negative: %s", poolState.FeeGrowthGlobal1)
	}
	return nil
}

func DerivePoolReserveAddress(poolId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("PoolReserveAddress/%d", poolId)))
}
