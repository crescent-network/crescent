package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v5/app/params"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgCreateMarket     = "op_weight_msg_create_market"
	OpWeightMsgPlaceLimitOrder  = "op_weight_msg_place_limit_order"
	OpWeightMsgPlaceMarketOrder = "op_weight_msg_place_market_order"

	DefaultWeightMsgCreateMarket     = 10
	DefaultWeightMsgPlaceLimitOrder  = 90
	DefaultWeightMsgPlaceMarketOrder = 90
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
		weightMsgCreateMarket     int
		weightMsgPlaceLimitOrder  int
		weightMsgPlaceMarketOrder int
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateMarket, &weightMsgCreateMarket, nil, func(_ *rand.Rand) {
		weightMsgCreateMarket = DefaultWeightMsgCreateMarket
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceLimitOrder, &weightMsgPlaceLimitOrder, nil, func(_ *rand.Rand) {
		weightMsgPlaceLimitOrder = DefaultWeightMsgPlaceLimitOrder
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceMarketOrder, &weightMsgPlaceMarketOrder, nil, func(_ *rand.Rand) {
		weightMsgPlaceMarketOrder = DefaultWeightMsgPlaceMarketOrder
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
			weightMsgPlaceMarketOrder,
			SimulateMsgPlaceMarketOrder(ak, bk, k),
		),
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
				types.ModuleName, types.TypeMsgPlaceLimitOrder, "unable to place market order"), nil, nil
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
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg legacytx.LegacyMsg, found bool) {
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
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg legacytx.LegacyMsg, found bool) {
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
				price = utils.RandomDec(r, utils.ParseDec("0.05"), utils.ParseDec("500"))
				price = types.PriceAtTick(types.TickAtPrice(price))
			}
			if r.Float64() <= 0.5 { // 50% chance to sell
				if balance := spendable.AmountOf(market.BaseDenom); balance.GT(sdk.NewInt(100_000000)) {
					qty := utils.RandomInt(r, sdk.NewInt(100), sdk.NewInt(100_000000))
					msg = types.NewMsgPlaceLimitOrder(
						acc.Address, market.Id, false, price, qty, lifespan)
					return acc, msg, true
				}
			}
			if balance := spendable.AmountOf(market.QuoteDenom); balance.GT(price.MulInt64(100_000000).TruncateInt()) {
				qty := utils.RandomInt(r, sdk.NewInt(100), sdk.NewInt(100_000000))
				msg = types.NewMsgPlaceLimitOrder(
					acc.Address, market.Id, true, price, qty, lifespan)
				return acc, msg, true
			}
		}
	}
	return acc, msg, false
}

func findMsgPlaceMarketOrderParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg legacytx.LegacyMsg, found bool) {
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
				if balance := spendable.AmountOf(market.BaseDenom); balance.GT(sdk.NewInt(100_000000)) {
					qty := utils.RandomInt(r, sdk.NewInt(100), sdk.NewInt(100_000000))
					msg = types.NewMsgPlaceMarketOrder(
						acc.Address, market.Id, false, qty)
					return acc, msg, true
				}
			}
			marketState := k.MustGetMarketState(ctx, market.Id)
			if marketState.LastPrice == nil {
				continue
			}
			estBuyPrice := marketState.LastPrice.Mul(utils.ParseDec("1.5")) // 150%
			if balance := spendable.AmountOf(market.QuoteDenom); balance.GT(estBuyPrice.MulInt64(100_000000).TruncateInt()) {
				qty := utils.RandomInt(r, sdk.NewInt(100), sdk.NewInt(100_000000))
				msg = types.NewMsgPlaceMarketOrder(
					acc.Address, market.Id, true, qty)
				return acc, msg, true
			}
		}
	}
	return acc, msg, false
}
