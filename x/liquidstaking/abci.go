package liquidstaking

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/liquidstaking/keeper"
	"github.com/tendermint/farming/x/liquidstaking/types"
)

// BeginBlocker collects liquidStakings for the current block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	// TODO: Unimplemented beginblock logic
	//if err != nil {
	//	panic(err)
	//}
}
