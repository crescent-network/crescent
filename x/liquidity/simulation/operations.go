package simulation

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v3/app/params"
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/liquidity/amm"
	"github.com/crescent-network/crescent/v3/x/liquidity/keeper"
	"github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgCreatePair       = "op_weight_msg_create_pair"
	OpWeightMsgCreatePool       = "op_weight_msg_create_pool"
	OpWeightMsgCreateRangedPool = "op_weight_msg_create_ranged_pool"
	OpWeightMsgDeposit          = "op_weight_msg_deposit"
	OpWeightMsgWithdraw         = "op_weight_msg_withdraw"
	OpWeightMsgLimitOrder       = "op_weight_msg_limit_order"
	OpWeightMsgMarketOrder      = "op_weight_msg_market_order"
	OpWeightMsgMMOrder          = "op_weight_msg_mm_order"
	OpWeightMsgCancelOrder      = "op_weight_msg_cancel_order"
	OpWeightMsgCancelAllOrders  = "op_weight_msg_cancel_all_orders"
	OpWeightMsgCancelMMOrder    = "op_weight_msg_cancel_mm_order"
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
	var weightMsgCreatePair int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePair, &weightMsgCreatePair, nil, func(_ *rand.Rand) {
		weightMsgCreatePair = appparams.DefaultWeightMsgCreatePair
	})

	var weightMsgCreatePool int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePool, &weightMsgCreatePool, nil, func(_ *rand.Rand) {
		weightMsgCreatePool = appparams.DefaultWeightMsgCreatePool
	})

	var weightMsgCreateRangedPool int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateRangedPool, &weightMsgCreateRangedPool, nil, func(_ *rand.Rand) {
		weightMsgCreateRangedPool = appparams.DefaultWeightMsgCreateRangedPool
	})

	var weightMsgDeposit int
	appParams.GetOrGenerate(cdc, OpWeightMsgDeposit, &weightMsgDeposit, nil, func(_ *rand.Rand) {
		weightMsgDeposit = appparams.DefaultWeightMsgDeposit
	})

	var weightMsgWithdraw int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdraw, &weightMsgWithdraw, nil, func(_ *rand.Rand) {
		weightMsgWithdraw = appparams.DefaultWeightMsgWithdraw
	})

	var weightMsgLimitOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgLimitOrder, &weightMsgLimitOrder, nil, func(_ *rand.Rand) {
		weightMsgLimitOrder = appparams.DefaultWeightMsgLimitOrder
	})

	var weightMsgMarketOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgMarketOrder, &weightMsgMarketOrder, nil, func(_ *rand.Rand) {
		weightMsgMarketOrder = appparams.DefaultWeightMsgMarketOrder
	})

	var weightMsgMMOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgMMOrder, &weightMsgMMOrder, nil, func(_ *rand.Rand) {
		weightMsgMMOrder = appparams.DefaultWeightMsgMMOrder
	})

	var weightMsgCancelOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelOrder, &weightMsgCancelOrder, nil, func(_ *rand.Rand) {
		weightMsgCancelOrder = appparams.DefaultWeightMsgCancelOrder
	})

	var weightMsgCancelAllOrders int
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelAllOrders, &weightMsgCancelAllOrders, nil, func(_ *rand.Rand) {
		weightMsgCancelAllOrders = appparams.DefaultWeightMsgCancelAllOrders
	})

	var weightMsgCancelMMOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelMMOrder, &weightMsgCancelMMOrder, nil, func(_ *rand.Rand) {
		weightMsgCancelMMOrder = appparams.DefaultWeightMsgCancelMMOrder
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
			weightMsgCreateRangedPool,
			SimulateMsgCreateRangedPool(ak, bk, k),
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
			weightMsgMMOrder,
			SimulateMsgMMOrder(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelOrder,
			SimulateMsgCancelOrder(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelAllOrders,
			SimulateMsgCancelAllOrders(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelMMOrder,
			SimulateMsgCancelMMOrder(ak, bk, k),
		),
	}
}

func SimulateMsgCreatePair(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		params := k.GetParams(ctx)

		denomA, denomB, found := findNonExistingPair(r, bk, k, ctx)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePair, "all pairs have already been created"), nil, nil
		}

		accs = utils.ShuffleSimAccounts(r, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)
			if params.PairCreationFee.IsAllLTE(spendable) {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePair, "no account to create a pair"), nil, nil
		}

		msg := types.NewMsgCreatePair(simAccount.Address, denomA, denomB)

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

