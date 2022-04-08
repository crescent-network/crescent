package simulation

import (
	"math/rand"
	"sort"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/app/params"
	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/claim/keeper"
	"github.com/crescent-network/crescent/x/claim/types"
)

const (
	OpWeightMsgClaim = "op_weight_msg_claim"
)

var (
	Gas  = uint64(20000000)
	Fees = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
)

func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgClaim int
	appParams.GetOrGenerate(cdc, OpWeightMsgClaim, &weightMsgClaim, nil, func(_ *rand.Rand) {
		weightMsgClaim = appparams.DefaultWeightMsgClaim
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgClaim,
			SimulateMsgClaim(ak, bk, k),
		),
	}
}

func SimulateMsgClaim(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		accs = utils.ShuffleSimAccounts(r, accs)

		airdrops := k.GetAllAirdrops(ctx)
		rand.Shuffle(len(airdrops), func(i, j int) {
			airdrops[i], airdrops[j] = airdrops[j], airdrops[i]
		})
		var simAccount simtypes.Account
		var airdrop types.Airdrop
		var claimRecord types.ClaimRecord
		var cond types.ConditionType
		skip := true
	loop:
		for _, simAccount = range accs {
			for _, airdrop = range airdrops {
				var found bool
				claimRecord, found = k.GetClaimRecordByRecipient(ctx, airdrop.Id, simAccount.Address)
				if !found {
					continue
				}

				conditions := unclaimedConditions(airdrop, claimRecord)
				for _, cond = range conditions {
					if err := k.ValidateCondition(ctx, simAccount.Address, cond); err == nil {
						skip = false
						break loop
					}
				}
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgClaim, "no account to claim"), nil, nil
		}

		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		msg := types.NewMsgClaim(airdrop.Id, simAccount.Address, cond)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           appparams.MakeTestEncodingConfig().TxConfig,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

func unclaimedConditions(airdrop types.Airdrop, claimRecord types.ClaimRecord) []types.ConditionType {
	conditionSet := map[types.ConditionType]struct{}{}
	for _, cond := range airdrop.Conditions {
		conditionSet[cond] = struct{}{}
	}
	for _, cond := range claimRecord.ClaimedConditions {
		delete(conditionSet, cond)
	}
	var conditions []types.ConditionType
	for cond := range conditionSet {
		conditions = append(conditions, cond)
	}
	// Sort conditions for deterministic simulation, since map keys are not sorted.
	sort.Slice(conditions, func(i, j int) bool {
		return conditions[i] < conditions[j]
	})
	return conditions
}
