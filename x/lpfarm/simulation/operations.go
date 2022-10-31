package simulation

import (
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v3/app/params"
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/lpfarm/keeper"
	"github.com/crescent-network/crescent/v3/x/lpfarm/types"
	minttypes "github.com/crescent-network/crescent/v3/x/mint/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgCreatePrivatePlan = "op_weight_msg_create_private_plan"
	OpWeightMsgFarm              = "op_weight_msg_farm"
	OpWeightMsgUnfarm            = "op_weight_msg_unfarm"
	OpWeightMsgHarvest           = "op_weight_msg_harvest"

	DefaultWeightCreatePrivatePlan = 10
	DefaultWeightFarm              = 40
	DefaultWeightUnfarm            = 50
	DefaultWeightHarvest           = 20
)

var (
	gas  = uint64(20000000)
	fees = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak types.AccountKeeper, bk types.BankKeeper, lk types.LiquidityKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreatePrivatePlan int
		weightMsgFarm              int
		weightMsgUnfarm            int
		weightMsgHarvest           int
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePrivatePlan, &weightMsgCreatePrivatePlan, nil, func(_ *rand.Rand) {
		weightMsgCreatePrivatePlan = DefaultWeightCreatePrivatePlan
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgFarm, &weightMsgFarm, nil, func(_ *rand.Rand) {
		weightMsgFarm = DefaultWeightFarm
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgUnfarm, &weightMsgUnfarm, nil, func(_ *rand.Rand) {
		weightMsgUnfarm = DefaultWeightUnfarm
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgHarvest, &weightMsgHarvest, nil, func(_ *rand.Rand) {
		weightMsgHarvest = DefaultWeightHarvest
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreatePrivatePlan,
			SimulateMsgCreatePrivatePlan(ak, bk, lk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgFarm,
			SimulateMsgFarm(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgUnfarm,
			SimulateMsgUnfarm(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgHarvest,
			SimulateMsgHarvest(ak, bk, k),
		),
	}
}

func SimulateMsgCreatePrivatePlan(
	ak types.AccountKeeper, bk types.BankKeeper, lk types.LiquidityKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if !k.CanCreatePrivatePlan(ctx) {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreatePrivatePlan,
				"cannot create more private plans"), nil, nil
		}

		rewardAllocs, ok := genRewardAllocs(r, ctx, lk)
		if !ok {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreatePrivatePlan, "no pairs"), nil, nil
		}

		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := types.NewMsgCreatePrivatePlan(
			simAccount.Address, "Farming Plan", rewardAllocs,
			ctx.BlockTime().AddDate(0, 0, 1),
			ctx.BlockTime().AddDate(0, 0, 2+r.Intn(5)))

		fundAddr(ctx, bk, simAccount.Address, k.GetPrivatePlanCreationFee(ctx))
		acc := ak.GetAccount(ctx, simAccount.Address)
		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			gas,
			chainID,
			[]uint64{acc.GetAccountNumber()},
			[]uint64{acc.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(
				types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}
		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		// Fund the newly created private farming plan's farming pool
		planId, _ := k.GetLastPlanId(ctx)
		plan, _ := k.GetPlan(ctx, planId)
		fundAddr(ctx, bk, plan.GetFarmingPoolAddress(), utils.ParseCoins("1000000_000000stake"))

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

func SimulateMsgFarm(ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		accs = utils.ShuffleSimAccounts(r, accs)
		var simAccount simtypes.Account
		var spendable sdk.Coins
		var coinToFarm sdk.Coin
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)
			poolCoins := sdk.Coins{}
			for _, coin := range spendable {
				if strings.HasPrefix(coin.Denom, "pool") {
					poolCoins = poolCoins.Add(coin)
				}
			}
			poolCoins = simtypes.RandSubsetCoins(r, poolCoins)
			if len(poolCoins) > 0 {
				coinToFarm = poolCoins[r.Intn(len(poolCoins))]
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgFarm, "no account to farm"), nil, nil
		}

		msg := types.NewMsgFarm(simAccount.Address, coinToFarm)

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

		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}

func SimulateMsgUnfarm(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		accs = utils.ShuffleSimAccounts(r, accs)
		var simAccount simtypes.Account
		var spendable sdk.Coins
		var coinToUnfarm sdk.Coin
		skip := true
		for _, simAccount = range accs {
			var positions []types.Position
			k.IteratePositionsByFarmer(ctx, simAccount.Address, func(position types.Position) (stop bool) {
				positions = append(positions, position)
				return false
			})
			if len(positions) > 0 {
				position := positions[r.Intn(len(positions))]
				coinToUnfarm = sdk.NewCoin(
					position.Denom,
					utils.RandomInt(r, sdk.OneInt(), position.FarmingAmount))
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgUnfarm, "no account to unfarm"), nil, nil
		}

		msg := types.NewMsgUnfarm(simAccount.Address, coinToUnfarm)

		spendable = bk.SpendableCoins(ctx, simAccount.Address)
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

		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}

func SimulateMsgHarvest(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		accs = utils.ShuffleSimAccounts(r, accs)
		var simAccount simtypes.Account
		var spendable sdk.Coins
		var denomToHarvest string
		skip := true
		for _, simAccount = range accs {
			var positions []types.Position
			k.IteratePositionsByFarmer(ctx, simAccount.Address, func(position types.Position) (stop bool) {
				positions = append(positions, position)
				return false
			})
			if len(positions) > 0 {
				position := positions[r.Intn(len(positions))]
				denomToHarvest = position.Denom
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgHarvest, "no account to harvest"), nil, nil
		}
		msg := types.NewMsgHarvest(simAccount.Address, denomToHarvest)

		spendable = bk.SpendableCoins(ctx, simAccount.Address)
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

		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}

func fundAddr(ctx sdk.Context, bk types.BankKeeper, addr sdk.AccAddress, amt sdk.Coins) {
	if err := bk.MintCoins(ctx, minttypes.ModuleName, amt); err != nil {
		panic(err)
	}
	if err := bk.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amt); err != nil {
		panic(err)
	}
}

func genRewardAllocs(r *rand.Rand, ctx sdk.Context, lk types.LiquidityKeeper) (rewardAllocs []types.RewardAllocation, ok bool) {
	pairs := lk.GetAllPairs(ctx)
	if len(pairs) == 0 {
		return nil, false
	}
	r.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})
	n := r.Intn(len(pairs)) + 1 // Number of reward allocs
	for i := 0; i < n; i++ {
		pair := pairs[i]
		rewardsPerDay := simtypes.RandSubsetCoins(r, utils.ParseCoins("1000_000000stake"))
		rewardAllocs = append(rewardAllocs, types.NewPairRewardAllocation(pair.Id, rewardsPerDay))
	}
	return rewardAllocs, true
}
