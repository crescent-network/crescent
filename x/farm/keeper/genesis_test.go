package keeper_test

import (
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func (s *KeeperTestSuite) TestImportExportGenesis() {
	pair := s.createPair("denom1", "denom2")
	pool := s.createPool(pair.Id, utils.ParseCoins("1000_000000denom1,1000_000000denom2"))
	s.createPrivatePlan([]types.RewardAllocation{
		{
			PairId:        pair.Id,
			RewardsPerDay: utils.ParseCoins("100_000000stake"),
		},
	})
	farmerAddr := utils.TestAddress(0)
	s.deposit(farmerAddr, pool.Id, utils.ParseCoins("100_000000denom1,100_000000denom2"))
	s.nextBlock()
	_, err := s.keeper.Farm(s.ctx, farmerAddr, s.getBalance(farmerAddr, pool.PoolCoinDenom))
	s.Require().NoError(err)
	s.nextBlock()
	_, err = s.keeper.Harvest(s.ctx, farmerAddr, pool.PoolCoinDenom)
	s.Require().NoError(err)
	s.nextBlock()

	genState := s.keeper.ExportGenesis(s.ctx)
	bz := s.app.AppCodec().MustMarshalJSON(genState)

	s.SetupTest()
	var genState2 types.GenesisState
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.keeper.InitGenesis(s.ctx, genState2)
	genState3 := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(*genState, *genState3)
}

func (s *KeeperTestSuite) TestImportExportGenesisEmpty() {
	genState := s.keeper.ExportGenesis(s.ctx)

	var genState2 types.GenesisState
	bz := s.app.AppCodec().MustMarshalJSON(genState)
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.keeper.InitGenesis(s.ctx, genState2)

	genState3 := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(*genState, genState2)
	s.Require().Equal(genState2, *genState3)
}