func SimulateMsgCreatePool(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		if uint32(len(k.GetAllPools(ctx))) >= k.GetMaxNumActivePoolsPerPair(ctx) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePool, "all pools have been created"), nil, nil
		}

		params := k.GetParams(ctx)
		minDepositAmt := params.MinInitialDepositAmount

		accs = utils.ShuffleSimAccounts(r, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var pair types.Pair
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			var found bool
			pair, found = findPairToCreatePool(r, k, ctx, spendable)
			if found {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePool, "no account to create a pool"), nil, nil
		}

		depositCoins := sdk.NewCoins(
			sdk.NewCoin(
				pair.BaseCoinDenom,
				utils.RandomInt(r, minDepositAmt, spendable.Sub(params.PoolCreationFee).AmountOf(pair.BaseCoinDenom)),
			),
			sdk.NewCoin(
				pair.QuoteCoinDenom,
				utils.RandomInt(r, minDepositAmt, spendable.Sub(params.PoolCreationFee).AmountOf(pair.QuoteCoinDenom)),
			),
		)

		msg := types.NewMsgCreatePool(simAccount.Address, pair.Id, depositCoins)

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

func SimulateMsgCreateRangedPool(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		if uint32(len(k.GetAllPools(ctx))) >= k.GetMaxNumActivePoolsPerPair(ctx) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePool, "all pools have been created"), nil, nil
		}

		accs = utils.ShuffleSimAccounts(r, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var pair types.Pair
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			var found bool
			pair, found = findPairToCreateRangedPool(r, k, ctx, spendable)
			if found {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePool, "no account to create a pool"), nil, nil
		}

		poolCreationFee := k.GetPoolCreationFee(ctx)
		minDepositAmt := k.GetMinInitialDepositAmount(ctx)
		tickPrec := amm.TickPrecision(k.GetTickPrecision(ctx))
		var (
			x, y                             sdk.Int
			minPrice, maxPrice, initialPrice sdk.Dec
		)
		depositableCoins := spendable.Sub(poolCreationFee)
		for {
			x = utils.RandomInt(r, sdk.ZeroInt(), depositableCoins.AmountOf(pair.QuoteCoinDenom))
			if x.LT(minDepositAmt) {
				y = utils.RandomInt(r, minDepositAmt, depositableCoins.AmountOf(pair.BaseCoinDenom))
			} else {
				y = utils.RandomInt(r, sdk.ZeroInt(), depositableCoins.AmountOf(pair.BaseCoinDenom))
			}
			minPrice = amm.RandomTick(r, utils.ParseDec("0.00001"), utils.ParseDec("1"), int(tickPrec))
			maxPrice = amm.RandomTick(r, minPrice.Mul(utils.ParseDec("1.01")), utils.ParseDec("10000"), int(tickPrec))
			initialPrice = amm.RandomTick(r, minPrice, maxPrice, int(tickPrec))
			pool, err := amm.CreateRangedPool(x, y, minPrice, maxPrice, initialPrice)
			ax, ay := pool.Balances()
			if err == nil && (ax.GTE(minDepositAmt) || ay.GTE(minDepositAmt)) {
				break
			}
		}

		depositCoins := sdk.NewCoins(
			sdk.NewCoin(pair.BaseCoinDenom, y),
			sdk.NewCoin(pair.QuoteCoinDenom, x))
		msg := types.NewMsgCreateRangedPool(
			simAccount.Address, pair.Id, depositCoins,
			minPrice, maxPrice, initialPrice)

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

