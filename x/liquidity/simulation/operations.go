package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	squadappparams "github.com/cosmosquad-labs/squad/app/params"
	liquiditykeeper "github.com/cosmosquad-labs/squad/x/liquidity/keeper"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgCreatePair      = "op_weight_msg_create_pair"
	OpWeightMsgCreatePool      = "op_weight_msg_create_pool"
	OpWeightMsgDeposit         = "op_weight_msg_deposit"
	OpWeightMsgWithdraw        = "op_weight_msg_withdraw"
	OpWeightMsgLimitOrder      = "op_weight_msg_limit_order"
	OpWeightMsgMarketOrder     = "op_weight_msg_market_order"
	OpWeightMsgCancelOrder     = "op_weight_msg_cancel_order"
	OpWeightMsgCancelAllOrders = "op_weight_msg_cancel_all_orders"
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper,
	bk types.BankKeeper, k liquiditykeeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgCreatePair int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePair, &weightMsgCreatePair, nil, func(_ *rand.Rand) {
		weightMsgCreatePair = squadappparams.DefaultWeightMsgCreatePair
	})

	var weightMsgCreatePool int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePool, &weightMsgCreatePool, nil, func(_ *rand.Rand) {
		weightMsgCreatePool = squadappparams.DefaultWeightMsgCreatePool
	})

	var weightMsgDeposit int
	appParams.GetOrGenerate(cdc, OpWeightMsgDeposit, &weightMsgDeposit, nil, func(_ *rand.Rand) {
		weightMsgCreatePool = squadappparams.DefaultWeightMsgCreatePool
	})

	var weightMsgWithdraw int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdraw, &weightMsgWithdraw, nil, func(_ *rand.Rand) {
		weightMsgCreatePool = squadappparams.DefaultWeightMsgWithdraw
	})

	var weightMsgLimitOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgLimitOrder, &weightMsgLimitOrder, nil, func(_ *rand.Rand) {
		weightMsgCreatePool = squadappparams.DefaultWeightMsgLimitOrder
	})

	var weightMsgMarketOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgMarketOrder, &weightMsgMarketOrder, nil, func(_ *rand.Rand) {
		weightMsgCreatePool = squadappparams.DefaultWeightMsgMarketOrder
	})

	var weightMsgCancelOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelOrder, &weightMsgCancelOrder, nil, func(_ *rand.Rand) {
		weightMsgCreatePool = squadappparams.DefaultWeightMsgCancelOrder
	})

	var weightMsgCancelAllOrders int
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelAllOrders, &weightMsgCancelAllOrders, nil, func(_ *rand.Rand) {
		weightMsgCreatePool = squadappparams.DefaultWeightMsgCancelAllOrders
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreatePair,
			SimulateMsgCreatePair(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCreatePool,
			SimulateMsgCreatePool(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgDeposit,
			SimulateMsgDeposit(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdraw,
			SimulateMsgWithdraw(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgLimitOrder,
			SimulateMsgLimitOrder(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgMarketOrder,
			SimulateMsgMarketOrder(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelOrder,
			SimulateMsgCancelOrder(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelAllOrders,
			SimulateMsgCancelAllOrders(ak, bk, k),
		),
	}
}

func SimulateMsgCreatePair(ak types.AccountKeeper, bk types.BankKeeper, k liquiditykeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		params := k.GetParams(ctx)

		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		if !spendable.IsAllGTE(params.PairCreationFee) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePair, "insufficient balance for pair creation fee"), nil, nil
		}

		// Find a non-existing denom pair from total supplies.
		var denoms []string
		bk.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
			denoms = append(denoms, coin.Denom)
			return false
		})
		r.Shuffle(len(denoms), func(i, j int) {
			denoms[i], denoms[j] = denoms[j], denoms[i]
		})
		var denomA, denomB string
		skip := true
	loop:
		for _, denomA = range denoms {
			for _, denomB = range denoms {
				if denomA != denomB {
					if _, found := k.GetPairByDenoms(ctx, denomA, denomB); !found {
						skip = false
						break loop
					}
				}
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePair, "all pairs have already been created"), nil, nil
		}

		msg := types.NewMsgCreatePair(simAccount.Address, denomA, denomB)

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

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgCreatePool(ak types.AccountKeeper, bk types.BankKeeper, k liquiditykeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		params := k.GetParams(ctx)
		minDepositAmt := params.MinInitialDepositAmount

		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		if !spendable.IsAllGTE(params.PoolCreationFee) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePool, "insufficient balance for pool creation fee"), nil, nil
		}

		// Select a random pair id.
		var pairs []types.Pair
		_ = k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool, err error) {
			pairs = append(pairs, pair)
			return false, nil
		})
		r.Shuffle(len(pairs), func(i, j int) {
			pairs[i], pairs[j] = pairs[j], pairs[i]
		})
		var pair types.Pair
		skip := true
		for _, pair = range pairs {
			found := false
			_ = k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool, err error) {
				if !pool.Disabled {
					found = true
					return true, nil
				}
				return false, nil
			})
			minDepositCoins := sdk.NewCoins(
				sdk.NewCoin(pair.BaseCoinDenom, minDepositAmt),
				sdk.NewCoin(pair.QuoteCoinDenom, minDepositAmt),
			)
			if !found && spendable.IsAllGTE(minDepositCoins.Add(params.PoolCreationFee...)) {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePool, "all possible pools have been created"), nil, nil
		}

		depositCoins := sdk.NewCoins(
			sdk.NewCoin(
				pair.BaseCoinDenom,
				minDepositAmt.Add(simtypes.RandomAmount(r, spendable.AmountOf(pair.BaseCoinDenom).Sub(minDepositAmt))),
			),
			sdk.NewCoin(
				pair.QuoteCoinDenom,
				minDepositAmt.Add(simtypes.RandomAmount(r, spendable.AmountOf(pair.QuoteCoinDenom).Sub(minDepositAmt))),
			),
		)

		msg := types.NewMsgCreatePool(simAccount.Address, pair.Id, depositCoins)

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

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgDeposit(ak types.AccountKeeper, bk types.BankKeeper, k liquiditykeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeposit, ""), nil, nil
		//txCtx := simulation.OperationInput{
		//	R:               r,
		//	App:             app,
		//	TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
		//	Cdc:             nil,
		//	Msg:             msg,
		//	MsgType:         msg.Type(),
		//	Context:         ctx,
		//	SimAccount:      simAccount,
		//	AccountKeeper:   ak,
		//	Bankkeeper:      bk,
		//	ModuleName:      types.ModuleName,
		//	CoinsSpentInMsg: spendable,
		//}
		//
		//return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgWithdraw(ak types.AccountKeeper, bk types.BankKeeper, k liquiditykeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWithdraw, ""), nil, nil
		//txCtx := simulation.OperationInput{
		//	R:               r,
		//	App:             app,
		//	TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
		//	Cdc:             nil,
		//	Msg:             msg,
		//	MsgType:         msg.Type(),
		//	Context:         ctx,
		//	SimAccount:      simAccount,
		//	AccountKeeper:   ak,
		//	Bankkeeper:      bk,
		//	ModuleName:      types.ModuleName,
		//	CoinsSpentInMsg: spendable,
		//}
		//
		//return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgLimitOrder(ak types.AccountKeeper, bk types.BankKeeper, k liquiditykeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLimitOrder, ""), nil, nil
		//txCtx := simulation.OperationInput{
		//	R:               r,
		//	App:             app,
		//	TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
		//	Cdc:             nil,
		//	Msg:             msg,
		//	MsgType:         msg.Type(),
		//	Context:         ctx,
		//	SimAccount:      simAccount,
		//	AccountKeeper:   ak,
		//	Bankkeeper:      bk,
		//	ModuleName:      types.ModuleName,
		//	CoinsSpentInMsg: spendable,
		//}
		//
		//return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgMarketOrder(ak types.AccountKeeper, bk types.BankKeeper, k liquiditykeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMarketOrder, ""), nil, nil
		//txCtx := simulation.OperationInput{
		//	R:               r,
		//	App:             app,
		//	TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
		//	Cdc:             nil,
		//	Msg:             msg,
		//	MsgType:         msg.Type(),
		//	Context:         ctx,
		//	SimAccount:      simAccount,
		//	AccountKeeper:   ak,
		//	Bankkeeper:      bk,
		//	ModuleName:      types.ModuleName,
		//	CoinsSpentInMsg: spendable,
		//}
		//
		//return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgCancelOrder(ak types.AccountKeeper, bk types.BankKeeper, k liquiditykeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCancelOrder, ""), nil, nil
		//txCtx := simulation.OperationInput{
		//	R:               r,
		//	App:             app,
		//	TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
		//	Cdc:             nil,
		//	Msg:             msg,
		//	MsgType:         msg.Type(),
		//	Context:         ctx,
		//	SimAccount:      simAccount,
		//	AccountKeeper:   ak,
		//	Bankkeeper:      bk,
		//	ModuleName:      types.ModuleName,
		//	CoinsSpentInMsg: spendable,
		//}
		//
		//return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgCancelAllOrders(ak types.AccountKeeper, bk types.BankKeeper, k liquiditykeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCancelAllOrders, ""), nil, nil
		//txCtx := simulation.OperationInput{
		//	R:               r,
		//	App:             app,
		//	TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
		//	Cdc:             nil,
		//	Msg:             msg,
		//	MsgType:         msg.Type(),
		//	Context:         ctx,
		//	SimAccount:      simAccount,
		//	AccountKeeper:   ak,
		//	Bankkeeper:      bk,
		//	ModuleName:      types.ModuleName,
		//	CoinsSpentInMsg: spendable,
		//}
		//
		//return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
