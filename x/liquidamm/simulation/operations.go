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
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgMintShare = "op_weight_msg_mint_share"
	OpWeightMsgBurnShare = "op_weight_msg_burn_share"
	OpWeightMsgPlaceBid  = "op_weight_msg_place_bid"

	DefaultWeightMsgMintShare int = 70
	DefaultWeightMsgBurnShare int = 20
	DefaultWeightMsgPlaceBid  int = 50
)

var (
	gas  = uint64(20000000)
	fees = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak types.AccountKeeper, bk types.BankKeeper, ammKeeper types.AMMKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgMintShare int
		weightMsgBurnShare int
		weightMsgPlaceBid  int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgMintShare, &weightMsgMintShare, nil, func(_ *rand.Rand) {
		weightMsgMintShare = DefaultWeightMsgMintShare
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgBurnShare, &weightMsgBurnShare, nil, func(_ *rand.Rand) {
		weightMsgBurnShare = DefaultWeightMsgBurnShare
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceBid, &weightMsgPlaceBid, nil, func(_ *rand.Rand) {
		weightMsgPlaceBid = DefaultWeightMsgPlaceBid
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgMintShare,
			SimulateMsgMintShare(ak, bk, ammKeeper, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBurnShare,
			SimulateMsgBurnShare(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgPlaceBid,
			SimulateMsgPlaceBid(ak, bk, k),
		),
	}
}

// SimulateMsgMintShare generates a MsgMintShare with random values
func SimulateMsgMintShare(ak types.AccountKeeper, bk types.BankKeeper, ammKeeper types.AMMKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgMintShareParams(r, accs, bk, ammKeeper, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgMintShare, "unable to mint share"), nil, nil
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

// SimulateMsgBurnShare generates a MsgBurnShare with random values
func SimulateMsgBurnShare(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgBurnShareParams(r, accs, bk, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgBurnShare, "unable to burn share"), nil, nil
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

// SimulateMsgPlaceBid generates a MsgPlaceBid with random values
func SimulateMsgPlaceBid(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgPlaceBidParams(r, accs, bk, k, ctx)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgPlaceBid, "unable to place bid"), nil, nil
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

func findMsgMintShareParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, ak types.AMMKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgMintShare, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	publicPositions := k.GetAllPublicPositions(ctx)
	minAmt := sdk.NewInt(10000)
	maxAmt := sdk.NewInt(1_000000)
	numInsufficientBalances := 0
	numZeroLiquidity := 0
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		utils.Shuffle(r, publicPositions)
		for _, publicPosition := range publicPositions {
			pool, found := ak.GetPool(ctx, publicPosition.PoolId)
			if !found { // sanity check
				panic("pool not found")
			}
			if !spendable.AmountOf(pool.Denom0).GTE(maxAmt) ||
				!spendable.AmountOf(pool.Denom1).GTE(maxAmt) {
				numInsufficientBalances++
				continue
			}
			poolState := ak.MustGetPoolState(ctx, pool.Id)
			lowerSqrtPrice := ammtypes.SqrtPriceAtTick(publicPosition.LowerTick)
			upperSqrtPrice := ammtypes.SqrtPriceAtTick(publicPosition.UpperTick)
			amt0 := utils.RandomInt(r, minAmt, maxAmt)
			amt1 := utils.RandomInt(r, minAmt, maxAmt)
			liquidity := ammtypes.LiquidityForAmounts(
				poolState.CurrentSqrtPrice, lowerSqrtPrice, upperSqrtPrice, amt0, amt1)
			if liquidity.IsZero() {
				numZeroLiquidity++
				continue
			}
			amt0, amt1 = ammtypes.AmountsForLiquidity(
				poolState.CurrentSqrtPrice, lowerSqrtPrice, upperSqrtPrice, liquidity)
			desiredAmt := sdk.NewCoins(
				sdk.NewCoin(pool.Denom0, amt0), sdk.NewCoin(pool.Denom1, amt1))
			msg = types.NewMsgMintShare(acc.Address, publicPosition.Id, desiredAmt)
			return acc, msg, true
		}
	}
	return acc, nil, false
}

func findMsgBurnShareParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgBurnShare, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		utils.Shuffle(r, spendable)
		for _, coin := range spendable {
			publicPositionId, err := types.ParseShareDenom(coin.Denom)
			if err != nil { // not a public position share denom
				continue
			}
			publicPosition, found := k.GetPublicPosition(ctx, publicPositionId)
			if !found { // sanity check
				panic("public position not found")
			}
			ammPosition := k.MustGetAMMPosition(ctx, publicPosition)
			// [min(10000, coin.Amount), coin.Amount]
			share := sdk.NewCoin(
				coin.Denom, utils.RandomInt(
					r,
					utils.MinInt(sdk.NewInt(10000), coin.Amount),
					coin.Amount.Add(utils.OneInt)))
			shareSupply := bk.GetSupply(ctx, share.Denom).Amount
			var prevWinningBidShareAmt sdk.Int
			auction, found := k.GetPreviousRewardsAuction(ctx, publicPosition)
			if found && auction.WinningBid != nil {
				prevWinningBidShareAmt = auction.WinningBid.Share.Amount
			} else {
				prevWinningBidShareAmt = utils.ZeroInt
			}
			if removedLiquidity := types.CalculateRemovedLiquidity(
				share.Amount, shareSupply,
				ammPosition.Liquidity, prevWinningBidShareAmt); removedLiquidity.IsZero() {
				continue
			}
			msg = types.NewMsgBurnShare(acc.Address, publicPositionId, share)
			return acc, msg, true
		}
	}
	return acc, nil, false
}

func findMsgPlaceBidParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgPlaceBid, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		utils.Shuffle(r, spendable)
		for _, coin := range spendable {
			publicPositionId, err := types.ParseShareDenom(coin.Denom)
			if err != nil { // not a public position share denom
				continue
			}
			auction, found := k.GetLastRewardsAuction(ctx, publicPositionId)
			if !found { // maybe rewards auction not started yet
				continue
			}
			var bidAmt sdk.Int
			if auction.WinningBid == nil {
				// [1, min(coin.Amount, 10^5)]
				bidAmt = utils.RandomInt(
					r,
					utils.ZeroInt,
					utils.MinInt(coin.Amount, sdk.NewInt(100000))).
					Add(utils.OneInt)
			} else {
				winningBidAmt := auction.WinningBid.Share.Amount
				if coin.Amount.LTE(winningBidAmt) {
					// cannot win the auction
					continue
				}
				// [winningBidAmt+1, min(coin.Amount, winningBidAmt)]
				bidAmt = utils.RandomInt(
					r,
					winningBidAmt,
					utils.MinInt(
						coin.Amount,
						winningBidAmt.ToDec().Mul(utils.ParseDec("1.05")).Ceil().TruncateInt())).
					Add(utils.OneInt)
			}
			share := sdk.NewCoin(coin.Denom, bidAmt)
			msg = types.NewMsgPlaceBid(acc.Address, publicPositionId, auction.Id, share)
			return acc, msg, true
		}
	}
	return acc, nil, false
}
