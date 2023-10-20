package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestImportExportGenesis() {
	s.CreateMarket("ucre", "uusd")
	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(1, enoughCoins)
	s.PlaceLimitOrder(1, ordererAddr1, true, utils.ParseDec("4.9"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceMMLimitOrder(1, ordererAddr2, false, utils.ParseDec("5"), sdk.NewInt(20_000000), time.Hour)

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
