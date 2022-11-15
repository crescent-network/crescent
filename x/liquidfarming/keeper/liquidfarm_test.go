package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	utils "github.com/crescent-network/crescent/v3/types"

	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	lpfarmtypes "github.com/crescent-network/crescent/v3/x/lpfarm/types"
)

func (s *KeeperTestSuite) TestLiquidFarm_Validation() {
	err := s.keeper.LiquidFarm(s.ctx, 1, s.addr(0), utils.ParseCoin("100_000_000pool1"))
	s.Require().EqualError(err, "pool 1 not found: not found")

	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	s.createLiquidFarm(pool.Id, sdk.NewInt(100_000_000), sdk.NewInt(100_000_000), sdk.ZeroDec())

	for _, tc := range []struct {
		name        string
		msg         *types.MsgLiquidFarm
		postRun     func(ctx sdk.Context, farmerAddr sdk.AccAddress)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgLiquidFarm(
				pool.Id,
				helperAddr.String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 1_000_000_000),
			),
			func(ctx sdk.Context, farmerAddr sdk.AccAddress) {
				reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
				position, found := s.app.LPFarmKeeper.GetPosition(ctx, reserveAddr, pool.PoolCoinDenom)
				s.Require().True(found)
				s.Require().Equal(sdk.NewInt(1_000_000_000), position.FarmingAmount)
			},
			"",
		},
		{
			"minimum farm amount",
			types.NewMsgLiquidFarm(
				pool.Id,
				s.addr(0).String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 100),
			),
			nil,
			"100 is smaller than 100000000: smaller than the minimum amount",
		},
		{
			"insufficient funds",
			types.NewMsgLiquidFarm(
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
			err := s.keeper.LiquidFarm(cacheCtx, tc.msg.PoolId, tc.msg.GetFarmer(), tc.msg.FarmingCoin)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(cacheCtx, tc.msg.GetFarmer())
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestLiquidFarm() {
	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	err := s.keeper.LiquidFarm(s.ctx, pool.Id, s.addr(0), utils.ParseCoin("1_000_000pool1"))
	s.Require().EqualError(err, "liquid farm by pool 1 not found: not found")

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	var (
		amount1 = sdk.NewInt(100_000_000)
		amount2 = sdk.NewInt(200_000_000)
		amount3 = sdk.NewInt(300_000_000)
	)

	s.liquidFarm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, amount2), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, amount3), true)
	s.nextBlock()

	// Check if the reserve account farmed the coin in the farm module
	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(amount1.Add(amount2).Add(amount3), position.FarmingAmount)
}

func (s *KeeperTestSuite) TestLiquidFarm_WithFarmPlan() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(helperAddr, []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("500_000_000stake"))

	err := s.keeper.LiquidFarm(s.ctx, pool.Id, s.addr(0), utils.ParseCoin("1_000_000pool1"))
	s.Require().EqualError(err, "liquid farm by pool 1 not found: not found")

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	s.liquidFarm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(100_000_000)), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(200_000_000)), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(300_000_000)), true)
	s.nextBlock()

	// Check if the reserve account farmed the coin in the farm module
	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().True(position.FarmingAmount.Equal(sdk.NewInt(600_000_000)))

	// Auto withdrawn rewards are transferred to the WithdrawnRewardsReserveAddress
	rewardsReserveAddr := types.WithdrawnRewardsReserveAddress(pool.Id)
	s.Require().True(!s.getBalances(rewardsReserveAddr).IsZero())
}

