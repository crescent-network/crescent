package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func SqrtPriceAtTick(tick int32) sdk.Dec {
	return utils.DecApproxSqrt(exchangetypes.PriceAtTick(tick))
}

// AdjustPriceToTickSpacing returns rounded tick price considering the tick
// spacing.
func AdjustPriceToTickSpacing(price sdk.Dec, tickSpacing uint32, roundUp bool) sdk.Dec {
	// We assume that the price is already a valid tick price, so here's
	// no check for that.
	tick := exchangetypes.TickAtPrice(price)
	ts := int32(tickSpacing)
	if roundUp {
		q, _ := utils.DivMod(tick+ts-1, ts)
		return exchangetypes.PriceAtTick(q * ts)
	}
	q, _ := utils.DivMod(tick, ts)
	return exchangetypes.PriceAtTick(q * ts)
}

func NewTickInfo(grossLiquidity, netLiquidity sdk.Int) TickInfo {
	return TickInfo{
		GrossLiquidity:              grossLiquidity,
		NetLiquidity:                netLiquidity,
		FeeGrowthOutside:            sdk.DecCoins{},
		FarmingRewardsGrowthOutside: sdk.DecCoins{},
	}
}

func (tickInfo TickInfo) Validate() error {
	if tickInfo.GrossLiquidity.IsNegative() {
		return fmt.Errorf("gross liquidity must not be negative: %s", tickInfo.GrossLiquidity)
	}
	if tickInfo.NetLiquidity.IsZero() {
		return fmt.Errorf("net liquidity must not be 0")
	}
	if err := tickInfo.FeeGrowthOutside.Validate(); err != nil {
		return fmt.Errorf("invalid fee growth outside: %w", err)
	}
	if len(tickInfo.FeeGrowthOutside) > 2 {
		return fmt.Errorf(
			"number of coins in fee growth outside must not be higher than 2: %d", len(tickInfo.FeeGrowthOutside))
	}
	if err := tickInfo.FarmingRewardsGrowthOutside.Validate(); err != nil {
		return fmt.Errorf("invalid farming rewards growth outside: %w", err)
	}
	return nil
}
