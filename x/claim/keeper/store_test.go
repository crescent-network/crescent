package keeper_test

import (
	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestSetAirdropId() {
	id := s.keeper.GetLastAirdropId(s.ctx)
	s.Require().Equal(uint64(0), id)

	s.createAirdrop(1, parseCoins("1000000000denom1"), s.ctx.BlockTime(), squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"), true)
	s.createAirdrop(2, parseCoins("1000000000denom1"), s.ctx.BlockTime(), squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"), true)
	s.createAirdrop(3, parseCoins("1000000000denom1"), s.ctx.BlockTime(), squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"), true)

	id = s.keeper.GetLastAirdropId(s.ctx)
	s.Require().Equal(uint64(3), id)
}

func (s *KeeperTestSuite) TestAllAirdrops() {
	s.createAirdrop(1, parseCoins("1000000000denom1"), s.ctx.BlockTime(), squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"), true)
	s.createAirdrop(2, parseCoins("5000000000denom1"), s.ctx.BlockTime(), squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"), true)
	s.createAirdrop(3, parseCoins("10000000000denom1"), s.ctx.BlockTime(), squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"), true)
	s.createAirdrop(4, parseCoins("7000000000denom1"), s.ctx.BlockTime(), squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"), true)

	airdrops := s.keeper.GetAllAirdrops(s.ctx)
	s.Require().Len(airdrops, 4)
}

func (s *KeeperTestSuite) TestAirdropStartAndEndTime() {
	airdrop := s.createAirdrop(
		1,
		parseCoins("1000000000denom1"),
		squadtypes.MustParseRFC3339("2022-02-01T00:00:00Z"),
		squadtypes.MustParseRFC3339("2022-07-01T00:00:00Z"),
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
		parseCoins("1000000000denom1"),
		squadtypes.MustParseRFC3339("2022-02-01T00:00:00Z"),
		squadtypes.MustParseRFC3339("2022-07-01T00:00:00Z"),
		true,
	)

	s.createClaimRecord(airdrop.Id, s.addr(0), parseCoins("300000000denom1"), parseCoins("300000000denom1"),
		[]types.Action{
			{ActionType: types.ActionTypeDeposit, Claimed: false},
			{ActionType: types.ActionTypeSwap, Claimed: false},
			{ActionType: types.ActionTypeFarming, Claimed: false}},
	)
	s.createClaimRecord(airdrop.Id, s.addr(1), parseCoins("300000000denom1"), parseCoins("300000000denom1"),
		[]types.Action{
			{ActionType: types.ActionTypeDeposit, Claimed: true},
			{ActionType: types.ActionTypeSwap, Claimed: false},
			{ActionType: types.ActionTypeFarming, Claimed: false}},
	)
	s.createClaimRecord(airdrop.Id, s.addr(2), parseCoins("400000000denom1"), parseCoins("400000000denom1"),
		[]types.Action{
			{ActionType: types.ActionTypeDeposit, Claimed: false},
			{ActionType: types.ActionTypeSwap, Claimed: true},
			{ActionType: types.ActionTypeFarming, Claimed: true}},
	)

	records := s.keeper.GetAllClaimRecordsByAirdropId(s.ctx, airdrop.Id)
	s.Require().Len(records, 3)
}
