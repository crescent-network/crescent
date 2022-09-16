package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	utils "github.com/crescent-network/crescent/v2/types"

	"github.com/crescent-network/crescent/v2/x/liquidfarming"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
)

func (s *KeeperTestSuite) TestFarm_Validation() {
	err := s.keeper.Farm(s.ctx, 1, s.addr(0), utils.ParseCoin("100000000pool1"))
	s.Require().EqualError(err, "pool 1 not found: not found")

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.createLiquidFarm(pool.Id, sdk.NewInt(100_000_000), sdk.NewInt(100_000_000), sdk.ZeroDec())

	for _, tc := range []struct {
		name        string
		msg         *types.MsgFarm
		postRun     func(ctx sdk.Context, farmerAcc sdk.AccAddress)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgFarm(
				pool.Id,
				s.addr(0).String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 1_000_000_000),
			),
			func(ctx sdk.Context, farmerAcc sdk.AccAddress) {
				reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
				queuedAmt := s.app.FarmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(ctx, reserveAddr, pool.PoolCoinDenom)
				farmerBalance := s.app.BankKeeper.GetBalance(ctx, farmerAcc, types.LiquidFarmCoinDenom(pool.Id))
				s.Require().Equal(sdk.NewInt(1_000_000_000), queuedAmt)
				s.Require().Equal(sdk.NewInt(1_000_000_000), farmerBalance.Amount)
			},
			"",
		},
		{
			"minimum farm amount",
			types.NewMsgFarm(
				pool.Id,
				s.addr(0).String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 100),
			),
			nil,
			"100 is smaller than 100000000: smaller than minimum amount",
		},
		{
			"insufficient funds",
			types.NewMsgFarm(
				pool.Id,
				s.addr(5).String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 500_000_000),
			),
			nil,
			"0pool1 is smaller than 500000000pool1: insufficient funds",
		},
	} {
		s.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cacheCtx, _ := s.ctx.CacheContext()
			err := s.keeper.Farm(cacheCtx, tc.msg.PoolId, tc.msg.GetFarmer(), tc.msg.FarmingCoin)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(cacheCtx, tc.msg.GetFarmer())
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestFarm() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	err := s.keeper.Farm(s.ctx, pool.Id, s.addr(0), utils.ParseCoin("1000000pool1"))
	s.Require().EqualError(err, "liquid farm by pool 1 not found: not found")

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	var (
		amount1 = sdk.NewInt(100000000)
		amount2 = sdk.NewInt(200000000)
		amount3 = sdk.NewInt(300000000)
	)

	s.farm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	s.farm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, amount2), true)
	s.nextBlock()

	s.farm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, amount3), true)
	s.nextBlock()

	// Check if the liquid farm reserve account staked in the farming module
	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	queuedAmt := s.app.FarmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().Equal(amount1.Add(amount2).Add(amount3), queuedAmt)
}

