package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v4/types"
	"github.com/crescent-network/crescent/v4/x/liquidity/amm"
)

func DeriveFarmingPoolAddress(planId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("FarmingPool/%d", planId)))
}

func DeriveFarmingReserveAddress(denom string) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("FarmingReserve/%s", denom)))
}

func RewardsForBlock(rewardsPerDay sdk.Coins, blockDuration time.Duration) sdk.DecCoins {
	return sdk.NewDecCoinsFromCoins(rewardsPerDay...).
		MulDecTruncate(sdk.NewDec(blockDuration.Milliseconds())).
		QuoDecTruncate(sdk.NewDec(day.Milliseconds()))
}

// PoolRewardWeight returns given pool's reward weight.
func PoolRewardWeight(pool amm.Pool) (weight sdk.Dec) {
	rx, ry := pool.Balances()
	sqrt := utils.DecApproxSqrt
	switch pool := pool.(type) {
	case *amm.BasicPool:
		weight = sqrt(sdk.NewDecFromInt(rx.Mul(ry)))
	case *amm.RangedPool:
		transX, transY := pool.Translation()
		weight = sqrt(transX.Add(sdk.NewDecFromInt(rx))).Mul(sqrt(transY.Add(sdk.NewDecFromInt(ry))))
	default:
		panic("invalid pool type")
	}
	return
}