func SimulateMsgDeposit(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		accs = utils.ShuffleSimAccounts(r, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var pair types.Pair
		var poolId uint64
		var depositCoins sdk.Coins
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			_ = k.IterateAllPools(ctx, func(pool types.Pool) (stop bool, err error) {
				if pool.Disabled {
					return false, nil
				}
				pair, _ = k.GetPair(ctx, pool.PairId)
				poolId = pool.Id
				depositCoins = sdk.NewCoins(
					sdk.NewCoin(
						pair.BaseCoinDenom,
						utils.RandomInt(r, sdk.OneInt(), spendable.AmountOf(pair.BaseCoinDenom))),
					sdk.NewCoin(
						pair.QuoteCoinDenom,
						utils.RandomInt(r, sdk.OneInt(), spendable.AmountOf(pair.QuoteCoinDenom))),
				)
				if depositCoins.IsAllLTE(spendable) {
					skip = false
					return true, nil
				}
				return false, nil
			})
			if !skip {
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeposit, "no account to deposit to pool"), nil, nil
		}

		msg := types.NewMsgDeposit(simAccount.Address, poolId, depositCoins)

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

func SimulateMsgWithdraw(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		accs = utils.ShuffleSimAccounts(r, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var pool types.Pool
		skip := true
	loop:
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)
			for _, coin := range spendable {
				poolId, err := types.ParsePoolCoinDenom(coin.Denom)
				if err != nil {
					continue
				}
				var found bool
				pool, found = k.GetPool(ctx, poolId)
				if found {
					skip = false
					break loop
				}
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWithdraw, "no account to withdraw from pool"), nil, nil
		}

		poolCoin := sdk.NewCoin(pool.PoolCoinDenom, utils.RandomInt(r, sdk.OneInt(), spendable.AmountOf(pool.PoolCoinDenom)))
		msg := types.NewMsgWithdraw(simAccount.Address, pool.Id, poolCoin)

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

func SimulateMsgLimitOrder(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		params := k.GetParams(ctx)

		accs = utils.ShuffleSimAccounts(r, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var pair types.Pair
		var pool types.Pool
		var dir types.OrderDirection
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			var found bool
			pair, pool, dir, found = findPairToMakeLimitOrder(r, k, ctx, spendable)
			if found {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLimitOrder, "no account to make a limit order"), nil, nil
		}

		var minPrice, maxPrice sdk.Dec
		if pair.LastPrice != nil {
			minPrice, maxPrice = minMaxPrice(k, ctx, *pair.LastPrice)
		} else {
			if pool != (types.Pool{}) {
				rx, ry := k.GetPoolBalances(ctx, pool)
				ammPool := pool.AMMPool(rx.Amount, ry.Amount, sdk.Int{})
				minPrice, maxPrice = minMaxPrice(k, ctx, ammPool.Price())
			} else {
				minPrice, maxPrice = utils.ParseDec("0.5"), utils.ParseDec("5.0")
			}
		}
		price := amm.PriceToDownTick(utils.RandomDec(r, minPrice, maxPrice), int(params.TickPrecision))

		minAmt := sdk.MaxInt(
			amm.MinCoinAmount,
			amm.MinCoinAmount.ToDec().QuoRoundUp(price).Ceil().TruncateInt(),
		)
		amt := utils.RandomInt(r, minAmt, minAmt.MulRaw(100))

		var offerCoin sdk.Coin
		var demandCoinDenom string
		switch dir {
		case types.OrderDirectionBuy:
			offerCoin = sdk.NewCoin(pair.QuoteCoinDenom, price.MulInt(amt).Ceil().TruncateInt())
			demandCoinDenom = pair.BaseCoinDenom
		case types.OrderDirectionSell:
			offerCoin = sdk.NewCoin(pair.BaseCoinDenom, amt)
			demandCoinDenom = pair.QuoteCoinDenom
		}
		if offerCoin.Amount.LT(amm.MinCoinAmount) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLimitOrder, "too small offer coin amount"), nil, nil
		}
		if !sdk.NewCoins(offerCoin).IsAllLTE(spendable) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLimitOrder, "insufficient funds"), nil, nil
		}

		lifespan := time.Duration(r.Int63n(int64(params.MaxOrderLifespan)))

		msg := types.NewMsgLimitOrder(
			simAccount.Address, pair.Id, dir, offerCoin, demandCoinDenom, price, amt, lifespan)

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

