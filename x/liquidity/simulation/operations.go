package simulation

import (
	"math/rand"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	squadappparams "github.com/cosmosquad-labs/squad/app/params"
	"github.com/cosmosquad-labs/squad/x/liquidity/keeper"
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
	bk types.BankKeeper, k keeper.Keeper,
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
		weightMsgDeposit = squadappparams.DefaultWeightMsgDeposit
	})

	var weightMsgWithdraw int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdraw, &weightMsgWithdraw, nil, func(_ *rand.Rand) {
		weightMsgWithdraw = squadappparams.DefaultWeightMsgWithdraw
	})

	var weightMsgLimitOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgLimitOrder, &weightMsgLimitOrder, nil, func(_ *rand.Rand) {
		weightMsgLimitOrder = squadappparams.DefaultWeightMsgLimitOrder
	})

	var weightMsgMarketOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgMarketOrder, &weightMsgMarketOrder, nil, func(_ *rand.Rand) {
		weightMsgMarketOrder = squadappparams.DefaultWeightMsgMarketOrder
	})

	var weightMsgCancelOrder int
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelOrder, &weightMsgCancelOrder, nil, func(_ *rand.Rand) {
		weightMsgCancelOrder = squadappparams.DefaultWeightMsgCancelOrder
	})

	var weightMsgCancelAllOrders int
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelAllOrders, &weightMsgCancelAllOrders, nil, func(_ *rand.Rand) {
		weightMsgCancelAllOrders = squadappparams.DefaultWeightMsgCancelAllOrders
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

		accs = shuffleAccs(accs)

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
			TxGen:           squadappparams.MakeTestEncodingConfig().TxConfig,
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

func SimulateMsgCreatePool(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		params := k.GetParams(ctx)
		minDepositAmt := params.MinInitialDepositAmount

		accs = shuffleAccs(accs)

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
				randomInt(r, minDepositAmt, spendable.Sub(params.PoolCreationFee).AmountOf(pair.BaseCoinDenom)),
			),
			sdk.NewCoin(
				pair.QuoteCoinDenom,
				randomInt(r, minDepositAmt, spendable.Sub(params.PoolCreationFee).AmountOf(pair.QuoteCoinDenom)),
			),
		)

		msg := types.NewMsgCreatePool(simAccount.Address, pair.Id, depositCoins)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           squadappparams.MakeTestEncodingConfig().TxConfig,
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

func SimulateMsgDeposit(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		accs = shuffleAccs(accs)

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
						randomInt(r, sdk.OneInt(), spendable.AmountOf(pair.BaseCoinDenom))),
					sdk.NewCoin(
						pair.QuoteCoinDenom,
						randomInt(r, sdk.OneInt(), spendable.AmountOf(pair.QuoteCoinDenom))),
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
			TxGen:           squadappparams.MakeTestEncodingConfig().TxConfig,
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

func SimulateMsgWithdraw(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		accs = shuffleAccs(accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var poolId uint64
		skip := true
	loop:
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)
			for _, coin := range spendable {
				if poolId = types.ParsePoolCoinDenom(coin.Denom); poolId != 0 {
					skip = false
					break loop
				}
			}
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWithdraw, "no account to withdraw from pool"), nil, nil
		}

		pool, _ := k.GetPool(ctx, poolId)
		poolCoin := sdk.NewCoin(pool.PoolCoinDenom, randomInt(r, sdk.OneInt(), spendable.AmountOf(pool.PoolCoinDenom)))
		msg := types.NewMsgWithdraw(simAccount.Address, poolId, poolCoin)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           squadappparams.MakeTestEncodingConfig().TxConfig,
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