func (s *KeeperTestSuite) TestLiquidUnfarm_Validation() {
	_, _, err := s.keeper.LiquidUnfarm(s.ctx, 1, s.addr(0), utils.ParseCoin("100_000_000pool1"))
	s.Require().EqualError(err, "pool 1 not found: not found")

	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.liquidFarm(pool.Id, s.addr(0), sdk.NewInt64Coin(pool.PoolCoinDenom, 1_000_000_000), true)

	for _, tc := range []struct {
		name        string
		msg         *types.MsgLiquidUnfarm
		postRun     func(ctx sdk.Context, unfarmedCoin sdk.Coin)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgLiquidUnfarm(
				pool.Id,
				s.addr(0).String(),
				sdk.NewInt64Coin(types.LiquidFarmCoinDenom(pool.Id), 1_000_000_000),
			),
			func(ctx sdk.Context, unfarmedCoin sdk.Coin) {
				s.Require().True(unfarmedCoin.Amount.Equal(sdk.NewInt(1_000_000_000)))
			},
			"",
		},
		{
			"insufficient balance",
			types.NewMsgLiquidUnfarm(
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
			unfarmInfo, _, err := s.keeper.LiquidUnfarm(cacheCtx, tc.msg.PoolId, tc.msg.GetFarmer(), tc.msg.UnfarmingCoin)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(cacheCtx, unfarmInfo)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestLiquidUnfarm_All() {
	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	var (
		farmerAddr = s.addr(1)
		amount1    = sdk.NewInt(100_000_000)
	)
	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)

	s.liquidFarm(pool.Id, farmerAddr, sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	// Farm amount must be 100
	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(amount1, position.FarmingAmount)

	// Ensure the amount of minted LFCoin
	balance := s.getBalance(farmerAddr, lfCoinDenom)
	s.Require().Equal(amount1, balance.Amount)

	// Unfarm all amounts
	s.liquidUnfarm(pool.Id, farmerAddr, balance, false)

	// Ensure that the position is deleted from the store
	position, found = s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().False(found)

	// Ensure the total supply
	supply := s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom)
	s.Require().True(supply.IsZero())
}

func (s *KeeperTestSuite) TestLiquidUnfarm_Partial() {
	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)

	var (
		farmerAddr1 = s.addr(1)
		farmerAddr2 = s.addr(2)

		amount1 = sdk.NewInt(5_000_000_000)
		amount2 = sdk.NewInt(1_000_000_000)
	)

	s.liquidFarm(pool.Id, farmerAddr1, sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, farmerAddr2, sdk.NewCoin(pool.PoolCoinDenom, amount2), true)
	s.nextBlock()

	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(amount1.Add(amount2), position.FarmingAmount)

	// Ensure the amount of minted LFCoin
	s.Require().Equal(amount1, s.getBalance(farmerAddr1, lfCoinDenom).Amount)
	s.Require().Equal(amount2, s.getBalance(farmerAddr2, lfCoinDenom).Amount)

	// Unfarm farmer2's all LFCoin amount
	s.liquidUnfarm(pool.Id, farmerAddr2, s.getBalance(farmerAddr2, lfCoinDenom), false)

	// Ensure the total supply
	s.Require().Equal(amount1, s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom).Amount)

	var (
		farmerAddr3 = s.addr(3)
		amount3     = sdk.NewInt(1_000_000_000)
	)

	s.liquidFarm(pool.Id, farmerAddr3, sdk.NewCoin(pool.PoolCoinDenom, amount3), true)
	s.nextBlock()

	// Ensure the amount of minted LFCoin
	s.Require().Equal(amount1, s.getBalance(farmerAddr1, lfCoinDenom).Amount)
	s.Require().Equal(sdk.ZeroInt(), s.getBalance(farmerAddr2, lfCoinDenom).Amount)
	s.Require().Equal(amount3, s.getBalance(farmerAddr3, lfCoinDenom).Amount)

	// Unfarm farmer3's all LFCoin amount
	s.liquidUnfarm(pool.Id, farmerAddr3, s.getBalance(farmerAddr3, lfCoinDenom), false)

	// Ensure the farming amount
	position, found = s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(amount1, position.FarmingAmount)

	// Ensure the total supply
	s.Require().Equal(amount1, s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom).Amount)
}