func SimulateMsgMarketOrder(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		params := k.GetParams(ctx)

		accs = utils.ShuffleSimAccounts(r, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var pair types.Pair
		var dir types.OrderDirection
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			var found bool
			pair, dir, found = findPairToMakeMarketOrder(r, k, ctx, spendable)
			if found {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMarketOrder, "no account to make a market order"), nil, nil
		}

		minPrice, maxPrice := types.PriceLimits(*pair.LastPrice, k.GetMaxPriceLimitRatio(ctx), int(k.GetTickPrecision(ctx)))

		minAmt := sdk.MaxInt(
			amm.MinCoinAmount,
			amm.MinCoinAmount.ToDec().QuoRoundUp(minPrice).Ceil().TruncateInt(),
		)
		amt := utils.RandomInt(r, minAmt, minAmt.MulRaw(100))

		var offerCoin sdk.Coin
		var demandCoinDenom string
		switch dir {
		case types.OrderDirectionBuy:
			offerCoin = sdk.NewCoin(pair.QuoteCoinDenom, maxPrice.MulInt(amt).Ceil().TruncateInt())
			demandCoinDenom = pair.BaseCoinDenom
		case types.OrderDirectionSell:
			offerCoin = sdk.NewCoin(pair.BaseCoinDenom, amt)
			demandCoinDenom = pair.QuoteCoinDenom
		}
		if offerCoin.Amount.LT(amm.MinCoinAmount) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMarketOrder, "too small offer coin amount"), nil, nil
		}
		if !sdk.NewCoins(offerCoin).IsAllLTE(spendable) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMarketOrder, "insufficient funds"), nil, nil
		}

		lifespan := time.Duration(r.Int63n(int64(params.MaxOrderLifespan)))

		msg := types.NewMsgMarketOrder(
			simAccount.Address, pair.Id, dir, offerCoin, demandCoinDenom, amt, lifespan)

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

func SimulateMsgMMOrder(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		accs = utils.ShuffleSimAccounts(r, accs)

		var (
			simAccount      simtypes.Account
			spendable       sdk.Coins
			pair            types.Pair
			canSell, canBuy bool
		)
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			var found bool
			pair, canSell, canBuy, found = findPairToMakeMMOrders(r, k, ctx, simAccount.Address, spendable)
			if found {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMMOrder, "no account to make market making orders"), nil, nil
		}

		var minPrice, maxPrice sdk.Dec
		tickPrec := int(k.GetTickPrecision(ctx))
		if pair.LastPrice == nil {
			minPrice = utils.ParseDec("0.5")
			maxPrice = utils.ParseDec("5.0")
		} else {
			minPrice, maxPrice = types.PriceLimits(*pair.LastPrice, k.GetMaxPriceLimitRatio(ctx), tickPrec)
		}
		midPrice := minPrice.Add(maxPrice).QuoInt64(2)

		maxNumTicks := k.GetMaxNumMarketMakingOrderTicks(ctx)
		ammTickPrec := amm.TickPrecision(tickPrec)
		var (
			maxSellPrice, minSellPrice sdk.Dec
			sellAmt                    = sdk.ZeroInt()
			maxBuyPrice, minBuyPrice   sdk.Dec
			buyAmt                     = sdk.ZeroInt()
		)
		if canSell {
			minSellPrice = ammTickPrec.PriceToDownTick(
				utils.RandomDec(r, midPrice.Mul(utils.ParseDec("0.98")), midPrice.Mul(utils.ParseDec("1.02"))))
			maxSellPrice = ammTickPrec.PriceToDownTick(
				utils.RandomDec(r, minSellPrice, maxPrice))
			sellAmt = utils.RandomInt(r, amm.MinCoinAmount, sdk.NewInt(10_000000))
		}
		if canBuy {
			maxBuyPrice = ammTickPrec.PriceToUpTick(
				utils.RandomDec(r, midPrice.Mul(utils.ParseDec("0.98")), midPrice.Mul(utils.ParseDec("1.02"))))
			minBuyPrice = ammTickPrec.PriceToUpTick(
				utils.RandomDec(r, minPrice, maxBuyPrice))
			buyAmt = utils.RandomInt(r, amm.MinCoinAmount, sdk.NewInt(10_000000))
		}

		offerBaseCoin := sdk.NewInt64Coin(pair.BaseCoinDenom, 0)
		offerQuoteCoin := sdk.NewInt64Coin(pair.QuoteCoinDenom, 0)
		if buyAmt.IsPositive() {
			buyTicks := types.MMOrderTicks(
				types.OrderDirectionBuy, minBuyPrice, maxBuyPrice, buyAmt, int(maxNumTicks), tickPrec)
			for _, tick := range buyTicks {
				offerQuoteCoin = offerQuoteCoin.AddAmount(tick.OfferCoinAmount)
			}
		}
		if sellAmt.IsPositive() {
			sellTicks := types.MMOrderTicks(
				types.OrderDirectionSell, minSellPrice, maxSellPrice, sellAmt, int(maxNumTicks), tickPrec)
			for _, tick := range sellTicks {
				offerBaseCoin = offerBaseCoin.AddAmount(tick.OfferCoinAmount)
			}
		}
		if !sdk.NewCoins(offerBaseCoin, offerQuoteCoin).IsAllLTE(spendable) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMMOrder, "insufficient funds"), nil, nil
		}

		lifespan := time.Duration(r.Int63n(int64(k.GetMaxOrderLifespan(ctx))))

		msg := types.NewMsgMMOrder(
			simAccount.Address, pair.Id,
			maxSellPrice, minSellPrice, sellAmt,
			maxBuyPrice, minBuyPrice, buyAmt,
			lifespan)

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

