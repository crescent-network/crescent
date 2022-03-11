package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/cosmosquad-labs/squad/app/params"
	utils "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/keeper"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgLiquidStake   = "op_weight_msg_liquid_stake"
	OpWeightMsgLiquidUnstake = "op_weight_msg_liquid_unstake"
	LiquidStakingGas         = 20000000
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper,
	bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {

	var weightMsgLiquidStake int
	appParams.GetOrGenerate(cdc, OpWeightMsgLiquidStake, &weightMsgLiquidStake, nil,
		func(_ *rand.Rand) {
			weightMsgLiquidStake = appparams.DefaultWeightMsgLiquidStake
		},
	)

	var weightMsgLiquidUnstake int
	appParams.GetOrGenerate(cdc, OpWeightMsgLiquidUnstake, &weightMsgLiquidUnstake, nil,
		func(_ *rand.Rand) {
			weightMsgLiquidUnstake = appparams.DefaultWeightMsgLiquidUnstake
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgLiquidStake,
			SimulateMsgLiquidStake(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgLiquidUnstake,
			SimulateMsgLiquidUnstake(ak, bk, k),
		),
	}
}

// SimulateMsgLiquidStake generates a MsgStake with random values
// nolint: interfacer
func SimulateMsgLiquidStake(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		params := k.GetParams(ctx)
		avs := k.GetActiveLiquidValidators(ctx, params.WhitelistedValMap())
		if len(avs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidStake, "active liquid validators not exists"), nil, nil
		}

		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		delegator := account.GetAddress()

		stakingCoin := sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simtypes.RandIntBetween(r, int(params.MinLiquidStakingAmount.Int64()), 1_000_000_000)))

		if !spendable.AmountOf(sdk.DefaultBondDenom).GTE(stakingCoin.Amount) {
			if err := bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(stakingCoin)); err != nil {
				panic(err)
			}
			if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegator, sdk.NewCoins(stakingCoin)); err != nil {
				panic(err)
			}
		}
		fmt.Println("## ADD liquid NetAmountState", stakingCoin)
		utils.PP(k.NetAmountState(ctx))

		msg := types.NewMsgLiquidStake(delegator, stakingCoin)
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
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}
		return GenAndDeliverTxWithRandFees(txCtx, LiquidStakingGas)
	}
}

// SimulateMsgLiquidUnstake generates a SimulateMsgLiquidUnstake with random values
// nolint: interfacer
func SimulateMsgLiquidUnstake(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		simAccount, _ := simtypes.RandomAcc(r, accs)
		var delegator sdk.AccAddress
		var unstakingCoin sdk.Coin
		var spendable sdk.Coins

		for i := 1; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)

			account := ak.GetAccount(ctx, simAccount.Address)
			spendable = bk.SpendableCoins(ctx, account.GetAddress())

			delegator = account.GetAddress()
			unstakingCoin = sdk.NewInt64Coin(types.DefaultLiquidBondDenom, int64(simtypes.RandIntBetween(r, 1_000_000, 100_000_000)))

			// spendable must be greater than unstaking coins
			if spendable.AmountOf(types.DefaultLiquidBondDenom).GTE(unstakingCoin.Amount) {
				fmt.Println("## UNBONDING NetAmountState", unstakingCoin)
				utils.PP(k.NetAmountState(ctx))
				break
			}
		}

		if !spendable.AmountOf(types.DefaultLiquidBondDenom).GTE(unstakingCoin.Amount) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidUnstake, "insufficient funds"), nil, nil
		}

		msg := types.NewMsgLiquidUnstake(delegator, unstakingCoin)
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
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}
		return GenAndDeliverTxWithRandFees(txCtx, LiquidStakingGas)
	}
}

// GenAndDeliverTx generates a transactions and delivers it.
func GenAndDeliverTx(txCtx simulation.OperationInput, fees sdk.Coins, gas uint64) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	tx, err := helpers.GenTx(
		txCtx.TxGen,
		[]sdk.Msg{txCtx.Msg},
		fees,
		gas,
		txCtx.Context.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		txCtx.SimAccount.PrivKey,
	)

	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate mock tx"), nil, err
	}

	_, _, err = txCtx.App.Deliver(txCtx.TxGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to deliver tx"), nil, err
	}

	return simtypes.NewOperationMsg(txCtx.Msg, true, "", txCtx.Cdc), nil, nil

}

// GenAndDeliverTxWithRandFees generates a transaction with a random fee and delivers it.
func GenAndDeliverTxWithRandFees(txCtx simulation.OperationInput, gas uint64) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	spendable := txCtx.Bankkeeper.SpendableCoins(txCtx.Context, account.GetAddress())

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(txCtx.CoinsSpentInMsg)
	if hasNeg {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "message doesn't leave room for fees"), nil, err
	}

	fees, err = simtypes.RandomFees(txCtx.R, txCtx.Context, coins)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(txCtx, fees, gas)
}
