package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestFarming() {
	_, pool := s.CreateSampleMarketAndPool()
	lpAddr1 := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	lpAddr2 := s.FundedAccount(2, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	position1, liquidity1, _ := s.AddLiquidity(
		lpAddr1, lpAddr1, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	position2, liquidity2, _ := s.AddLiquidity(
		lpAddr2, lpAddr2, pool.Id, utils.ParseDec("4.8"), utils.ParseDec("5.2"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	fmt.Println(liquidity1)
	fmt.Println(liquidity2)

	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1uatom")) // make initial supply
	s.CreatePrivateFarmingPlan(
		utils.TestAddress(0), "", utils.TestAddress(0), []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("1000000uatom")),
		},
		utils.ParseTime("0001-01-01T00:00:00Z"), utils.ParseTime("9999-12-31T23:59:59Z"),
		utils.ParseCoins("10000_000000uatom"), true)

	s.NextBlock()

	s.Collect(lpAddr1, lpAddr1, position1.Id, utils.ParseCoins("9uatom"))
	s.Collect(lpAddr2, lpAddr2, position2.Id, utils.ParseCoins("47uatom"))

	ordererAddr := s.FundedAccount(3, utils.ParseCoins("10000_000000uusd"))
	s.PlaceMarketOrder(pool.MarketId, ordererAddr, true, sdk.NewInt(120_000000))

	poolState := s.App.AMMKeeper.MustGetPoolState(s.Ctx, pool.Id)
	fmt.Println(poolState.CurrentPrice)

	s.NextBlock()

	s.Collect(lpAddr1, lpAddr1, position1.Id, utils.ParseCoins("56uatom"))
	s.Collect(lpAddr2, lpAddr2, position2.Id, utils.ParseCoins(""))
}
