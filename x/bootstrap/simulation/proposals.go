package simulation

//import (
//	"math/rand"
//
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
//	"github.com/cosmos/cosmos-sdk/x/simulation"
//
//	"github.com/crescent-network/crescent/v4/app/params"
//	"github.com/crescent-network/crescent/v4/x/bootstrap/keeper"
//	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
//	minttypes "github.com/crescent-network/crescent/v4/x/mint/types"
//)
//
//// Simulation operation weights constants.
//const (
//	OpWeightSimulateBootstrapProposal            = "op_weight_market_maker_proposal"
//	OpWeightSimulateChangeIncentivePairsProposal = "op_weight_change_incentive_pairs_proposal"
//	OpWeightSimulateChangeDepositAmountProposal  = "op_weight_change_deposit_amount_proposal"
//)
//
//// ProposalContents defines the module weighted proposals' contents
//func ProposalContents(bk types.BankKeeper, k keeper.Keeper) []simtypes.WeightedProposalContent {
//	return []simtypes.WeightedProposalContent{
//		simulation.NewWeightedProposalContent(
//			OpWeightSimulateBootstrapProposal,
//			params.DefaultWeightBootstrapProposal,
//			SimulateBootstrapProposal(bk, k),
//		),
//		simulation.NewWeightedProposalContent(
//			OpWeightSimulateChangeIncentivePairsProposal,
//			params.DefaultWeightChangeIncentivePairs,
//			SimulateChangeIncentivePairs(k),
//		),
//		simulation.NewWeightedProposalContent(
//			OpWeightSimulateChangeDepositAmountProposal,
//			params.DefaultWeightChangeDepositAmount,
//			SimulateChangeDepositAmount(k),
//		),
//	}
//}
//
//// SimulateBootstrapProposal generates random market maker proposal content.
//func SimulateBootstrapProposal(bk types.BankKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
//	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
//
//		params := k.GetParams(ctx)
//		_, err := fundBalances(ctx, r, bk, params.IncentiveBudgetAcc(), []string{"stake"})
//		if err != nil {
//			panic(err)
//		}
//		spendableIncentives := bk.SpendableCoins(ctx, params.IncentiveBudgetAcc())
//
//		inclusions := []types.BootstrapHandle{}
//		distributions := []types.IncentiveDistribution{}
//		exclusions := []types.BootstrapHandle{}
//		rejections := []types.BootstrapHandle{}
//
//		mms := k.GetAllBootstraps(ctx)
//		for _, mm := range mms {
//			if !mm.Eligible {
//				// inclusion
//				if simtypes.RandIntBetween(r, 0, 2) == 1 {
//					inclusions = append(inclusions, types.BootstrapHandle{
//						Address: mm.Address,
//						PairId:  mm.PairId,
//					})
//				} else {
//					// rejection
//					rejections = append(rejections, types.BootstrapHandle{
//						Address: mm.Address,
//						PairId:  mm.PairId,
//					})
//				}
//			}
//
//			// distribution
//			if mm.Eligible && simtypes.RandIntBetween(r, 0, 3) == 1 {
//				incentive := sdk.NewCoins(sdk.NewCoin("stake", spendableIncentives.AmountOf("stake").QuoRaw(1000)))
//				distributions = append(distributions, types.IncentiveDistribution{
//					Address: mm.Address,
//					PairId:  mm.PairId,
//					Amount:  incentive,
//				})
//				spendableIncentives = spendableIncentives.Sub(incentive)
//			}
//
//			// exclusion
//			if mm.Eligible && simtypes.RandIntBetween(r, 0, 7) == 1 {
//				exclusions = append(exclusions, types.BootstrapHandle{
//					Address: mm.Address,
//					PairId:  mm.PairId,
//				})
//			}
//		}
//
//		if len(inclusions) == 0 && len(exclusions) == 0 && len(rejections) == 0 && len(distributions) == 0 {
//			return nil
//		}
//
//		proposal := types.NewBootstrapProposal(
//			simtypes.RandStringOfLength(r, 10),
//			simtypes.RandStringOfLength(r, 100),
//			inclusions,
//			exclusions,
//			rejections,
//			distributions,
//		)
//		// force execute proposal to avoid waiting voting period
//		err = keeper.HandleBootstrapProposal(ctx, k, proposal)
//		if err != nil {
//			panic(err)
//		}
//		return nil
//	}
//}
//
//// SimulateChangeIncentivePairs generates random incentive pairs param change proposal content.
//func SimulateChangeIncentivePairs(k keeper.Keeper) simtypes.ContentSimulatorFn {
//	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
//		params := k.GetParams(ctx)
//
//		params.IncentivePairs = GenIncentivePairs(r)
//		k.SetParams(ctx, params)
//		return nil
//	}
//}
//
//// SimulateChangeDepositAmount generates random deposit amount param change proposal content.
//func SimulateChangeDepositAmount(k keeper.Keeper) simtypes.ContentSimulatorFn {
//	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
//		params := k.GetParams(ctx)
//
//		params.DepositAmount = GenDepositAmount(r)
//		k.SetParams(ctx, params)
//		return nil
//	}
//}
//
//// fundBalances mints random amount of coins with the provided coin denoms and
//// send them to the simulated account.
//func fundBalances(ctx sdk.Context, r *rand.Rand, bk types.BankKeeper, acc sdk.AccAddress, denoms []string) (mintCoins sdk.Coins, err error) {
//	for _, denom := range denoms {
//		mintCoins = mintCoins.Add(sdk.NewInt64Coin(denom, int64(simtypes.RandIntBetween(r, 1e14, 1e15))))
//	}
//
//	if err := bk.MintCoins(ctx, minttypes.ModuleName, mintCoins); err != nil {
//		return nil, err
//	}
//
//	if err := bk.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, acc, mintCoins); err != nil {
//		return nil, err
//	}
//	return mintCoins, nil
//}