func (s *KeeperTestSuite) TestLiquidUnfarm_RemoveLiquidFarm() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(helperAddr, []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("1_000_000_000stake"))

	s.nextBlock()
}

func (s *KeeperTestSuite) TestLiquidUnfarm_Complex_WithRewards() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(helperAddr, []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("500_000_000stake"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)

	var (
		farmerAddr1 = s.addr(1)
		farmerAddr2 = s.addr(2)

		amount1 = sdk.NewInt(1_000_000_000)
		amount2 = sdk.NewInt(1_000_000_000)
	)

	s.liquidFarm(pool.Id, farmerAddr1, sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, farmerAddr2, sdk.NewCoin(pool.PoolCoinDenom, amount2), true)
	s.nextBlock()

	// Ensure that the farmers received the minted LFCoin
	s.Require().Equal(amount1, s.getBalance(farmerAddr1, lfCoinDenom).Amount)
	s.Require().Equal(amount2, s.getBalance(farmerAddr2, lfCoinDenom).Amount)

	// Ensure the amount of farmed coin farmed by the reserve account
	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(amount1.Add(amount2), position.FarmingAmount)

	// Move time to auctionTime so that rewards auction is created
	s.nextAuction()

	// Ensure rewards auction is created
	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)
	auction, found := s.keeper.GetRewardsAuction(s.ctx, auctionId, pool.Id)
	s.Require().True(found)

	var (
		bidderAddr1 = s.addr(3)
		bidderAddr2 = s.addr(4)

		biddingAmt1 = sdk.NewInt(100_000_000)
		biddingAmt2 = sdk.NewInt(200_000_000)
	)

	s.placeBid(auction.PoolId, bidderAddr1, sdk.NewCoin(pool.PoolCoinDenom, biddingAmt1), true)
	s.placeBid(auction.PoolId, bidderAddr2, sdk.NewCoin(pool.PoolCoinDenom, biddingAmt2), true)
	s.nextBlock()

	// Finish the first auction and create the next rewards auction
	s.nextAuction()

	// Ensure compounding rewards are set in the store
	rewards, found := s.keeper.GetCompoundingRewards(s.ctx, pool.Id)
	s.Require().True(found)
	s.Require().Equal(biddingAmt2, rewards.Amount)

	// Ensure bidderAddr2 has received farming rewards
	s.Require().True(s.getBalance(bidderAddr2, "stake").Amount.GT(sdk.NewInt(1)))

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

	s.nextAuction()

	// Ensure compounding rewards are updated with the new bidding amount in the store
	rewards, found = s.keeper.GetCompoundingRewards(s.ctx, pool.Id)
	s.Require().True(found)
	s.Require().Equal(biddingAmt4, rewards.Amount)

	// Ensure bidderAddr24has received farming rewards
	s.Require().True(s.getBalance(bidderAddr2, "stake").Amount.GT(sdk.NewInt(1)))

	// Unfarm all
	s.liquidUnfarm(pool.Id, farmerAddr1, s.getBalance(farmerAddr1, lfCoinDenom), false)
	s.liquidUnfarm(pool.Id, farmerAddr2, s.getBalance(farmerAddr2, lfCoinDenom), false)

	// Ensure the total supply
	s.Require().Equal(sdk.ZeroInt(), s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom).Amount)
}

func (s *KeeperTestSuite) TestLiquidUnfarmAndWithdraw() {
	depositCoins := utils.ParseCoins("100_000_000denom1, 100_000_000denom2")
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, depositCoins)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	poolAmt := sdk.NewInt(1_000_000_000_000)
	s.liquidFarm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, poolAmt), false)
	s.nextBlock()

	s.nextAuction()

	// Ensure that reserve account farms the coin and its amount
	farm, found := s.app.LPFarmKeeper.GetFarm(s.ctx, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(poolAmt, farm.TotalFarmingAmount)

	// Ensure the minted LFCoin
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)
	lfCoinBalance := s.getBalance(s.addr(0), lfCoinDenom)
	s.Require().Equal(poolAmt, lfCoinBalance.Amount)

	err := s.keeper.LiquidUnfarmAndWithdraw(s.ctx, pool.Id, s.addr(0), lfCoinBalance)
	s.Require().NoError(err)

	// Call nextBlock as Withdraw is executed in batch
	s.nextBlock()

	// Ensure that depositCoins are returned to the farmer's balance
	s.Require().EqualValues(depositCoins, s.getBalances(s.addr(0)))

}

