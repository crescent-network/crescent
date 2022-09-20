package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
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