func (s *KeeperTestSuite) TestUnfarm_Validation() {
	_, err := s.keeper.Unfarm(s.ctx, 1, s.addr(0), utils.ParseCoin("100000000pool1"))
	s.Require().EqualError(err, "pool 1 not found: not found")

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.farm(pool.Id, s.addr(0), sdk.NewInt64Coin(pool.PoolCoinDenom, 1_000_000_000), true)
	s.advanceEpochDays()

	for _, tc := range []struct {
		name        string
		msg         *types.MsgUnfarm
		postRun     func(ctx sdk.Context, unfarmInfo keeper.UnfarmInfo)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgUnfarm(
				pool.Id,
				s.addr(0).String(),
				sdk.NewInt64Coin(types.LiquidFarmCoinDenom(pool.Id), 1_000_000_000),
			),
			func(ctx sdk.Context, unfarmInfo keeper.UnfarmInfo) {
				s.Require().Equal(s.addr(0), unfarmInfo.Farmer)
			},
			"",
		},
		{
			"insufficient balance",
			types.NewMsgUnfarm(
				pool.Id,
				s.addr(5).String(),
				sdk.NewInt64Coin(types.LiquidFarmCoinDenom(pool.Id), 1_000_000_000),
			),
			nil,
			"0lf1 is smaller than 1000000000lf1: insufficient funds",
		},
	} {
		s.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cacheCtx, _ := s.ctx.CacheContext()
			unfarmInfo, err := s.keeper.Unfarm(cacheCtx, tc.msg.PoolId, tc.msg.GetFarmer(), tc.msg.BurningCoin)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(cacheCtx, unfarmInfo)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnfarm_All() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	var (
		farmerAddr  = s.addr(1)
		amount1     = sdk.NewInt(100_000_000)
		reserveAddr = types.LiquidFarmReserveAddress(pool.Id)
		lfCoinDenom = types.LiquidFarmCoinDenom(pool.Id)
	)

	s.farm(pool.Id, farmerAddr, sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	// Farm amount must be 100
	queuedAmt := s.app.FarmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().Equal(amount1, queuedAmt)

	// Ensure the amount of minted LFCoin
	balance := s.getBalance(farmerAddr, lfCoinDenom)
	s.Require().Equal(amount1, balance.Amount)

	// Unfarm all amounts
	s.unfarm(pool.Id, farmerAddr, balance, false)

	// Ensure that queued coins must be zero amount
	queuedAmt = s.app.FarmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().Equal(sdk.ZeroInt(), queuedAmt)

	// Ensure the total supply
	supply := s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom)
	s.Require().Equal(sdk.ZeroInt(), supply.Amount)
}

func (s *KeeperTestSuite) TestUnfarm_Partial() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)

	var (
		farmerAddr1 = s.addr(1)
		amount1     = sdk.NewInt(5_000_000_000)

		farmerAddr2 = s.addr(2)
		amount2     = sdk.NewInt(1_000_000_000)
	)

	s.farm(pool.Id, farmerAddr1, sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	s.farm(pool.Id, farmerAddr2, sdk.NewCoin(pool.PoolCoinDenom, amount2), true)
	s.nextBlock()

	queuedAmt := s.app.FarmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(s.ctx, reserveAddr, pool.PoolCoinDenom)
	stakedCoins := s.app.FarmingKeeper.GetAllStakedCoinsByFarmer(s.ctx, reserveAddr)
	s.Require().Equal(amount1.Add(amount2), queuedAmt)
	s.Require().Equal(sdk.ZeroInt(), stakedCoins.AmountOf(pool.PoolCoinDenom))

	// Ensure the amount of minted LFCoin
	s.Require().Equal(amount1, s.getBalance(farmerAddr1, lfCoinDenom).Amount)
	s.Require().Equal(amount2, s.getBalance(farmerAddr2, lfCoinDenom).Amount)

	// Unfarm farmer2's all LFCoin amount
	s.unfarm(pool.Id, farmerAddr2, s.getBalance(farmerAddr2, lfCoinDenom), false)

	// Ensure the total supply
	s.Require().Equal(amount1, s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom).Amount)

	var (
		farmerAddr3 = s.addr(3)
		amount3     = sdk.NewInt(1_000_000_000)
	)

	// Farm with the farmerAddr3
	s.farm(pool.Id, farmerAddr3, sdk.NewCoin(pool.PoolCoinDenom, amount3), true)
	s.nextBlock()

	// Ensure the amount of minted LFCoin
	s.Require().Equal(amount1, s.getBalance(farmerAddr1, lfCoinDenom).Amount)
	s.Require().Equal(sdk.ZeroInt(), s.getBalance(farmerAddr2, lfCoinDenom).Amount)
	s.Require().Equal(amount3, s.getBalance(farmerAddr3, lfCoinDenom).Amount)

	// Advance epoch to see if it makes any difference
	s.advanceEpochDays()

	// Unfarm farmer3's all LFCoin amount
	s.unfarm(pool.Id, farmerAddr3, s.getBalance(farmerAddr3, lfCoinDenom), false)

	// Ensure that queued and staked coins
	queuedAmt = s.app.FarmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(s.ctx, reserveAddr, pool.PoolCoinDenom)
	stakedCoins = s.app.FarmingKeeper.GetAllStakedCoinsByFarmer(s.ctx, reserveAddr)
	s.Require().Equal(sdk.ZeroInt(), queuedAmt)
	s.Require().Equal(amount1, stakedCoins.AmountOf(pool.PoolCoinDenom))

	// Ensure the total supply
	s.Require().Equal(amount1, s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom).Amount)
}