func (s *KeeperTestSuite) TestDeleteLiquidFarmInStore() {
	pair1 := s.createPair(helperAddr, "denom1", "denom2")
	pool1 := s.createPool(helperAddr, pair1.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	pair2 := s.createPair(helperAddr, "denom3", "denom4")
	pool2 := s.createPool(helperAddr, pair2.Id, utils.ParseCoins("100_000_000denom3, 100_000_000denom4"))

	liquidFarm1 := s.createLiquidFarm(pool1.Id, sdk.NewInt(100), sdk.NewInt(100), sdk.ZeroDec())
	s.nextBlock()

	s.createLiquidFarm(pool2.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.nextBlock()

	// Ensure that the param and KVStore has the same length of liquid farms
	s.Require().Len(s.keeper.GetLiquidFarmsInStore(s.ctx), 2)
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 2)

	// Remove LiquidFarm1 from the store
	s.keeper.DeleteLiquidFarm(s.ctx, liquidFarm1)

	// Ensure that the liquid farm is deleted in KVStore
	s.Require().Len(s.keeper.GetLiquidFarmsInStore(s.ctx), 1)
	s.Require().Len(s.keeper.GetLiquidFarmsInParams(s.ctx), 2)

	// Synchronize LiquidFarms in begin blocker
	s.nextBlock()

	// Ensure the length of liquid farms in the param and KVstore is the same
	s.Require().Len(s.keeper.GetLiquidFarmsInStore(s.ctx), 2)
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 2)
}

func (s *KeeperTestSuite) TestDeleteLiquidFarmInParam() {
	pair1 := s.createPair(helperAddr, "denom1", "denom2")
	pool1 := s.createPool(helperAddr, pair1.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	pair2 := s.createPair(helperAddr, "denom3", "denom4")
	pool2 := s.createPool(helperAddr, pair2.Id, utils.ParseCoins("100_000_000denom3, 100_000_000denom4"))

	liquidFarm1 := s.createLiquidFarm(pool1.Id, sdk.NewInt(100), sdk.NewInt(100), sdk.ZeroDec())
	s.nextBlock()

	s.createLiquidFarm(pool2.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.nextBlock()

	// Ensure that the param and KVStore has the same length of liquid farms
	s.Require().Len(s.keeper.GetLiquidFarmsInStore(s.ctx), 2)
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 2)

	// Remove the liquid farm object in params
	params := s.keeper.GetParams(s.ctx)
	params.LiquidFarms = []types.LiquidFarm{liquidFarm1}
	s.keeper.SetParams(s.ctx, params)

	// Ensure that the liquid farm is deleted in KVStore
	s.Require().Len(s.keeper.GetLiquidFarmsInStore(s.ctx), 2)
	s.Require().Len(s.keeper.GetLiquidFarmsInParams(s.ctx), 1)

	// Synchronize LiquidFarms in begin blocker
	s.nextBlock()

	// Ensure the length of liquid farms in the param and KVstore is the same
	s.Require().Len(s.keeper.GetLiquidFarmsInStore(s.ctx), 1)
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)
}

