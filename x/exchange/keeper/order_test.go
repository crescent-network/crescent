package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
	minttypes "github.com/crescent-network/crescent/v5/x/mint/types"
)

func (s *KeeperTestSuite) TestFoo() {
	aliceAddr := utils.TestAddress(1)
	bobAddr := utils.TestAddress(2)

	market, err := s.k.CreateSpotMarket(s.ctx, utils.TestAddress(3), "ucre", "uusd")
	s.Require().NoError(err)

	s.Require().NoError(s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, utils.ParseCoins("1000000ucre,1000000uusd")))
	s.Require().NoError(s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, aliceAddr, utils.ParseCoins("1000000ucre,1000000uusd")))
	s.Require().NoError(s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, utils.ParseCoins("1000000ucre,1000000uusd")))
	s.Require().NoError(s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, bobAddr, utils.ParseCoins("1000000ucre,1000000uusd")))

	_, _, err = s.k.PlaceSpotOrder(
		s.ctx, aliceAddr, market.Id, true, utils.ParseDecP("100"), sdk.NewInt(1000))
	s.Require().NoError(err)
	_, _, err = s.k.PlaceSpotOrder(
		s.ctx, aliceAddr, market.Id, true, utils.ParseDecP("99"), sdk.NewInt(1000))
	s.Require().NoError(err)
	_, _, err = s.k.PlaceSpotOrder(
		s.ctx, aliceAddr, market.Id, true, utils.ParseDecP("97"), sdk.NewInt(1000))
	s.Require().NoError(err)

	_, _, err = s.k.PlaceSpotOrder(
		s.ctx, bobAddr, market.Id, false, utils.ParseDecP("98"), sdk.NewInt(1500))
	s.Require().NoError(err)

	s.Require().Equal("1001500ucre,704000uusd", s.app.BankKeeper.GetAllBalances(s.ctx, aliceAddr).String())
	s.Require().Equal("998500ucre,1149500uusd", s.app.BankKeeper.GetAllBalances(s.ctx, bobAddr).String())
	s.Require().Equal("146500uusd", s.app.BankKeeper.GetAllBalances(s.ctx, s.app.AccountKeeper.GetModuleAddress(types.ModuleName)).String())

	_, _, err = s.k.PlaceSpotOrder(
		s.ctx, bobAddr, market.Id, false, utils.ParseDecP("96"), sdk.NewInt(1500))
	s.Require().NoError(err)

	s.Require().Equal("1003000ucre,704000uusd", s.app.BankKeeper.GetAllBalances(s.ctx, aliceAddr).String())
	s.Require().Equal("997000ucre,1296000uusd", s.app.BankKeeper.GetAllBalances(s.ctx, bobAddr).String())
	s.Require().Equal("", s.app.BankKeeper.GetAllBalances(s.ctx, s.app.AccountKeeper.GetModuleAddress(types.ModuleName)).String())
}