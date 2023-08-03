package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// Simulation operation weights constants.
const (
	OpWeightSubmitPublicPositionCreateProposal = "op_weight_submit_public_position_create_proposal"

	DefaultWeightPublicPositionCreateProposal = 10
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(ammK types.AMMKeeper, k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitPublicPositionCreateProposal,
			DefaultWeightPublicPositionCreateProposal,
			SimulatePublicPositionCreateProposalContent(ammK, k),
		),
	}
}

// SimulatePublicPositionCreateProposalContent generates random public-position-create proposal content.
func SimulatePublicPositionCreateProposalContent(ammK types.AMMKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		var pools []ammtypes.Pool
		ammK.IterateAllPools(ctx, func(pool ammtypes.Pool) (stop bool) {
			pools = append(pools, pool)
			return false
		})
		if len(pools) == 0 {
			return nil
		}
		pool := pools[r.Intn(len(pools))]
		poolState := ammK.MustGetPoolState(ctx, pool.Id)
		lowerPrice := ammtypes.AdjustPriceToTickSpacing(
			poolState.CurrentPrice.Mul(utils.ParseDec("0.8")), pool.TickSpacing, false)
		upperPrice := ammtypes.AdjustPriceToTickSpacing(
			poolState.CurrentPrice.Mul(utils.ParseDec("1.25")), pool.TickSpacing, true)
		minBidAmt := utils.RandomInt(r, sdk.NewInt(10000), sdk.NewInt(1000000))

		p := types.NewPublicPositionCreateProposal(
			simtypes.RandStringOfLength(r, 10), simtypes.RandStringOfLength(r, 100),
			pool.Id, lowerPrice, upperPrice, minBidAmt, utils.ParseDec("0.003"))
		// Manually handle the proposal.
		// TODO: return proposal content rather than executing it here
		if err := keeper.HandlePublicPositionCreateProposal(ctx, k, p); err != nil {
			panic(err)
		}
		return nil
	}
}
