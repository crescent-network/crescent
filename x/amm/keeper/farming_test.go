package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestFarming() {
	pool := s.CreateSamplePool()
	lpAddr1 := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	lpAddr2 := s.FundedAccount(2, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	position1, liquidity1, _, _ := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		sdk.NewInt(100_000000), sdk.NewInt(500_000000), sdk.ZeroInt(), sdk.ZeroInt())
	position2, liquidity2, _, _ := s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("4.8"), utils.ParseDec("5.2"),
		sdk.NewInt(100_000000), sdk.NewInt(500_000000), sdk.ZeroInt(), sdk.ZeroInt())
	fmt.Println(liquidity1)
	fmt.Println(liquidity2)

	s.CreatePrivateFarmingPlan(
		utils.TestAddress(0), []types.RewardAllocation{
			types.NewRewardAllocation(pool.Id, utils.ParseCoins("1000000uatom")),
		},
		utils.ParseTime("0001-01-01T00:00:00Z"), utils.ParseTime("9999-12-31T23:59:59Z"),
		utils.ParseCoins("10000_000000uatom"), true)

	s.NextBlock()

	fmt.Println(s.App.AMMKeeper.Harvest(s.Ctx, lpAddr1, position1.Id))
	fmt.Println(s.App.AMMKeeper.Harvest(s.Ctx, lpAddr2, position2.Id))

	ordererAddr := s.FundedAccount(3, utils.ParseCoins("10000_000000uusd"))
	s.PlaceMarketOrder(pool.MarketId, ordererAddr, true, sdk.NewInt(120_000000))

	poolState := s.App.AMMKeeper.MustGetPoolState(s.Ctx, pool.Id)
	fmt.Println(poolState.CurrentPrice)

	s.NextBlock()

	fmt.Println(s.App.AMMKeeper.Harvest(s.Ctx, lpAddr1, position1.Id))
	fmt.Println(s.App.AMMKeeper.Harvest(s.Ctx, lpAddr2, position2.Id))
}