// [Scenario]
// There is a farming plan that allocates rewards per day
// One RewardsAuction is finished for auto compounding rewards
// While the next RewardsAuction is ongoing, LiquidFarm is removed
//
// [Expected results]
// 1. Bidders for the second RewardsAuction must be refunded
// 2. Send the accumulated farming rewards from the first to the second RewardsAuction to the fee collector
// 3. The status of RewardsAuction must be AuctionStatusFinished
func (s *KeeperTestSuite) TestDeleteLiquidFarm_EdgeCase1() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(helperAddr, []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("500_000_000stake"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	farmCoin := utils.ParseCoin("100_000_000_000pool1")
	s.liquidFarm(pool.Id, s.addr(0), farmCoin, true)
	s.nextBlock()

	// Create the first rewards auction
	s.nextAuction()

	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("200_000pool1"), true)
	s.nextBlock()

	// Ensure that the reserve account farmed the pool coin
	farm, found := s.app.LPFarmKeeper.GetPosition(s.ctx, types.LiquidFarmReserveAddress(pool.Id), pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(farmCoin.Amount, farm.FarmingAmount)

	s.nextAuction()

	// Ensure that the total farming amount is increased
	farm, found = s.app.LPFarmKeeper.GetPosition(s.ctx, types.LiquidFarmReserveAddress(pool.Id), pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().True(farm.FarmingAmount.GT(farmCoin.Amount))

	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("400_000pool1"), true)
	s.nextBlock()

	// Remove the liquid farm object in params
	params := s.keeper.GetParams(s.ctx)
	params.LiquidFarms = []types.LiquidFarm{}
	s.keeper.SetParams(s.ctx, params)

	// Synchronize and handle the removed liquid farm
	s.nextBlock()

	// Ensure that the liquid farm is removed and synchronized
	s.Require().Len(s.keeper.GetLiquidFarmsInStore(s.ctx), 0)
	s.Require().Len(s.keeper.GetLiquidFarmsInParams(s.ctx), 0)

	// Ensure the auction status
	auction, found := s.keeper.GetRewardsAuction(s.ctx, 1, pool.Id)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusFinished, auction.Status)

	// Ensure that the fee collector is not empty
	feeCollectorAddr, _ := sdk.AccAddressFromBech32(s.keeper.GetFeeCollector(s.ctx))
	s.Require().True(!s.getBalances(feeCollectorAddr).IsZero())

	// Unfarm all LFCoin
	s.liquidUnfarm(pool.Id, s.addr(0), s.getBalance(s.addr(0), types.LiquidFarmCoinDenom(pool.Id)), false)
	s.nextBlock()

	// Ensure that the returned pool coin is increased
	s.Require().True(s.getBalance(s.addr(0), pool.PoolCoinDenom).Amount.GT(farmCoin.Amount))

	// Ensure that the bidder got their funds back
	s.Require().Equal(utils.ParseCoin("400_000pool1"), s.getBalance(s.addr(7), pool.PoolCoinDenom))
}

