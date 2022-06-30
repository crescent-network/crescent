package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v2/app/params"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/claim/keeper"
	"github.com/crescent-network/crescent/v2/x/claim/types"
	minttypes "github.com/crescent-network/crescent/v2/x/mint/types"
)

const (
	OpWeightMsgClaim = "op_weight_msg_claim"
)

var (
	Gas  = uint64(20000000)
	Fees = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
)

var (
	airdropDenom = "airdrop"
)

func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak types.AccountKeeper, bk types.BankKeeper,
	lk types.LiquidityKeeper, lsk types.LiquidStakingKeeper,
	gk types.GovKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgClaim int
	appParams.GetOrGenerate(cdc, OpWeightMsgClaim, &weightMsgClaim, nil, func(_ *rand.Rand) {
		weightMsgClaim = appparams.DefaultWeightMsgClaim
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgClaim,
			SimulateMsgClaim(ak, bk, lk, lsk, gk, k),
		),
	}
}

func SimulateMsgClaim(
	ak types.AccountKeeper, bk types.BankKeeper,
	lk types.LiquidityKeeper, lsk types.LiquidStakingKeeper,
	gk types.GovKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		accs = utils.ShuffleSimAccounts(r, accs)

		airdrop := setAirdrop(r, ctx, bk, k, accs)

		// Look for an account that has executed any condition
		var simAccount simtypes.Account
		var cond types.ConditionType
		skip := true
	loop:
		for _, simAccount = range accs {
			for _, cond = range []types.ConditionType{
				types.ConditionTypeDeposit,
				types.ConditionTypeSwap,
				types.ConditionTypeLiquidStake,
				types.ConditionTypeVote,
			} {
				if err := k.ValidateCondition(ctx, simAccount.Address, cond); err == nil {
					skip = false
					break loop
				}
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgClaim, "no recipient that has executed any condition"), nil, nil
		}

		recipient := simAccount.Address
		spendable := bk.SpendableCoins(ctx, recipient)

		// To reduce complexity, skip if the recipient has already claim record
		_, found := k.GetClaimRecordByRecipient(ctx, airdrop.Id, recipient)
		if found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgClaim, "recipient already has claim record"), nil, nil
		}

		initialClaimableCoins := sdk.NewCoins(
			sdk.NewInt64Coin(airdropDenom, int64(simtypes.RandIntBetween(r, 100_000_000, 1_000_000_000))))

		// Set new claim record for the recipient
		record := types.ClaimRecord{
			AirdropId:             airdrop.Id,
			Recipient:             recipient.String(),
			InitialClaimableCoins: initialClaimableCoins,
			ClaimableCoins:        initialClaimableCoins,
			ClaimedConditions:     []types.ConditionType{},
		}
		k.SetClaimRecord(ctx, record)

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

func setAirdrop(r *rand.Rand, ctx sdk.Context, bk types.BankKeeper, k keeper.Keeper, accs []simtypes.Account) types.Airdrop {
	sourceAddr := accs[r.Intn(len(accs))].Address
	coins := sdk.NewCoins(sdk.NewInt64Coin(airdropDenom, 10_000_000_000_000))

	if err := bk.MintCoins(ctx, minttypes.ModuleName, coins); err != nil {
		panic(err)
	}

	if err := bk.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sourceAddr, coins); err != nil {
		panic(err)
	}

	airdrop := types.Airdrop{
		Id:            1,
		SourceAddress: sourceAddr.String(),
		Conditions: []types.ConditionType{
			types.ConditionTypeDeposit,
			types.ConditionTypeSwap,
			types.ConditionTypeLiquidStake,
			types.ConditionTypeVote,
		},
		StartTime: ctx.BlockTime(),
		EndTime:   ctx.BlockTime().AddDate(0, simtypes.RandIntBetween(r, 1, 24), 0),
	}

	k.SetAirdrop(ctx, airdrop)

	return airdrop
}
