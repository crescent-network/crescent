package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v5/app/params"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgCreateMarket      = "op_weight_msg_create_market"
	OpWeightMsgPlaceLimitOrder   = "op_weight_msg_place_limit_order"
	OpWeightMsgPlaceMMLimitOrder = "op_weight_msg_place_mm_limit_order"
	OpWeightMsgPlaceMarketOrder  = "op_weight_msg_place_market_order"
	OpWeightMsgCancelOrder       = "op_weight_msg_cancel_order"
	OpWeightMsgCancelAllOrders   = "op_weight_msg_cancel_all_orders"
	OpWeightMsgSwapExactAmountIn = "op_weight_msg_swap_exact_amount_in"

	DefaultWeightMsgCreateMarket      = 10
	DefaultWeightMsgPlaceLimitOrder   = 90
	DefaultWeightMsgPlaceMMLimitOrder = 50
	DefaultWeightMsgPlaceMarketOrder  = 90
	DefaultWeightMsgCancelOrder       = 20
	DefaultWeightMsgCancelAllOrders   = 10
	DefaultWeightMsgSwapExactAmountIn = 80
)

var (
	gas  = uint64(20000000)
	fees = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateMarket      int
		weightMsgPlaceLimitOrder   int
		weightMsgPlaceMMLimitOrder int
		weightMsgPlaceMarketOrder  int
		weightMsgCancelOrder       int
		weightMsgCancelAllOrders   int
		weightMsgSwapExactAmountIn int
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateMarket, &weightMsgCreateMarket, nil, func(_ *rand.Rand) {
		weightMsgCreateMarket = DefaultWeightMsgCreateMarket
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceLimitOrder, &weightMsgPlaceLimitOrder, nil, func(_ *rand.Rand) {
		weightMsgPlaceLimitOrder = DefaultWeightMsgPlaceLimitOrder
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceMMLimitOrder, &weightMsgPlaceMMLimitOrder, nil, func(_ *rand.Rand) {
		weightMsgPlaceMMLimitOrder = DefaultWeightMsgPlaceMMLimitOrder
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceMarketOrder, &weightMsgPlaceMarketOrder, nil, func(_ *rand.Rand) {
		weightMsgPlaceMarketOrder = DefaultWeightMsgPlaceMarketOrder
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelOrder, &weightMsgCancelOrder, nil, func(_ *rand.Rand) {
		weightMsgCancelOrder = DefaultWeightMsgCancelOrder
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelAllOrders, &weightMsgCancelAllOrders, nil, func(_ *rand.Rand) {
		weightMsgCancelAllOrders = DefaultWeightMsgCancelAllOrders
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgSwapExactAmountIn, &weightMsgSwapExactAmountIn, nil, func(_ *rand.Rand) {
		weightMsgSwapExactAmountIn = DefaultWeightMsgSwapExactAmountIn
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateMarket,
			SimulateMsgCreateMarket(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgPlaceLimitOrder,
			SimulateMsgPlaceLimitOrder(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgPlaceMMLimitOrder,
			SimulateMsgPlaceMMLimitOrder(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgPlaceMarketOrder,
			SimulateMsgPlaceMarketOrder(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelOrder,
			SimulateMsgCancelOrder(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelAllOrders,
			SimulateMsgCancelAllOrders(ak, bk, k),
		),
		// XXX
		//simulation.NewWeightedOperation(
		//	weightMsgSwapExactAmountIn,
		//	SimulateMsgSwapExactAmountIn(ak, bk, k),
		//),
	}
}

func SimulateMsgCreateMarket(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgCreateMarketParams(r, accs, bk, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreateMarket, "unable to create market"), nil, nil
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

func SimulateMsgPlaceLimitOrder(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgPlaceLimitOrderParams(r, accs, bk, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgPlaceLimitOrder, "unable to place limit order"), nil, nil
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

func SimulateMsgPlaceMMLimitOrder(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgPlaceMMLimitOrderParams(r, accs, bk, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgPlaceMMLimitOrder, "unable to place mm limit order"), nil, nil
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

func SimulateMsgPlaceMarketOrder(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgPlaceMarketOrderParams(r, accs, bk, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgPlaceMarketOrder, "unable to place market order"), nil, nil
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

func SimulateMsgCancelOrder(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgCancelOrderParams(r, accs, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCancelOrder, "unable to cancel order"), nil, nil
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

func SimulateMsgCancelAllOrders(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgCancelAllOrdersParams(r, accs, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCancelAllOrders, "unable to cancel all orders"), nil, nil
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

func SimulateMsgSwapExactAmountIn(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgSwapExactAmountInParams(r, accs, bk, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSwapExactAmountIn, "unable to swap exact amount in"), nil, nil
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

func findMsgCreateMarketParams(r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgCreateMarket, found bool) {
	var allDenoms []string
	bk.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
		allDenoms = append(allDenoms, coin.Denom)
		return false
	})
	r.Shuffle(len(allDenoms), func(i, j int) {
		allDenoms[i], allDenoms[j] = allDenoms[j], allDenoms[i]
	})
	for _, denomA := range allDenoms {
		for _, denomB := range allDenoms {
			if denomA != denomB {
				if _, found := k.GetMarketIdByDenoms(ctx, denomA, denomB); !found {
					acc, _ = simtypes.RandomAcc(r, accs)
					spendable := bk.SpendableCoins(ctx, acc.Address)
					if !spendable.IsAllGTE(k.GetMarketCreationFee(ctx)) {
						continue
					}
					msg = types.NewMsgCreateMarket(acc.Address, denomA, denomB)
					return acc, msg, true
				}
			}
		}
	}
	return acc, msg, false
}

func findMsgPlaceLimitOrderParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgPlaceLimitOrder, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	var markets []types.Market
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		markets = append(markets, market)
		return false
	})
	r.Shuffle(len(markets), func(i, j int) {
		markets[i], markets[j] = markets[j], markets[i]
	})
	lifespan := time.Duration(1+r.Intn(8)) * time.Hour
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		for _, market := range markets {
			marketState := k.MustGetMarketState(ctx, market.Id)
			var price sdk.Dec
			if marketState.LastPrice != nil {
				price = types.PriceAtTick(
					types.TickAtPrice(*marketState.LastPrice) + int32(r.Intn(1000)) - 500)
			} else {
				price = utils.RandomDec(r, utils.ParseDec("0.1"), utils.ParseDec("10"))
				price = types.PriceAtTick(types.TickAtPrice(price))
			}
			if r.Float64() <= 0.5 { // 50% chance to sell
				if balance := spendable.AmountOf(market.BaseDenom); balance.GT(sdk.NewInt(100_000000)) {
					qty := utils.RandomDec(r, sdk.NewDec(100), sdk.NewDec(100_000000)).TruncateDec()
					msg = types.NewMsgPlaceLimitOrder(
						acc.Address, market.Id, false, price, qty, lifespan)
					return acc, msg, true
				}
			}
			if balance := spendable.AmountOf(market.QuoteDenom); balance.GT(price.MulInt64(100_000000).TruncateInt()) {
				qty := utils.RandomDec(r, sdk.NewDec(100), sdk.NewDec(100_000000)).TruncateDec()
				msg = types.NewMsgPlaceLimitOrder(
					acc.Address, market.Id, true, price, qty, lifespan)
				return acc, msg, true
			}
		}
	}
	return acc, msg, false
}

func findMsgPlaceMMLimitOrderParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgPlaceMMLimitOrder, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	var markets []types.Market
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		markets = append(markets, market)
		return false
	})
	r.Shuffle(len(markets), func(i, j int) {
		markets[i], markets[j] = markets[j], markets[i]
	})
	lifespan := time.Duration(1+r.Intn(8)) * time.Hour
	maxNumMMOrders := k.GetMaxNumMMOrders(ctx)
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		for _, market := range markets {
			numMMOrders, _ := k.GetNumMMOrders(ctx, acc.Address, market.Id)
			if numMMOrders >= maxNumMMOrders {
				continue
			}
			marketState := k.MustGetMarketState(ctx, market.Id)
			var price sdk.Dec
			if marketState.LastPrice != nil {
				price = types.PriceAtTick(
					types.TickAtPrice(*marketState.LastPrice) + int32(r.Intn(1000)) - 500)
			} else {
				price = utils.RandomDec(r, utils.ParseDec("0.1"), utils.ParseDec("10"))
				price = types.PriceAtTick(types.TickAtPrice(price))
			}
			if r.Float64() <= 0.5 { // 50% chance to sell
				if balance := spendable.AmountOf(market.BaseDenom); balance.GT(sdk.NewInt(100_000000)) {
					qty := utils.RandomDec(r, sdk.NewDec(100), sdk.NewDec(100_000000)).TruncateDec()
					msg = types.NewMsgPlaceMMLimitOrder(
						acc.Address, market.Id, false, price, qty, lifespan)
					return acc, msg, true
				}
			}
			if balance := spendable.AmountOf(market.QuoteDenom); balance.GT(price.MulInt64(100_000000).TruncateInt()) {
				qty := utils.RandomDec(r, sdk.NewDec(100), sdk.NewDec(100_000000)).TruncateDec()
				msg = types.NewMsgPlaceMMLimitOrder(
					acc.Address, market.Id, true, price, qty, lifespan)
				return acc, msg, true
			}
		}
	}
	return acc, msg, false
}

func findMsgPlaceMarketOrderParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgPlaceMarketOrder, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	var markets []types.Market
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		markets = append(markets, market)
		return false
	})
	r.Shuffle(len(markets), func(i, j int) {
		markets[i], markets[j] = markets[j], markets[i]
	})
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		for _, market := range markets {
			if r.Float64() <= 0.5 { // 50% chance to sell
				if balance := spendable.AmountOf(market.BaseDenom); balance.GT(sdk.NewInt(1_000000)) {
					qty := utils.RandomDec(r, sdk.NewDec(100), sdk.NewDec(1_000000)).TruncateDec()
					msg = types.NewMsgPlaceMarketOrder(
						acc.Address, market.Id, false, qty)
					return acc, msg, true
				}
			}
			marketState := k.MustGetMarketState(ctx, market.Id)
			if marketState.LastPrice == nil {
				continue
			}
			qty := utils.RandomDec(r, sdk.NewDec(100), sdk.NewDec(1_000000)).TruncateDec()
			cacheCtx, _ := ctx.CacheContext()
			if _, _, err := k.PlaceMarketOrder(cacheCtx, market.Id, acc.Address, true, qty); err != nil {
				continue
			}
			msg = types.NewMsgPlaceMarketOrder(acc.Address, market.Id, true, qty)
			return acc, msg, true
		}
	}
	return acc, msg, false
}

func findMsgCancelOrderParams(
	r *rand.Rand, accs []simtypes.Account, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgCancelOrder, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	for _, acc = range accs {
		var orders []types.Order
		k.IterateOrdersByOrderer(ctx, acc.Address, func(order types.Order) (stop bool) {
			if order.MsgHeight < ctx.BlockHeight() {
				orders = append(orders, order)
			}
			return false
		})
		if len(orders) > 0 {
			order := orders[r.Intn(len(orders))]
			msg = types.NewMsgCancelOrder(acc.Address, order.Id)
			return acc, msg, true
		}
	}
	return acc, nil, false
}

func findMsgCancelAllOrdersParams(
	r *rand.Rand, accs []simtypes.Account, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgCancelAllOrders, found bool) {
	acc, _ = simtypes.RandomAcc(r, accs)
	var markets []types.Market
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		markets = append(markets, market)
		return false
	})
	if len(markets) == 0 {
		return acc, nil, false
	}
	// CancelAllOrders will succeed even if the orderer has no orders in the market.
	// So we just choose random market.
	market := markets[r.Intn(len(markets))]
	msg = types.NewMsgCancelAllOrders(acc.Address, market.Id)
	return acc, msg, true
}

func findMsgSwapExactAmountInParams(
	r *rand.Rand, accs []simtypes.Account, bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgSwapExactAmountIn, found bool) {
	var allDenoms []string
	bk.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
		allDenoms = append(allDenoms, coin.Denom)
		return false
	})
	accs = utils.ShuffleSimAccounts(r, accs)
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		// We shuffle denoms every time for better randomization of candidate
		// denom pair.
		r.Shuffle(len(allDenoms), func(i, j int) {
			allDenoms[i], allDenoms[j] = allDenoms[j], allDenoms[i]
		})
		for _, denomIn := range allDenoms {
			if !spendable.AmountOf(denomIn).GTE(sdk.NewInt(1_000000)) {
				continue
			}
			input := sdk.NewDecCoin(denomIn, utils.RandomInt(r, sdk.NewInt(10000), sdk.NewInt(1_000000)))
			for _, denomOut := range allDenoms {
				querier := keeper.Querier{Keeper: k}
				resp, err := querier.BestSwapExactAmountInRoutes(sdk.WrapSDKContext(ctx), &types.QueryBestSwapExactAmountInRoutesRequest{
					Input:       input.String(),
					OutputDenom: denomOut,
				})
				if err != nil {
					continue
				}
				// If there's no error than the output amount always be positive.
				msg = types.NewMsgSwapExactAmountIn(acc.Address, resp.Routes, input, resp.Output)
				return acc, msg, true
			}
		}
	}
	return acc, nil, false
}
