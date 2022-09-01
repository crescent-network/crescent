package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v2/app/params"
	utils "github.com/crescent-network/crescent/v2/types"
	marketmakerkeeper "github.com/crescent-network/crescent/v2/x/marketmaker/keeper"
	marketmakertypes "github.com/crescent-network/crescent/v2/x/marketmaker/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgApplyMarketMaker = "op_weight_msg_apply_market_maker"
	OpWeightMsgClaimIncentives  = "op_weight_msg_claim_incentives"
)

var (
	Gas  = uint64(20000000)
	Fees = sdk.Coins{
		{
			Denom:  "stake",
			Amount: sdk.NewInt(0),
		},
	}
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak marketmakertypes.AccountKeeper,
	bk marketmakertypes.BankKeeper, k marketmakerkeeper.Keeper,
) simulation.WeightedOperations {

	var weightMsgApplyMarketMaker int
	appParams.GetOrGenerate(cdc, OpWeightMsgApplyMarketMaker, &weightMsgApplyMarketMaker, nil,
		func(_ *rand.Rand) {
			weightMsgApplyMarketMaker = appparams.DefaultWeightMsgApplyMarketMaker
		},
	)

	var weightMsgClaimIncentives int
	appParams.GetOrGenerate(cdc, OpWeightMsgClaimIncentives, &weightMsgClaimIncentives, nil,
		func(_ *rand.Rand) {
			weightMsgClaimIncentives = appparams.DefaultWeightMsgClaimIncentives
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgApplyMarketMaker,
			SimulateMsgApplyMarketMaker(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgClaimIncentives,
			SimulateMsgClaimIncentives(ak, bk, k),
		),
	}
}

// SimulateMsgApplyMarketMaker generates a MsgApplyMarketMaker with random values
// nolint: interfacer
func SimulateMsgApplyMarketMaker(ak marketmakertypes.AccountKeeper, bk marketmakertypes.BankKeeper, k marketmakerkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)

		params := k.GetParams(ctx)
		var pairs []uint64
		applyDeposit := sdk.Coins{}

		if simtypes.RandIntBetween(r, 0, 2) == 1 {
			pairs = []uint64{2}
			applyDeposit = applyDeposit.Add(params.DepositAmount...)
		} else {
			pairs = []uint64{2, 3}
			applyDeposit = applyDeposit.Add(params.DepositAmount...).Add(params.DepositAmount...)
		}

		for _, pair := range pairs {
			_, found := k.GetMarketMaker(ctx, account.GetAddress(), pair)
			if found {
				return simtypes.NoOpMsg(marketmakertypes.ModuleName, marketmakertypes.TypeMsgApplyMarketMaker, "already exist market maker"), nil, nil
			}
		}

		msg := marketmakertypes.NewMsgApplyMarketMaker(account.GetAddress(), pairs)
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      marketmakertypes.ModuleName,
			CoinsSpentInMsg: applyDeposit,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgClaimIncentives generates a MsgClaimIncentives with random values
// nolint: interfacer
func SimulateMsgClaimIncentives(ak marketmakertypes.AccountKeeper, bk marketmakertypes.BankKeeper, k marketmakerkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account

		skip := true
		// find incentive from the simulated accounts
		for _, acc := range accs {
			_, found := k.GetIncentive(ctx, acc.Address)
			if found {
				simAccount = acc
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(marketmakertypes.ModuleName, marketmakertypes.TypeMsgClaimIncentives, "no account to claim rewards"), nil, nil
		}

		msg := marketmakertypes.NewMsgClaimIncentives(simAccount.Address)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      marketmakertypes.ModuleName,
			CoinsSpentInMsg: sdk.Coins{},
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
