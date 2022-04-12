package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/app"
	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/claim/simulation"
	"github.com/crescent-network/crescent/x/claim/types"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"
)

func TestSimulateMsgClaim(t *testing.T) {
	app := chain.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	r := rand.New(rand.NewSource(0))
	accs := getTestingAccounts(t, r, app, ctx, 1)

	srcAddr := utils.TestAddress(0)
	airdrop := types.Airdrop{
		Id:            1,
		SourceAddress: srcAddr.String(),
		Conditions: []types.ConditionType{
			types.ConditionTypeDeposit,
			types.ConditionTypeSwap,
			types.ConditionTypeLiquidStake,
			types.ConditionTypeVote,
		},
		StartTime: utils.ParseTime("2022-01-01T00:00:00Z"),
		EndTime:   utils.ParseTime("2023-01-01T00:00:00Z"),
	}
	app.ClaimKeeper.SetAirdrop(ctx, airdrop)
	err := chain.FundAccount(app.BankKeeper, ctx, srcAddr, utils.ParseCoins("1000000stake"))
	require.NoError(t, err)
	claimRecord := types.ClaimRecord{
		AirdropId:             airdrop.Id,
		Recipient:             accs[0].Address.String(),
		InitialClaimableCoins: utils.ParseCoins("1000000stake"),
		ClaimableCoins:        utils.ParseCoins("1000000stake"),
		ClaimedConditions:     nil,
	}
	app.ClaimKeeper.SetClaimRecord(ctx, claimRecord)

	pair := liquiditytypes.NewPair(1, "stake", "denom1")
	app.LiquidityKeeper.SetPair(ctx, pair)
	_, err = app.LiquidityKeeper.LimitOrder(ctx, liquiditytypes.NewMsgLimitOrder(
		accs[0].Address, pair.Id, liquiditytypes.OrderDirectionSell,
		utils.ParseCoin("10000stake"), "denom1", utils.ParseDec("1.0"), sdk.NewInt(10000), 0))
	require.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	op := simulation.SimulateMsgClaim(app.AccountKeeper, app.BankKeeper, app.LiquidStakingKeeper, app.ClaimKeeper)
	opMsg, futureOps, err := op(r, app.BaseApp, ctx, accs, "")
	require.NoError(t, err)
	require.True(t, opMsg.OK)
	require.Len(t, futureOps, 0)

	var msg types.MsgClaim
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	require.Equal(t, types.TypeMsgClaim, msg.Type())
	require.Equal(t, types.ModuleName, msg.Route())
	require.Equal(t, "cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Recipient)
	require.Equal(t, uint64(1), msg.AirdropId)
	require.Equal(t, types.ConditionTypeSwap, msg.ConditionType)
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *chain.App, ctx sdk.Context, n int) []simtypes.Account {
	accs := simtypes.RandomAccounts(r, n)

	initAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accs {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		require.NoError(t, chain.FundAccount(app.BankKeeper, ctx, account.Address, initCoins))
	}

	return accs
}