func SimulateMsgCancelOrder(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var orders []types.Order
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			found := false
			_ = k.IterateOrdersByOrderer(ctx, simAccount.Address, func(order types.Order) (stop bool, err error) {
				pair, _ := k.GetPair(ctx, order.PairId)
				if order.Status.CanBeCanceled() && order.BatchId < pair.CurrentBatchId {
					orders = append(orders, order)
					found = true
					return true, nil
				}
				return false, nil
			})
			if found {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCancelOrder, "no account to cancel an order"), nil, nil
		}

		order := orders[r.Intn(len(orders))]

		msg := types.NewMsgCancelOrder(simAccount.Address, order.PairId, order.Id)

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

func SimulateMsgCancelAllOrders(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		pairIds := map[uint64]struct{}{}
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			found := false
			_ = k.IterateOrdersByOrderer(ctx, simAccount.Address, func(order types.Order) (stop bool, err error) {
				pair, _ := k.GetPair(ctx, order.PairId)
				if order.Status.CanBeCanceled() && order.BatchId < pair.CurrentBatchId {
					pairIds[order.PairId] = struct{}{}
					found = true
				}
				return false, nil
			})
			if found {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCancelOrder, "no account to cancel an order"), nil, nil
		}

		var selectedPairIds []uint64
		for pairId := range pairIds {
			selectedPairIds = append(selectedPairIds, pairId)
		}
		// Sort pair ids since the order of keys in a map is not deterministic.
		sort.SliceStable(selectedPairIds, func(i, j int) bool {
			return selectedPairIds[i] < selectedPairIds[j]
		})
		r.Shuffle(len(selectedPairIds), func(i, j int) {
			selectedPairIds[i], selectedPairIds[j] = selectedPairIds[j], selectedPairIds[i]
		})
		selectedPairIds = selectedPairIds[:r.Intn(len(selectedPairIds))]

		msg := types.NewMsgCancelAllOrders(simAccount.Address, selectedPairIds)

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