func SimulateMsgLimitOrder(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		params := k.GetParams(ctx)

		accs = shuffleAccs(accs)

		var simAccount simtypes.Account
		var spendable sdk.Coins
		var pair types.Pair
		var pool types.Pool
		var dir types.SwapDirection
		skip := true
		for _, simAccount = range accs {
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			var found bool
			pair, pool, dir, found = findPairToMakeOrder(r, k, ctx, spendable)
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
			rx, ry := k.GetPoolBalance(ctx, pool, pair)
			poolInfo := types.NewPoolInfo(rx, ry, sdk.Int{})
			minPrice, maxPrice = minMaxPrice(k, ctx, poolInfo.Price())
		}
		price := types.PriceToTick(randomDec(r, minPrice, maxPrice), int(params.TickPrecision))

		amt := randomInt(r, types.MinCoinAmount, sdk.NewInt(1000000))

		var offerCoin sdk.Coin
		var demandCoinDenom string
		switch dir {
		case types.SwapDirectionBuy:
			offerCoin = sdk.NewCoin(pair.QuoteCoinDenom, price.MulInt(amt).Ceil().TruncateInt())
			demandCoinDenom = pair.BaseCoinDenom
		case types.SwapDirectionSell:
			offerCoin = sdk.NewCoin(pair.BaseCoinDenom, amt)
			demandCoinDenom = pair.QuoteCoinDenom
		}
		if offerCoin.Amount.LT(types.MinCoinAmount) {
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
			TxGen:           squadappparams.MakeTestEncodingConfig().TxConfig,
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

func SimulateMsgMarketOrder(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMarketOrder, ""), nil, nil
		//txCtx := simulation.OperationInput{
		//	R:               r,
		//	App:             app,
		//	TxGen:           squadappparams.MakeTestEncodingConfig().TxConfig,
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

func SimulateMsgCancelOrder(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCancelOrder, ""), nil, nil
		//txCtx := simulation.OperationInput{
		//	R:               r,
		//	App:             app,
		//	TxGen:           squadappparams.MakeTestEncodingConfig().TxConfig,
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

func SimulateMsgCancelAllOrders(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fundAccountsOnce(r, ctx, bk, accs)

		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCancelAllOrders, ""), nil, nil
		//txCtx := simulation.OperationInput{
		//	R:               r,
		//	App:             app,
		//	TxGen:           squadappparams.MakeTestEncodingConfig().TxConfig,
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

// randomInt returns an integer within a range [min, max].
func randomInt(r *rand.Rand, min, max sdk.Int) sdk.Int {
	return sdk.MaxInt(min, min.Add(simtypes.RandomAmount(r, max.Sub(min))))
}

// randomDec returns a decimal within a range [min, max].
func randomDec(r *rand.Rand, min, max sdk.Dec) sdk.Dec {
	return sdk.MaxDec(min, min.Add(simtypes.RandomDecAmount(r, max.Sub(min))))
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

func shuffleAccs(accs []simtypes.Account) []simtypes.Account {
	accs2 := make([]simtypes.Account, len(accs))
	copy(accs2, accs)
	return accs2
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

func findPairToMakeOrder(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, spendable sdk.Coins) (types.Pair, types.Pool, types.SwapDirection, bool) {
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
		found := false // Found a non-disabled pool?
		_ = k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool, err error) {
			if !pool.Disabled {
				resPool = pool
				found = true
				return true, nil
			}
			return false, nil
		})
		if !found {
			continue
		}

		dirs := []types.SwapDirection{types.SwapDirectionBuy, types.SwapDirectionSell}
		r.Shuffle(len(dirs), func(i, j int) {
			dirs[i], dirs[j] = dirs[j], dirs[i]
		})

		for _, dir := range dirs {
			var minOfferCoinAmt sdk.Coin
			switch dir {
			case types.SwapDirectionBuy:
				minOfferCoinAmt = sdk.NewCoin(pair.QuoteCoinDenom, types.MinCoinAmount)
			case types.SwapDirectionSell:
				minOfferCoinAmt = sdk.NewCoin(pair.BaseCoinDenom, types.MinCoinAmount)
			}

			if sdk.NewCoins(minOfferCoinAmt).IsAllLTE(spendable) {
				return pair, resPool, dir, true
			}
		}
	}

	return types.Pair{}, types.Pool{}, 0, false
}

func minMaxPrice(k keeper.Keeper, ctx sdk.Context, lastPrice sdk.Dec) (sdk.Dec, sdk.Dec) {
	params := k.GetParams(ctx)
	tickPrec := int(params.TickPrecision)
	minPrice := lastPrice.Mul(sdk.OneDec().Sub(params.MaxPriceLimitRatio))
	minPriceTick := types.PriceToTick(minPrice, tickPrec)
	if !minPriceTick.Equal(minPrice) {
		minPrice = types.UpTick(minPriceTick, tickPrec)
	}
	maxPrice := types.PriceToTick(lastPrice.Mul(sdk.OneDec().Add(params.MaxPriceLimitRatio)), tickPrec)
	return minPrice, maxPrice
}
