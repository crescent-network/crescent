package keeper_test

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	_ "github.com/stretchr/testify/suite"

	chain "github.com/crescent-network/crescent/v4/app"
)

func (suite *KeeperTestSuite) TestVesting() {
	app := suite.app
	ctx := suite.ctx
	k := suite.keeper

	t, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	ctx = ctx.WithBlockTime(t)

	_, pub1, addr1 := testdata.KeyTestPubAddr()
	//_, pub2, addr2 := testdata.KeyTestPubAddr()

	err := chain.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, initialBalances)
	suite.Require().NoError(err)

	acc := app.AccountKeeper.GetAccount(ctx, addr1)
	acc.SetPubKey(pub1)
	acc.SetSequence(10)
	app.AccountKeeper.SetAccount(ctx, acc)

	fmt.Println(acc)
	fmt.Println(app.BankKeeper.SpendableCoins(ctx, acc.GetAddress()))

	originalVesting := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(333333)))
	startTime := ctx.BlockTime().Unix()
	periods := vestingtypes.Periods{
		vestingtypes.Period{Length: int64(23 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 111111)}},
		vestingtypes.Period{Length: int64(23 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 111111)}},
		vestingtypes.Period{Length: int64(23 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 111111)}},
	}

	k.Vesting(ctx, acc.GetAddress(), originalVesting, startTime, periods)

	bacc := app.AccountKeeper.GetAccount(ctx, acc.GetAddress())
	_, ok := bacc.(exported.VestingAccount)
	suite.Require().True(ok)

	fmt.Println(app.BankKeeper.SpendableCoins(ctx, acc.GetAddress()))

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(24 * time.Hour))

	fmt.Println(app.BankKeeper.SpendableCoins(ctx, acc.GetAddress()), startTime, ctx.BlockTime().Unix())

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(24 * time.Hour))

	fmt.Println(app.BankKeeper.SpendableCoins(ctx, acc.GetAddress()), startTime, ctx.BlockTime().Unix())

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(24 * time.Hour))

	fmt.Println(app.BankKeeper.SpendableCoins(ctx, acc.GetAddress()), startTime, ctx.BlockTime().Unix())
	//suite.EqualValues(balanceAfter, balanceAfter2)

}
