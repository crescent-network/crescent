package farming_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingapp "github.com/tendermint/farming/app"
	"github.com/tendermint/farming/x/farming"
	"github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// createTestInput returns a simapp with custom FarmingKeeper
// to avoid messing with the hooks.
func createTestInput() (*farmingapp.FarmingApp, sdk.Context, []sdk.AccAddress) {
	app := farmingapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.FarmingKeeper = keeper.NewKeeper(
		app.AppCodec(),
		app.GetKey(types.StoreKey),
		app.GetSubspace(types.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		map[string]bool{},
	)

	addrs := farmingapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(200_000_000))

	return app, ctx, addrs
}

func TestMsgCreateFixedAmountPlan(t *testing.T) {
	app, ctx, addrs := createTestInput()

	name := "test"
	creator := addrs[0]
	stakingCoinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)
	startTime := time.Now().UTC()
	endTime := startTime.AddDate(1, 0, 0)
	epochAmount := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1)))

	msg := types.NewMsgCreateFixedAmountPlan(
		name,
		creator,
		stakingCoinWeights,
		startTime,
		endTime,
		epochAmount,
	)

	handler := farming.NewHandler(app.FarmingKeeper)
	_, err := handler(ctx, msg)
	require.NoError(t, err)

	plans := app.FarmingKeeper.GetAllPlans(ctx)
	require.Equal(t, 1, len(plans))
	require.Equal(t, creator.String(), plans[0].GetTerminationAddress().String())
	require.Equal(t, types.PrivatePlanFarmingPoolAddress(name, 1), plans[0].GetFarmingPoolAddress())
}

func TestMsgCreateRatioPlan(t *testing.T) {
	app, ctx, addrs := createTestInput()

	name := "test"
	creator := addrs[0]
	stakingCoinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)
	startTime := time.Now().UTC()
	endTime := startTime.AddDate(1, 0, 0)
	epochAmount := sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1)))

	msg := types.NewMsgCreateFixedAmountPlan(
		name,
		creator,
		stakingCoinWeights,
		startTime,
		endTime,
		epochAmount,
	)

	handler := farming.NewHandler(app.FarmingKeeper)
	_, err := handler(ctx, msg)
	require.NoError(t, err)

	plans := app.FarmingKeeper.GetAllPlans(ctx)
	require.Equal(t, 1, len(plans))
	require.Equal(t, creator.String(), plans[0].GetTerminationAddress().String())
	require.Equal(t, types.PrivatePlanFarmingPoolAddress(name, 1), plans[0].GetFarmingPoolAddress())
}

func TestMsgStake(t *testing.T) {
	// TODO: not implemented yet
}

func TestMsgUnstake(t *testing.T) {
	// TODO: not implemented yet
}

func TestMsgHarvest(t *testing.T) {
	// TODO: not implemented yet
}