func (s *KeeperTestSuite) TestUnfarm_Complex_WithRewards() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	epochAmt := int64(100_000_000)
	s.createPrivateFixedAmountPlan(
		s.addr(0),
		map[string]string{pool.PoolCoinDenom: "1"},
		map[string]int64{"denom3": epochAmt},
		true,
	)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)

	var (
		farmerAddr1 = s.addr(1)
		amount1     = sdk.NewInt(1_000_000_000)

		farmerAddr2 = s.addr(2)
		amount2     = sdk.NewInt(1_000_000_000)
	)

	s.farm(pool.Id, farmerAddr1, sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	s.farm(pool.Id, farmerAddr2, sdk.NewCoin(pool.PoolCoinDenom, amount2), true)
	s.nextBlock()

	// Ensure the amount of minted LFCoin
	s.Require().Equal(amount1, s.getBalance(farmerAddr1, lfCoinDenom).Amount)
	s.Require().Equal(amount2, s.getBalance(farmerAddr2, lfCoinDenom).Amount)

	// Ensure queued and staked coins
	queuedCoins := s.app.FarmingKeeper.GetAllQueuedCoinsByFarmer(s.ctx, reserveAddr)
	stakedCoins := s.app.FarmingKeeper.GetAllStakedCoinsByFarmer(s.ctx, reserveAddr)
	s.Require().Equal(amount1.Add(amount2), queuedCoins.AmountOf(pool.PoolCoinDenom))
	s.Require().Equal(sdk.ZeroInt(), stakedCoins.AmountOf(pool.PoolCoinDenom))

	s.advanceEpochDays()

	// Ensure queued and staked coins
	queuedCoins = s.app.FarmingKeeper.GetAllQueuedCoinsByFarmer(s.ctx, reserveAddr)
	stakedCoins = s.app.FarmingKeeper.GetAllStakedCoinsByFarmer(s.ctx, reserveAddr)
	s.Require().Equal(sdk.ZeroInt(), queuedCoins.AmountOf(pool.PoolCoinDenom))
	s.Require().Equal(amount1.Add(amount2), stakedCoins.AmountOf(pool.PoolCoinDenom))

	// Ensure rewards auction is created
	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)
	auction, found := s.keeper.GetRewardsAuction(s.ctx, pool.Id, auctionId)
	s.Require().True(found)

	var (
		bidderAddr1 = s.addr(3)
		biddingAmt1 = sdk.NewInt(100_000_000)

		bidderAddr2 = s.addr(4)
		biddingAmt2 = sdk.NewInt(200_000_000)
	)

	s.placeBid(auction.PoolId, bidderAddr1, sdk.NewCoin(pool.PoolCoinDenom, biddingAmt1), true)
	s.placeBid(auction.PoolId, bidderAddr2, sdk.NewCoin(pool.PoolCoinDenom, biddingAmt2), true)

	s.advanceEpochDays()

	// Ensure compounding rewards are set in the store
	rewards, found := s.keeper.GetCompoundingRewards(s.ctx, pool.Id)
	s.Require().True(found)
	s.Require().Equal(biddingAmt2, rewards.Amount)

	// Ensure the farming rewards are in the balance
	s.Require().Equal(sdk.NewInt(epochAmt), s.getBalance(bidderAddr2, "denom3").Amount)

	// Ensure the next rewards auction is created
	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.ctx), 2)

	var (
		bidderAddr3 = s.addr(5)
		biddingAmt3 = sdk.NewInt(10_000_000)

		bidderAddr4 = s.addr(6)
		biddingAmt4 = sdk.NewInt(30_000_000)
	)

	s.placeBid(auction.PoolId, bidderAddr3, sdk.NewCoin(pool.PoolCoinDenom, biddingAmt3), true)
	s.placeBid(auction.PoolId, bidderAddr4, sdk.NewCoin(pool.PoolCoinDenom, biddingAmt4), true)

	s.advanceEpochDays()

	// Ensure compounding rewards are updated with the new bidding amount in the store
	rewards, found = s.keeper.GetCompoundingRewards(s.ctx, pool.Id)
	s.Require().True(found)
	s.Require().Equal(biddingAmt4, rewards.Amount)

	// Unfarm all
	s.unfarm(pool.Id, farmerAddr1, s.getBalance(farmerAddr1, lfCoinDenom), false)
	s.unfarm(pool.Id, farmerAddr2, s.getBalance(farmerAddr2, lfCoinDenom), false)

	// Ensure the balances
	// farmerAddr1: pool coin is auto compounded by 100000000 amount
	// farmerAddr2: pool coin is auto compounded by 100000000 amount + gets all rewards
	s.Require().Equal(sdk.NewInt(1_100_000_000), s.getBalance(farmerAddr1, pool.PoolCoinDenom).Amount)
	s.Require().Equal(sdk.NewInt(1_130_000_000), s.getBalance(farmerAddr2, pool.PoolCoinDenom).Amount)

	// Ensure the total supply
	s.Require().Equal(sdk.ZeroInt(), s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom).Amount)

}

