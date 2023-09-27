package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

const (
	OpWeightSubmitMarketParameterChangeProposal = "op_weight_submit_market_parameter_change_proposal"

	DefaultWeightMarketParameterChangeProposal = 5
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitMarketParameterChangeProposal,
			DefaultWeightMarketParameterChangeProposal,
			SimulateMarketParameterChangeProposal(k),
		),
	}
}

func SimulateMarketParameterChangeProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		var changes []types.MarketParameterChange
		k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
			if r.Float64() <= 0.5 {
				fees := GenFees(r)
				changes = append(changes, types.NewMarketParameterChange(
					market.Id,
					fees.DefaultMakerFeeRate, fees.DefaultTakerFeeRate, fees.DefaultOrderSourceFeeRatio,
					nil, nil, nil, nil)) // XXX
			}
			return false
		})
		if len(changes) == 0 {
			return nil
		}
		return types.NewMarketParameterChangeProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			changes,
		)
	}
}
