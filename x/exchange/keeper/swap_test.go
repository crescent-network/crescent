package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (s *KeeperTestSuite) TestSwapExactIn() {
	creatorAddr := utils.TestAddress(0)
	s.FundAccount(creatorAddr, utils.ParseCoins("10000_000000ucre,10000_000000uatom,10000_000000uusd"))

	market1 := s.CreateSpotMarket(creatorAddr, "ucre", "uusd", true)
	market2 := s.CreateSpotMarket(creatorAddr, "uatom", "ucre", true)

	pool1 := s.CreatePool(creatorAddr, market1.Id, 100, utils.ParseDec("5"), true)
	s.AddLiquidity(
		creatorAddr, pool1.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		sdk.NewInt(1000_000000), sdk.NewInt(1000_000000), sdk.NewInt(1), sdk.NewInt(1))
	pool2 := s.CreatePool(creatorAddr, market2.Id, 100, utils.ParseDec("2"), true)
	s.AddLiquidity(
		creatorAddr, pool2.Id, utils.ParseDec("1.5"), utils.ParseDec("3"),
		sdk.NewInt(1000_000000), sdk.NewInt(1000_000000), sdk.NewInt(1), sdk.NewInt(1))

	ordererAddr := utils.TestAddress(1)
	s.FundAccount(ordererAddr, utils.ParseCoins("10000_000000ucre,10000_000000uatom,10000_000000uusd"))

	routes := []uint64{market1.Id, market2.Id}
	input := sdk.NewInt64Coin("uusd", 100_000000)
	minOutput := sdk.NewInt64Coin("uatom", 9_000000)
	output := s.SwapExactIn(ordererAddr, routes, input, minOutput)
	s.Require().Equal("9886485uatom", output.String())
}
