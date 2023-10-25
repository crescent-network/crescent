package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

const (
	OpWeightSubmitPublicPositionCreateProposal = "op_weight_submit_public_position_create_proposal"

	DefaultWeightPublicPositionCreateProposal = 30
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(ak types.AMMKeeper, k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitPublicPositionCreateProposal,
			DefaultWeightPublicPositionCreateProposal,
			SimulatePublicPositionCreateProposal(ak, k),
		),
	}
}

// SimulatePublicPositionCreateProposal generates random public-position-create proposal content.
// NOTE: SimulatePublicPositionCreateProposal executes the proposal manually.
func SimulatePublicPositionCreateProposal(ak types.AMMKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		var pools []ammtypes.Pool
		ak.IterateAllPools(ctx, func(pool ammtypes.Pool) (stop bool) {
			pools = append(pools, pool)
			return false
		})
		if len(pools) == 0 {
			return nil
		}
		pool := pools[r.Intn(len(pools))]
		poolState := ak.MustGetPoolState(ctx, pool.Id)
		currentPrice := poolState.CurrentSqrtPrice.Power(2)
		lowerPrice := exchangetypes.PriceAtTick(
			ammtypes.AdjustPriceToTickSpacing(
				currentPrice.Mul(utils.ParseBigDec("0.8")).Dec(), pool.TickSpacing, false))
		upperPrice := exchangetypes.PriceAtTick(
			ammtypes.AdjustPriceToTickSpacing(
				currentPrice.Mul(utils.ParseBigDec("1.25")).Dec(), pool.TickSpacing, true))
		if found := k.LookupPublicPositionByParams(
			ctx, pool.Id, exchangetypes.TickAtPrice(lowerPrice), exchangetypes.TickAtPrice(upperPrice)); found {
			return nil
		}

		p := types.NewPublicPositionCreateProposal(
			simtypes.RandStringOfLength(r, 10), simtypes.RandStringOfLength(r, 100),
			pool.Id, lowerPrice, upperPrice, utils.ParseDec("0.003"))
		// Manually execute the proposal.
		// TODO: return proposal content rather than executing it here
		if err := keeper.HandlePublicPositionCreateProposal(ctx, k, p); err != nil {
			panic(err)
		}
		return nil
	}
}
