package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v5/app/params"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgCreatePool      = "op_weight_msg_create_pool"
	OpWeightMsgAddLiquidity    = "op_weight_msg_add_liquidity"
	OpWeightMsgRemoveLiquidity = "op_weight_msg_remove_liquidity"

	DefaultWeightMsgCreatePool      = 5
	DefaultWeightMsgAddLiquidity    = 70
	DefaultWeightMsgRemoveLiquidity = 50
)

var (
	gas  = uint64(20000000)
	fees = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak types.AccountKeeper, bk types.BankKeeper, ek types.ExchangeKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreatePool      int
		weightMsgAddLiquidity    int
		weightMsgRemoveLiquidity int
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePool, &weightMsgCreatePool, nil, func(_ *rand.Rand) {
		weightMsgCreatePool = DefaultWeightMsgCreatePool
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgAddLiquidity, &weightMsgAddLiquidity, nil, func(_ *rand.Rand) {
		weightMsgAddLiquidity = DefaultWeightMsgAddLiquidity
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgRemoveLiquidity, &weightMsgRemoveLiquidity, nil, func(_ *rand.Rand) {
		weightMsgRemoveLiquidity = DefaultWeightMsgRemoveLiquidity
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreatePool,
			SimulateMsgCreatePool(ak, bk, ek, k),
		),
		simulation.NewWeightedOperation(
			weightMsgAddLiquidity,
			SimulateMsgAddLiquidity(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgRemoveLiquidity,
			SimulateMsgRemoveLiquidity(ak, bk, k),
		),
	}
}

func SimulateMsgCreatePool(
	ak types.AccountKeeper, bk types.BankKeeper, ek types.ExchangeKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgCreatePoolParams(r, accs, ek, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreatePool, "unable to create pool"), nil, nil
		}

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
			CoinsSpentInMsg: bk.SpendableCoins(ctx, simAccount.Address),
		}
		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}

func SimulateMsgAddLiquidity(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgAddLiquidityParams(r, accs, bk, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgAddLiquidity, "unable to add liquidity"), nil, nil
		}
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
			CoinsSpentInMsg: bk.SpendableCoins(ctx, simAccount.Address),
		}
		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}

func SimulateMsgRemoveLiquidity(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgPlaceRemoveLiquidityParams(r, accs, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgRemoveLiquidity, "unable to remove liquidity"), nil, nil
		}
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
			CoinsSpentInMsg: bk.SpendableCoins(ctx, simAccount.Address),
		}
		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}

func findMsgCreatePoolParams(r *rand.Rand, accs []simtypes.Account,
	ek types.ExchangeKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgCreatePool, found bool) {
	var markets []exchangetypes.Market
	ek.IterateAllMarkets(ctx, func(market exchangetypes.Market) (stop bool) {
		markets = append(markets, market)
		return false
	})
	utils.Shuffle(r, markets)
	for _, market := range markets {
		if found := k.LookupPoolByMarket(ctx, market.Id); !found {
			acc, _ = simtypes.RandomAcc(r, accs)
			marketState := ek.MustGetMarketState(ctx, market.Id)
			var price sdk.Dec
			if marketState.LastPrice != nil {
				price = *marketState.LastPrice
			} else {
				price = utils.RandomDec(r, utils.ParseDec("0.05"), utils.ParseDec("500"))
			}
			msg = types.NewMsgCreatePool(acc.Address, market.Id, price)
			return acc, msg, true
		}
	}
	return acc, msg, false
}

func findMsgAddLiquidityParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgAddLiquidity, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	var pools []types.Pool
	k.IterateAllPools(ctx, func(pool types.Pool) (stop bool) {
		pools = append(pools, pool)
		return false
	})
	utils.Shuffle(r, pools)
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		for _, pool := range pools {
			if spendable.AmountOf(pool.Denom0).GT(sdk.NewInt(100_000000)) &&
				spendable.AmountOf(pool.Denom1).GT(sdk.NewInt(100_000000)) {
				ts := int32(pool.TickSpacing)
				poolState := k.MustGetPoolState(ctx, pool.Id)
				var lowerPrice, upperPrice sdk.Dec
				if r.Float64() <= 0.2 {
					lowerPrice = exchangetypes.PriceAtTick(types.MinTick)
				} else if r.Float64() <= 0.5 {
					lowerPrice = exchangetypes.PriceAtTick(
						exchangetypes.TickAtPrice(
							utils.RandomDec(r,
								poolState.CurrentPrice.Mul(utils.ParseDec("0.5")),
								poolState.CurrentPrice)) / ts * ts)
				} else {
					lowerPrice = exchangetypes.PriceAtTick(
						exchangetypes.TickAtPrice(
							utils.RandomDec(r,
								poolState.CurrentPrice,
								poolState.CurrentPrice.Mul(utils.ParseDec("1.5")))))
				}
				if r.Float64() <= 0.2 {
					upperPrice = exchangetypes.PriceAtTick(types.MaxTick)
				} else {
					upperPrice = exchangetypes.PriceAtTick(
						exchangetypes.TickAtPrice(
							utils.RandomDec(r,
								lowerPrice.Mul(utils.ParseDec("0.01")),
								poolState.CurrentPrice.Mul(utils.ParseDec("3")))) / ts * ts)
				}
				desiredAmt := sdk.NewCoins(
					sdk.NewCoin(
						pool.Denom0,
						utils.RandomInt(r, sdk.NewInt(100), sdk.NewInt(100_000000))),
					sdk.NewCoin(
						pool.Denom1,
						utils.RandomInt(r, sdk.NewInt(100), sdk.NewInt(100_000000))))
				msg = types.NewMsgAddLiquidity(
					acc.Address, pool.Id, lowerPrice, upperPrice, desiredAmt)
				return acc, msg, true
			}
		}
	}
	return acc, msg, false
}

func findMsgPlaceRemoveLiquidityParams(
	r *rand.Rand, accs []simtypes.Account,
	k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgRemoveLiquidity, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	for _, acc = range accs {
		var positions []types.Position
		k.IteratePositionsByOwner(ctx, acc.Address, func(position types.Position) (stop bool) {
			if position.Liquidity.IsPositive() {
				positions = append(positions, position)
			}
			return false
		})
		if len(positions) > 0 {
			position := positions[r.Intn(len(positions))]
			liquidity := utils.RandomInt(r, sdk.NewInt(100), position.Liquidity)
			msg = types.NewMsgRemoveLiquidity(acc.Address, position.Id, liquidity)
			return acc, msg, true
		}
	}
	return acc, msg, false
}
