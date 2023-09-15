package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	exchangekeeper "github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestSwapExactAmountIn() {
	creatorAddr := utils.TestAddress(0)
	s.FundAccount(creatorAddr, utils.ParseCoins("10000_000000ucre,10000_000000uatom,10000_000000uusd"))

	market1 := s.CreateMarket("ucre", "uusd")
	market2 := s.CreateMarket("uatom", "ucre")
	s.MakeLastPrice(market1.Id, creatorAddr, utils.ParseDec("5"))
	s.MakeLastPrice(market2.Id, creatorAddr, utils.ParseDec("2"))

	pool1 := s.CreatePool(market1.Id, utils.ParseDec("5"))
	s.AddLiquidity(
		creatorAddr, pool1.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseCoins("1000_000000ucre,1000_000000uusd"))
	pool2 := s.CreatePool(market2.Id, utils.ParseDec("2"))
	s.AddLiquidity(
		creatorAddr, pool2.Id, utils.ParseDec("1.5"), utils.ParseDec("3"),
		utils.ParseCoins("1000_000000uatom,1000_000000ucre"))

	ordererAddr := utils.TestAddress(1)
	s.FundAccount(ordererAddr, utils.ParseCoins("10000_000000ucre,10000_000000uatom,10000_000000uusd"))

	routes := []uint64{market1.Id, market2.Id}
	input := sdk.NewInt64DecCoin("uusd", 100_000000)
	minOutput := sdk.NewInt64DecCoin("uatom", 9_000000)
	output, _ := s.SwapExactAmountIn(ordererAddr, routes, input, minOutput, false)
	s.AssertEqual(utils.ParseDecCoin("9874881uatom"), output)
}

func (s *KeeperTestSuite) TestQueryBestSwapExactAmountInRoutes() {
	market1, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("9.7"))
	market2, pool2 := s.CreateMarketAndPool("uatom", "ucre", utils.ParseDec("1.04"))
	market3, pool3 := s.CreateMarketAndPool("uatom", "uusd", utils.ParseDec("10.3"))
	creatorAddr := s.FundedAccount(0, enoughCoins)
	s.MakeLastPrice(market1.Id, creatorAddr, utils.ParseDec("9.7"))
	s.MakeLastPrice(market2.Id, creatorAddr, utils.ParseDec("1.04"))
	s.MakeLastPrice(market3.Id, creatorAddr, utils.ParseDec("10.3"))

	s.AddLiquidity(creatorAddr, pool1.Id, utils.ParseDec("9.5"), utils.ParseDec("10"),
		utils.ParseCoins("1000_000000ucre,10000_000000uusd"))
	s.AddLiquidity(creatorAddr, pool2.Id, utils.ParseDec("1"), utils.ParseDec("1.2"),
		utils.ParseCoins("1000_000000uatom,1000_000000ucre"))
	s.AddLiquidity(creatorAddr, pool3.Id, utils.ParseDec("9.7"), utils.ParseDec("11"),
		utils.ParseCoins("1000_000000uatom,10000_000000uusd"))

	querier := exchangekeeper.Querier{Keeper: s.App.ExchangeKeeper}
	resp, err := querier.BestSwapExactAmountInRoutes(sdk.WrapSDKContext(s.Ctx), &types.QueryBestSwapExactAmountInRoutesRequest{
		Input:       "100000000ucre",
		OutputDenom: "uusd",
	})
	s.Require().NoError(err)

	s.Require().EqualValues([]uint64{2, 3}, resp.Routes)
	s.AssertEqual(utils.ParseDecCoin("972699556uusd"), resp.Output)
	s.Require().Len(resp.Results, 2)
	s.Require().EqualValues(2, resp.Results[0].MarketId)
	s.AssertEqual(utils.ParseDecCoin("100000000ucre"), resp.Results[0].Input)
	s.AssertEqual(utils.ParseDecCoin("95135827uatom"), resp.Results[0].Output)
	s.AssertEqual(utils.ParseDecCoin("142918.117689604923534808uatom"), resp.Results[0].Fee)
	s.Require().EqualValues(3, resp.Results[1].MarketId)
	s.AssertEqual(utils.ParseDecCoin("95135827uatom"), resp.Results[1].Input)
	s.AssertEqual(utils.ParseDecCoin("972699556uusd"), resp.Results[1].Output)
	s.AssertEqual(utils.ParseDecCoin("1461241.196265277978706624uusd"), resp.Results[1].Fee)
}
