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
	"github.com/crescent-network/crescent/v2/x/liquidstaking/keeper"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgLiquidStake   = "op_weight_msg_liquid_stake"
	OpWeightMsgLiquidUnstake = "op_weight_msg_liquid_unstake"
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
		avs := k.GetActiveLiquidValidators(ctx, params.WhitelistedValsMap())
		if len(avs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidStake, "active liquid validators not exists"), nil, nil
		}

		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		delegator := account.GetAddress()

		nas := k.GetNetAmountState(ctx)
		var btokenUnitAmount sdk.Dec
		if nas.BtokenTotalSupply.IsZero() {
			btokenUnitAmount = sdk.OneDec()
		} else {
			btokenUnitAmount = types.BTokenToNativeToken(sdk.OneInt(), nas.BtokenTotalSupply, nas.NetAmount)
		}
		stakingAmt := int64(simtypes.RandIntBetween(r, int(btokenUnitAmount.TruncateInt64()), 100000000000000))
		if stakingAmt < params.MinLiquidStakingAmount.Int64() {
			stakingAmt = params.MinLiquidStakingAmount.Int64()
		}
		stakingCoin := sdk.NewInt64Coin(sdk.DefaultBondDenom, stakingAmt)
		if !spendable.AmountOf(sdk.DefaultBondDenom).GTE(stakingCoin.Amount) {
			if err := bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(stakingCoin)); err != nil {
				panic(err)
			}
			if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegator, sdk.NewCoins(stakingCoin)); err != nil {
				panic(err)
			}
			spendable = bk.SpendableCoins(ctx, account.GetAddress())
		}

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
		return utils.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
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
		return utils.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}
