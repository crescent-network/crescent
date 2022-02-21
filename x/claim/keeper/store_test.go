package keeper_test

import (
	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestSetAirdropId() {
	id := s.keeper.GetLastAirdropId(s.ctx)
	s.Require().Equal(uint64(0), id)

	conditions := []types.ConditionType{
		types.ConditionTypeDeposit,
		types.ConditionTypeSwap,
		types.ConditionTypeFarming,
	}

	s.createAirdrop(1, s.addr(1), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(2, s.addr(2), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(3, s.addr(3), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)

	id = s.keeper.GetLastAirdropId(s.ctx)
	s.Require().Equal(uint64(3), id)
}

func (s *KeeperTestSuite) TestAllAirdrops() {
	conditions := []types.ConditionType{
		types.ConditionTypeDeposit,
		types.ConditionTypeSwap,
		types.ConditionTypeFarming,
	}

	s.createAirdrop(1, s.addr(1), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(2, s.addr(2), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(3, s.addr(3), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(4, s.addr(4), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)

	airdrops := s.keeper.GetAllAirdrops(s.ctx)
	s.Require().Len(airdrops, 4)
}

func (s *KeeperTestSuite) TestAirdropStartAndEndTime() {
	airdrop := s.createAirdrop(
		1,
		s.addr(0),
		squad.ParseCoins("1000000000denom1"),
		[]types.ConditionType{
			types.ConditionTypeDeposit,
			types.ConditionTypeSwap,
			types.ConditionTypeFarming,
		},
		s.ctx.BlockTime(),
		s.ctx.BlockTime().AddDate(0, 1, 0),
		true,
	)

	_, found := s.keeper.GetAirdrop(s.ctx, airdrop.Id)
	s.Require().True(found)

	startTime := s.keeper.GetStartTime(s.ctx, airdrop.Id)
	s.Require().Equal(airdrop.StartTime, *startTime)

	endTime := s.keeper.GetEndTime(s.ctx, airdrop.Id)
	s.Require().Equal(airdrop.EndTime, *endTime)
}

func (s *KeeperTestSuite) TestAllClaimRecords() {
	airdrop := s.createAirdrop(
		1,
		s.addr(0),
		squad.ParseCoins("1000000000denom1"),
		[]types.ConditionType{
			types.ConditionTypeDeposit,
			types.ConditionTypeSwap,
			types.ConditionTypeFarming,
		},
		s.ctx.BlockTime(),
		s.ctx.BlockTime().AddDate(0, 1, 0),
		true,
	)

	s.createClaimRecord(airdrop.Id, s.addr(0), squad.ParseCoins("300000000denom1"), squad.ParseCoins("300000000denom1"), []types.ConditionType{})
	s.createClaimRecord(airdrop.Id, s.addr(1), squad.ParseCoins("300000000denom1"), squad.ParseCoins("300000000denom1"), []types.ConditionType{})
	s.createClaimRecord(airdrop.Id, s.addr(2), squad.ParseCoins("400000000denom1"), squad.ParseCoins("400000000denom1"), []types.ConditionType{})

	records := s.keeper.GetAllClaimRecordsByAirdropId(s.ctx, airdrop.Id)
	s.Require().Len(records, 3)
}