func SimulateMsgCancelMMOrder(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		accs = utils.ShuffleSimAccounts(r, accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var pair types.Pair
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			var found bool
			pair, found = findPairToCancelMMOrders(r, k, ctx, simAccount.Address)
			if found {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCancelMMOrder, "no account to cancel market making orders"), nil, nil
		}

		msg := types.NewMsgCancelMMOrder(simAccount.Address, pair.Id)

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

var once sync.Once

func fundAccountsOnce(r *rand.Rand, ctx sdk.Context, bk types.BankKeeper, accs []simtypes.Account) {
	once.Do(func() {
		denoms := []string{"denom1", "denom2", "denom3"}
		maxAmt := sdk.NewInt(1_000_000_000_000_000)
		for _, acc := range accs {
			var coins sdk.Coins
			for _, denom := range denoms {
				coins = coins.Add(sdk.NewCoin(denom, simtypes.RandomAmount(r, maxAmt)))
			}
			if err := bk.MintCoins(ctx, types.ModuleName, coins); err != nil {
				panic(err)
			}
			if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, acc.Address, coins); err != nil {
				panic(err)
			}
		}
	})
}

func findNonExistingPair(r *rand.Rand, bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (string, string, bool) {
	var denoms []string
	bk.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
		denoms = append(denoms, coin.Denom)
		return false
	})
	r.Shuffle(len(denoms), func(i, j int) {
		denoms[i], denoms[j] = denoms[j], denoms[i]
	})

	for _, denomA := range denoms {
		for _, denomB := range denoms {
			if denomA != denomB {
				if _, found := k.GetPairByDenoms(ctx, denomA, denomB); !found {
					return denomA, denomB, true
				}
			}
		}
	}

	return "", "", false
}

func findPairToCreatePool(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, spendable sdk.Coins) (types.Pair, bool) {
	params := k.GetParams(ctx)

	var pairs []types.Pair
	_ = k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool, err error) {
		pairs = append(pairs, pair)
		return false, nil
	})
	r.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

	for _, pair := range pairs {
		found := false // Found a non-disabled pool?
		_ = k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool, err error) {
			if !pool.Disabled {
				found = true
				return true, nil
			}
			return false, nil
		})
		if found {
			continue
		}

		minDepositCoins := sdk.NewCoins(
			sdk.NewCoin(pair.BaseCoinDenom, params.MinInitialDepositAmount),
			sdk.NewCoin(pair.QuoteCoinDenom, params.MinInitialDepositAmount),
		)
		if minDepositCoins.Add(params.PoolCreationFee...).IsAllLTE(spendable) {
			return pair, true
		}
	}

	return types.Pair{}, false
}

func findPairToCreateRangedPool(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, spendable sdk.Coins) (types.Pair, bool) {
	var hasNeg bool
	spendable, hasNeg = spendable.SafeSub(k.GetPoolCreationFee(ctx))
	if hasNeg {
		return types.Pair{}, false
	}

	var pairs []types.Pair
	_ = k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool, err error) {
		pairs = append(pairs, pair)
		return false, nil
	})
	r.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

	minDepositAmt := k.GetMinInitialDepositAmount(ctx)
	for _, pair := range pairs {
		if spendable.AmountOf(pair.BaseCoinDenom).GTE(minDepositAmt) &&
			spendable.AmountOf(pair.QuoteCoinDenom).GTE(minDepositAmt) {
			return pair, true
		}
	}

	return types.Pair{}, false
}

func findPairToMakeLimitOrder(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, spendable sdk.Coins) (types.Pair, types.Pool, types.OrderDirection, bool) {
	var pairs []types.Pair
	_ = k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool, err error) {
		pairs = append(pairs, pair)
		return false, nil
	})
	r.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

	for _, pair := range pairs {
		var resPool types.Pool
		_ = k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool, err error) {
			if !pool.Disabled {
				resPool = pool
				return true, nil
			}
			return false, nil
		})

		dirs := []types.OrderDirection{types.OrderDirectionBuy, types.OrderDirectionSell}
		r.Shuffle(len(dirs), func(i, j int) {
			dirs[i], dirs[j] = dirs[j], dirs[i]
		})

		for _, dir := range dirs {
			var minOfferCoinAmt sdk.Coin
			switch dir {
			case types.OrderDirectionBuy:
				minOfferCoinAmt = sdk.NewCoin(pair.QuoteCoinDenom, amm.MinCoinAmount)
			case types.OrderDirectionSell:
				minOfferCoinAmt = sdk.NewCoin(pair.BaseCoinDenom, amm.MinCoinAmount)
			}

			if sdk.NewCoins(minOfferCoinAmt).IsAllLTE(spendable) {
				return pair, resPool, dir, true
			}
		}
	}

	return types.Pair{}, types.Pool{}, 0, false
}

