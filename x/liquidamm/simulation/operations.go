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
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgMintShare = "op_weight_msg_mint_share"
	OpWeightMsgBurnShare = "op_weight_msg_burn_share"
	OpWeightMsgPlaceBid  = "op_weight_msg_place_bid"

	DefaultWeightMsgMintShare int = 50
	DefaultWeightMsgBurnShare int = 30
	DefaultWeightMsgPlaceBid  int = 20
)

var (
	gas  = uint64(20000000)
	fees = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak types.AccountKeeper, bk types.BankKeeper, ammK types.AMMKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgMintShare int
	appParams.GetOrGenerate(cdc, OpWeightMsgMintShare, &weightMsgMintShare, nil,
		func(_ *rand.Rand) {
			weightMsgMintShare = DefaultWeightMsgMintShare
		},
	)

	var weightMsgBurnShare int
	appParams.GetOrGenerate(cdc, OpWeightMsgBurnShare, &weightMsgBurnShare, nil,
		func(_ *rand.Rand) {
			weightMsgBurnShare = DefaultWeightMsgBurnShare
		},
	)

	var weightMsgPlaceBid int
	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceBid, &weightMsgPlaceBid, nil,
		func(_ *rand.Rand) {
			weightMsgPlaceBid = DefaultWeightMsgPlaceBid
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgMintShare,
			SimulateMsgMintShare(ak, bk, ammK, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBurnShare,
			SimulateMsgBurnShare(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgPlaceBid,
			SimulateMsgPlaceBid(ak, bk, k),
		),
	}
}

// SimulateMsgMintShare generates a MsgMintShare with random values
func SimulateMsgMintShare(ak types.AccountKeeper, bk types.BankKeeper, ammK types.AMMKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgMintShareParams(r, accs, bk, ammK, k, ctx)
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
func SimulateMsgBurnShare(ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, msg, found := findMsgBurnShareParams(r, accs, bk, ctx)
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
// nolint: interfacer
func SimulateMsgPlaceBid(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// TODO: not implemented yet

		return simtypes.OperationMsg{}, nil, nil
		// return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

func findMsgMintShareParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, ammK types.AMMKeeper, k keeper.Keeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgMintShare, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	var publicPositions []types.PublicPosition
	k.IterateAllPublicPositions(ctx, func(publicPosition types.PublicPosition) (stop bool) {
		publicPositions = append(publicPositions, publicPosition)
		return false
	})
	if len(publicPositions) == 0 {
		return acc, nil, false
	}
	utils.Shuffle(r, publicPositions)
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		for _, publicPosition := range publicPositions {
			pool, found := ammK.GetPool(ctx, publicPosition.PoolId)
			if !found { // sanity check
				panic("pool not found")
			}
			cacheCtx, _ := ctx.CacheContext()
			lowerPrice := exchangetypes.PriceAtTick(publicPosition.LowerTick)
			upperPrice := exchangetypes.PriceAtTick(publicPosition.UpperTick)
			desiredAmt := simtypes.RandSubsetCoins(
				r, sdk.NewCoins(
					sdk.NewCoin(pool.Denom0, spendable.AmountOf(pool.Denom0)),
					sdk.NewCoin(pool.Denom1, spendable.AmountOf(pool.Denom1))))
			_, _, _, err := ammK.AddLiquidity(
				cacheCtx, k.GetModuleAddress(), acc.Address, publicPosition.PoolId, lowerPrice, upperPrice,
				desiredAmt)
			if err != nil {
				continue
			}
			msg = types.NewMsgMintShare(acc.Address, publicPosition.Id, desiredAmt)
			return acc, msg, true
		}
	}
	return acc, msg, false
}

func findMsgBurnShareParams(
	r *rand.Rand, accs []simtypes.Account,
	bk types.BankKeeper, ctx sdk.Context) (acc simtypes.Account, msg *types.MsgBurnShare, found bool) {
	accs = utils.ShuffleSimAccounts(r, accs)
	for _, acc = range accs {
		spendable := bk.SpendableCoins(ctx, acc.Address)
		var shares []sdk.Coin
		for _, coin := range spendable {
			if _, err := types.ParseShareDenom(coin.Denom); err == nil {
				shares = append(shares, coin)
			}
		}
		if len(shares) > 0 {
			share := shares[r.Intn(len(shares))]
			publicPositionId, _ := types.ParseShareDenom(share.Denom)
			share.Amount = utils.SimRandomInt(r, sdk.NewInt(1), share.Amount)
			msg = types.NewMsgBurnShare(acc.Address, publicPositionId, share)
			return acc, msg, true
		}
	}
	return acc, nil, false
}
