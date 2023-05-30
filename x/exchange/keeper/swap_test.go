package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (s *KeeperTestSuite) TestSwapExactAmountIn() {
	creatorAddr := utils.TestAddress(0)
	s.FundAccount(creatorAddr, utils.ParseCoins("10000_000000ucre,10000_000000uatom,10000_000000uusd"))

	market1 := s.CreateMarket(creatorAddr, "ucre", "uusd", true)
	market2 := s.CreateMarket(creatorAddr, "uatom", "ucre", true)

	pool1 := s.CreatePool(creatorAddr, market1.Id, utils.ParseDec("5"), true)
	s.AddLiquidity(
		creatorAddr, creatorAddr, pool1.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseCoins("1000_000000ucre,1000_000000uusd"))
	pool2 := s.CreatePool(creatorAddr, market2.Id, utils.ParseDec("2"), true)
	s.AddLiquidity(
		creatorAddr, creatorAddr, pool2.Id, utils.ParseDec("1.5"), utils.ParseDec("3"),
		utils.ParseCoins("1000_000000uatom,1000_000000ucre"))

	ordererAddr := utils.TestAddress(1)
	s.FundAccount(ordererAddr, utils.ParseCoins("10000_000000ucre,10000_000000uatom,10000_000000uusd"))

	routes := []uint64{market1.Id, market2.Id}
	input := sdk.NewInt64Coin("uusd", 100_000000)
	minOutput := sdk.NewInt64Coin("uatom", 9_000000)
	output := s.SwapExactAmountIn(ordererAddr, routes, input, minOutput, false)
	s.Require().Equal("9845282uatom", output.String())
}