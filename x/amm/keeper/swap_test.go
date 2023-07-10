package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	exchangekeeper "github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestQueryBestSwapExactAmountInRoutes() {
	_, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("9.7"))
	_, pool2 := s.CreateMarketAndPool("uatom", "ucre", utils.ParseDec("1.04"))
	_, pool3 := s.CreateMarketAndPool("uatom", "uusd", utils.ParseDec("10.3"))

	creatorAddr := s.FundedAccount(0, enoughCoins)
	s.AddLiquidity(creatorAddr, creatorAddr, pool1.Id, utils.ParseDec("9.5"), utils.ParseDec("10"),
		utils.ParseCoins("1000_000000ucre,10000_000000uusd"))
	s.AddLiquidity(creatorAddr, creatorAddr, pool2.Id, utils.ParseDec("1"), utils.ParseDec("1.2"),
		utils.ParseCoins("1000_000000uatom,1000_000000ucre"))
	s.AddLiquidity(creatorAddr, creatorAddr, pool3.Id, utils.ParseDec("9.7"), utils.ParseDec("11"),
		utils.ParseCoins("1000_000000uatom,10000_000000uusd"))

	querier := exchangekeeper.Querier{Keeper: s.App.ExchangeKeeper}
	resp, err := querier.BestSwapExactAmountInRoutes(sdk.WrapSDKContext(s.Ctx), &types.QueryBestSwapExactAmountInRoutesRequest{
		Input:       "100000000ucre",
		OutputDenom: "uusd",
	})
	s.Require().NoError(err)

	s.Require().EqualValues([]uint64{2, 3}, resp.Routes)
	s.Require().Equal("972699534uusd", resp.Output.String())
	s.Require().Len(resp.Results, 2)
	s.Require().EqualValues(2, resp.Results[0].MarketId)
	s.Require().Equal("100000000ucre", resp.Results[0].Input.String())
	s.Require().Equal("95135825uatom", resp.Results[0].Output.String())
	s.Require().Equal("142919uatom", resp.Results[0].Fee.String())
	s.Require().EqualValues(3, resp.Results[1].MarketId)
	s.Require().Equal("95135825uatom", resp.Results[1].Input.String())
	s.Require().Equal("972699534uusd", resp.Results[1].Output.String())
	s.Require().Equal("1461242uusd", resp.Results[1].Fee.String())
}