func (s *KeeperTestSuite) TestUnfarmAndWithdraw() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	depositCoins := utils.ParseCoins("100_000_000denom1, 100_000_000denom2")
	pool := s.createPool(s.addr(0), pair.Id, depositCoins, true)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	poolAmt := sdk.NewInt(1_000_000_000_000)

	s.farm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, poolAmt), false)

	s.advanceEpochDays()

	// Staked amount must be 100
	totalStakings, found := s.app.FarmingKeeper.GetTotalStakings(s.ctx, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(poolAmt, totalStakings.Amount)

	// Check minted LFCoin
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)
	lfCoinBalance := s.getBalance(s.addr(0), lfCoinDenom)
	s.Require().Equal(sdk.NewCoin(lfCoinDenom, poolAmt), lfCoinBalance)

	// Call UnfarmAndWithdraw
	err := s.keeper.UnfarmAndWithdraw(s.ctx, pool.Id, s.addr(0), lfCoinBalance)
	s.Require().NoError(err)

	// Call nextBlock as Withdraw is executed in batch
	s.nextBlock()

	// Verify
	s.Require().EqualValues(depositCoins, s.getBalances(s.addr(0)))
}

func (s *KeeperTestSuite) TestTerminateLiquidFarm() {
	pair1 := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool1 := s.createPool(s.addr(0), pair1.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	pair2 := s.createPair(s.addr(0), "denom3", "denom4", true)
	pool2 := s.createPool(s.addr(0), pair2.Id, utils.ParseCoins("100_000_000denom3, 100_000_000denom4"), true)

	// Add liquid farms in params
	params := s.keeper.GetParams(s.ctx)
	params.LiquidFarms = append(params.LiquidFarms, types.NewLiquidFarm(pool1.Id, sdk.NewInt(100), sdk.NewInt(100), sdk.ZeroDec()))
	params.LiquidFarms = append(params.LiquidFarms, types.NewLiquidFarm(pool2.Id, sdk.NewInt(500), sdk.NewInt(500), sdk.ZeroDec()))
	s.keeper.SetParams(s.ctx, params)
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, len(params.LiquidFarms))

	// Execute BeginBlocker to store all registered LiquidFarms in params
	liquidfarming.BeginBlocker(s.ctx, s.keeper)

	// Ensure the length of liquid farms in the store and params are the same
	s.Require().Len(s.keeper.GetAllLiquidFarms(s.ctx), 2)
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 2)

	// Forcefully remove one liquid farm from the store
	liquidFarm, found := s.keeper.GetLiquidFarm(s.ctx, pool1.Id)
	s.Require().True(found)
	s.keeper.DeleteLiquidFarm(s.ctx, liquidFarm)
	s.Require().Len(s.keeper.GetAllLiquidFarms(s.ctx), 1)
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 2)

	// Execute BeginBlocker again to store all registered LiquidFarms in params
	liquidfarming.BeginBlocker(s.ctx, s.keeper)

	// Ensure the length of liquid farms in the store and params are the same
	s.Require().Len(s.keeper.GetAllLiquidFarms(s.ctx), 2)
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 2)

	// Now, in order to test adding new liquid farm in the params
	// Farm some coins
	s.farm(pool1.Id, s.addr(1), sdk.NewCoin(pool1.PoolCoinDenom, sdk.NewInt(200_000)), true)
	s.farm(pool1.Id, s.addr(2), sdk.NewCoin(pool1.PoolCoinDenom, sdk.NewInt(200_000)), true)
	s.advanceEpochDays() // trigger AfterStaked hook to mint LFCoin and delete QueuedFarming
	s.advanceEpochDays() // trigger AllocateRewards hook to create rewards auctions

	s.placeBid(pool1.Id, s.addr(3), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool1.Id, s.addr(4), utils.ParseCoin("200_000pool1"), true)
	s.placeBid(pool1.Id, s.addr(5), utils.ParseCoin("500_000pool1"), true)

	// Remove the first liquid farm object in params
	params = s.keeper.GetParams(s.ctx)
	params.LiquidFarms = []types.LiquidFarm{types.NewLiquidFarm(pool2.Id, sdk.NewInt(100), sdk.NewInt(100), sdk.ZeroDec())}
	s.keeper.SetParams(s.ctx, params)

	// Execute BeginBlocker again to store all registered LiquidFarms in params
	liquidfarming.BeginBlocker(s.ctx, s.keeper)

	// Ensure that bidders got their funds back
	s.Require().Equal(utils.ParseCoin("100_000pool1"), s.getBalance(s.addr(3), pool1.PoolCoinDenom))
	s.Require().Equal(utils.ParseCoin("200_000pool1"), s.getBalance(s.addr(4), pool1.PoolCoinDenom))
	s.Require().Equal(utils.ParseCoin("500_000pool1"), s.getBalance(s.addr(5), pool1.PoolCoinDenom))
	s.Require().Equal(utils.ParseCoin("400_000pool1"), s.getBalance(types.LiquidFarmReserveAddress(1), pool1.PoolCoinDenom))

	// Ensure the auction status
	auction, found := s.keeper.GetRewardsAuction(s.ctx, pool1.Id, 1)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusFinished, auction.Status)

	// Ensure all staked coins is zero
	s.Require().True(s.app.FarmingKeeper.GetAllStakedCoinsByFarmer(s.ctx, types.LiquidFarmReserveAddress(1)).IsZero())
}