func findPairToMakeMarketOrder(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, spendable sdk.Coins) (types.Pair, types.OrderDirection, bool) {
	var pairs []types.Pair
	_ = k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool, err error) {
		pairs = append(pairs, pair)
		return false, nil
	})
	r.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

	for _, pair := range pairs {
		if pair.LastPrice == nil {
			continue
		}

		dirs := []types.OrderDirection{types.OrderDirectionBuy, types.OrderDirectionSell}
		r.Shuffle(len(dirs), func(i, j int) {
			dirs[i], dirs[j] = dirs[j], dirs[i]
		})

		for _, dir := range dirs {
			var minOfferCoinAmt sdk.Coin
			switch dir {
			case types.OrderDirectionBuy:
				minOfferCoinAmt = sdk.NewCoin(pair.QuoteCoinDenom, amm.MinCoinAmount)
			case types.OrderDirectionSell:
				minOfferCoinAmt = sdk.NewCoin(pair.BaseCoinDenom, amm.MinCoinAmount)
			}

			if sdk.NewCoins(minOfferCoinAmt).IsAllLTE(spendable) {
				return pair, dir, true
			}
		}
	}

	return types.Pair{}, 0, false
}

func findPairToMakeMMOrders(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, orderer sdk.AccAddress, spendable sdk.Coins) (pair types.Pair, canSell, canBuy, found bool) {
	pairs := k.GetAllPairs(ctx)
	r.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

loop:
	for _, pair = range pairs {
		// Skip pairs that already have market making orders from this orderer.
		index, f := k.GetMMOrderIndex(ctx, orderer, pair.Id)
		if f {
			for _, orderId := range index.OrderIds {
				order, f := k.GetOrder(ctx, pair.Id, orderId)
				if f && (order.BatchId == pair.CurrentBatchId) {
					continue loop
				}
			}
		}

		baseCoinAmt := spendable.AmountOf(pair.BaseCoinDenom)
		quoteCoinAmt := spendable.AmountOf(pair.QuoteCoinDenom)
		if baseCoinAmt.GTE(sdk.NewInt(100_000000)) {
			canSell = true
		} else {
			canSell = false
		}
		if quoteCoinAmt.GTE(sdk.NewInt(100_000000)) {
			canBuy = true
		} else {
			canBuy = false
		}
		if canBuy || canSell {
			found = true
			return
		}
	}

	return
}

func findPairToCancelMMOrders(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, orderer sdk.AccAddress) (pair types.Pair, found bool) {
	pairs := k.GetAllPairs(ctx)
	r.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

loop:
	for _, pair = range pairs {
		index, f := k.GetMMOrderIndex(ctx, orderer, pair.Id)
		if !f {
			continue
		}

		for _, orderId := range index.OrderIds {
			order, f := k.GetOrder(ctx, pair.Id, orderId)
			if f && (order.BatchId == pair.CurrentBatchId) {
				continue loop
			}
		}

		found = true
		return
	}
	return
}

func minMaxPrice(k keeper.Keeper, ctx sdk.Context, lastPrice sdk.Dec) (sdk.Dec, sdk.Dec) {
	params := k.GetParams(ctx)
	tickPrec := int(params.TickPrecision)
	maxPrice := amm.PriceToDownTick(lastPrice.Mul(sdk.OneDec().Add(params.MaxPriceLimitRatio)), tickPrec)
	minPrice := amm.PriceToUpTick(lastPrice.Mul(sdk.OneDec().Sub(params.MaxPriceLimitRatio)), tickPrec)
	return minPrice, maxPrice
}
