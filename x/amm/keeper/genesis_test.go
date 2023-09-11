package keeper_test

import (
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestImportExportGenesis() {
	s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("0.005"))
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, 1, utils.ParseDec("0.003"), utils.ParseDec("0.007"),
		utils.ParseCoins("1000_000000ucre,1000_000000uusd"))
	s.CreatePrivateFarmingPlan(
		utils.TestAddress(1), "Farming plan", utils.TestAddress(2), []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(1, utils.ParseCoins("10_000000uatom")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000uatom"), true)

	s.NextBlock()
	s.NextBlock()

	genState := s.keeper.ExportGenesis(s.Ctx)
	bz := s.App.AppCodec().MustMarshalJSON(genState)

	s.SetupTest()
	var genState2 types.GenesisState
	s.App.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.keeper.InitGenesis(s.Ctx, genState2)
	genState3 := s.keeper.ExportGenesis(s.Ctx)
	s.Require().Equal(*genState, *genState3)
}

func (s *KeeperTestSuite) TestImportExportGenesisEmpty() {
	genState := s.keeper.ExportGenesis(s.Ctx)

	var genState2 types.GenesisState
	bz := s.App.AppCodec().MustMarshalJSON(genState)
	s.App.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.keeper.InitGenesis(s.Ctx, genState2)

	genState3 := s.keeper.ExportGenesis(s.Ctx)
	s.Require().Equal(*genState, genState2)
	s.Require().Equal(genState2, *genState3)
}