func (s *KeeperTestSuite) TestMintAndBurnRate() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	epochAmt := int64(100_000_000)
	s.createPrivateFixedAmountPlan(
		s.addr(0),
		map[string]string{pool.PoolCoinDenom: "1"},
		map[string]int64{"denom3": epochAmt},
		true,
	)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)

	var (
		farmerAddr1 = s.addr(1)
		amount1     = sdk.NewInt(5_000_000_000)
	)

	s.farm(pool.Id, farmerAddr1, sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	// Ensure the amount of minted LFCoin
	s.Require().Equal(amount1, s.getBalance(farmerAddr1, lfCoinDenom).Amount)

	s.advanceEpochDays()

	// Ensure rewards auction is created
	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)
	auction, found := s.keeper.GetRewardsAuction(s.ctx, pool.Id, auctionId)
	s.Require().True(found)

	var (
		bidderAddr1 = s.addr(10)
		biddingAmt1 = sdk.NewInt(100_000_000)

		bidderAddr2 = s.addr(11)
		biddingAmt2 = sdk.NewInt(200_000_000)
	)

	s.placeBid(auction.PoolId, bidderAddr1, sdk.NewCoin(pool.PoolCoinDenom, biddingAmt1), true)
	s.placeBid(auction.PoolId, bidderAddr2, sdk.NewCoin(pool.PoolCoinDenom, biddingAmt2), true)

	//
	// Farm -> Unfarm in different epoch
	//

	var (
		farmerAddr2 = s.addr(2)
		amount2     = sdk.NewInt(1_000_000_000)
	)

	s.farm(pool.Id, farmerAddr2, sdk.NewCoin(pool.PoolCoinDenom, amount2), true)
	s.nextBlock()

	// Ensure the amount of minted LFCoin
	s.Require().Equal(amount2, s.getBalance(farmerAddr2, lfCoinDenom).Amount) // 1000000000lf1

	s.advanceEpochDays()

	// Ensure compounding rewards are updated with the new bidding amount in the store
	rewards, found := s.keeper.GetCompoundingRewards(s.ctx, pool.Id)
	s.Require().True(found)
	s.Require().Equal(biddingAmt2, rewards.Amount)

	s.unfarm(pool.Id, farmerAddr2, s.getBalance(farmerAddr2, lfCoinDenom), false)
	s.nextBlock()

	// Ensure the balance
	s.Require().Equal(amount2, s.getBalance(farmerAddr2, pool.PoolCoinDenom).Amount) // 1000000000pool1

	//
	// Farm & Unfarm within the same epoch
	//

	var (
		farmerAddr3 = s.addr(3)
		amount3     = sdk.NewInt(1_000_000_000)
	)

	s.farm(pool.Id, farmerAddr3, sdk.NewCoin(pool.PoolCoinDenom, amount3), true)
	s.nextBlock()

	// Ensure the amount of minted LFCoin is less than the amount of farmed pool coin
	s.Require().True(s.getBalance(farmerAddr3, lfCoinDenom).Amount.LT(amount3)) // 961538461lf1

	lfCoin := s.getBalance(farmerAddr3, lfCoinDenom)
	s.unfarm(pool.Id, farmerAddr3, lfCoin, false)
	s.nextBlock()

	// Ensure the received pool coin is less than the amount of lfCoinBalance
	s.Require().True(s.getBalance(farmerAddr3, pool.PoolCoinDenom).Amount.LT(amount3)) // 967741935pool1

	//
	// Farm -> Rewards -> Unfarm in the different epoch
	//

	var (
		farmerAddr4 = s.addr(4)
		amount4     = sdk.NewInt(1_000_000_000)
	)

	s.farm(pool.Id, farmerAddr4, sdk.NewCoin(pool.PoolCoinDenom, amount4), true)
	s.nextBlock()

	// Ensure the amount of minted LFCoin is less than the amount of farmed pool coin
	s.Require().True(s.getBalance(farmerAddr4, lfCoinDenom).Amount.LT(amount4)) // 955610357lf1

	s.advanceEpochDays()

	lfCoin = s.getBalance(farmerAddr4, lfCoinDenom)
	s.unfarm(pool.Id, farmerAddr4, lfCoin, false)
	s.nextBlock()

	// Ensure the received pool coin is less than the amount of lfCoinBalance
	s.Require().True(s.getBalance(farmerAddr3, pool.PoolCoinDenom).Amount.LT(amount3)) // 999999999pool1
}