func (s *KeeperTestSuite) TestMintRate() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(helperAddr, []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("500_000_000stake"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	var (
		farmerAddr1 = s.addr(1)
		amount1     = sdk.NewInt(5_000_000_000)
	)

	s.liquidFarm(pool.Id, farmerAddr1, sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	// Ensure the amount of minted LFCoin
	s.Require().Equal(amount1, s.getBalance(farmerAddr1, types.LiquidFarmCoinDenom(pool.Id)).Amount)

	s.nextAuction()

	// Ensure rewards auction is created
	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)
	auction, found := s.keeper.GetRewardsAuction(s.ctx, auctionId, pool.Id)
	s.Require().True(found)

	var (
		bidderAddr = s.addr(10)
		biddingAmt = sdk.NewInt(100_000_000)
	)

	s.placeBid(auction.PoolId, bidderAddr, sdk.NewCoin(pool.PoolCoinDenom, biddingAmt), true)
	s.nextBlock()

	s.nextAuction()

	// Ensure that the farming amount is increased due to auto compounding rewards
	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(amount1.Add(biddingAmt), position.FarmingAmount)

	// Ensure that the bidder received the farming rewards
	s.Require().True(!s.getBalance(bidderAddr, "stake").IsZero())

	var (
		farmerAddr2 = s.addr(2)
		amount2     = sdk.NewInt(5_100_000_000)
	)

	// Now, let's liquid farm again to see the changed mint rate
	s.liquidFarm(pool.Id, farmerAddr2, sdk.NewCoin(pool.PoolCoinDenom, amount2), true)
	s.nextBlock()

	// Ensure the less amount of minted LFCoin
	s.Require().True(
		s.getBalance(farmerAddr1, types.LiquidFarmCoinDenom(pool.Id)).Amount.Equal(
			s.getBalance(farmerAddr2, types.LiquidFarmCoinDenom(pool.Id)).Amount))
}

func (s *KeeperTestSuite) TestBurnRate() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(s.addr(0), []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("500_000_000stake"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.nextBlock()

	var (
		farmerAddr1 = s.addr(1)
		amount1     = sdk.NewInt(200_000)
	)

	s.liquidFarm(pool.Id, farmerAddr1, sdk.NewCoin(pool.PoolCoinDenom, amount1), true)
	s.nextBlock()

	var (
		farmerAddr2 = s.addr(2)
		amount2     = sdk.NewInt(200_000)
	)

	s.liquidFarm(pool.Id, farmerAddr2, sdk.NewCoin(pool.PoolCoinDenom, amount2), true)
	s.nextBlock()

	// Ensure that the reserve account farmed the amount
	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(amount1.Add(amount2), position.FarmingAmount)

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(10), utils.ParseCoin("100_000pool1"), true)
	s.nextBlock()

	s.placeBid(pool.Id, s.addr(11), utils.ParseCoin("200_000pool1"), true)
	s.nextBlock()

	_, found = s.keeper.GetCompoundingRewards(s.ctx, pool.Id)
	s.Require().False(found)

	s.nextAuction()

	_, found = s.keeper.GetCompoundingRewards(s.ctx, pool.Id)
	s.Require().True(found)

	s.liquidUnfarm(pool.Id, farmerAddr1, utils.ParseCoin("200_000lf1"), false)
	s.nextBlock()

	s.Require().Equal(amount1, s.getBalance(farmerAddr1, pool.PoolCoinDenom).Amount)

	s.liquidUnfarm(pool.Id, farmerAddr2, utils.ParseCoin("200_000lf1"), false)
	s.nextBlock()

	// Farmed + WinningBid (last one to unfarm)
	s.Require().Equal(amount1.Add(amount2), s.getBalance(farmerAddr2, pool.PoolCoinDenom).Amount)
}

// This case fixes invalid minting and burning amount when LiquidFarm and LiquidUnfarm used to use GetFarm function.
func (s *KeeperTestSuite) TestMintAndBurnRate_EdgeCase() {
	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("100000000pool1"), true)
	s.nextBlock()

	// Farm directly in the lpfarm module
	_, err := s.app.LPFarmKeeper.Farm(s.ctx, helperAddr, utils.ParseCoin("200000000pool1"))
	s.Require().NoError(err)

	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("100000000pool1"), true)
	s.nextBlock()

	// Addr1 and Addr2 must have the same amount
	addr1LFCoinBalance := s.getBalance(s.addr(1), types.LiquidFarmCoinDenom(pool.Id))
	addr2LFCoinBalance := s.getBalance(s.addr(2), types.LiquidFarmCoinDenom(pool.Id))
	s.Require().True(addr1LFCoinBalance.IsEqual(addr2LFCoinBalance))

	// Addr2 must receive the same amount
	s.liquidUnfarm(pool.Id, s.addr(2), utils.ParseCoin("100000000lf1"), false)

	addr2PoolCoinBalance := s.getBalance(s.addr(2), pool.PoolCoinDenom)
	s.Require().True(addr2LFCoinBalance.Amount.Equal(addr2PoolCoinBalance.Amount))
}
