package keeper_test

import (
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestAllAirdrops() {
	conditions := []types.ConditionType{
		types.ConditionTypeDeposit,
		types.ConditionTypeSwap,
		types.ConditionTypeLiquidStake,
		types.ConditionTypeVote,
	}

	s.createAirdrop(1, s.addr(1), utils.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(2, s.addr(2), utils.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(3, s.addr(3), utils.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(4, s.addr(4), utils.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)

	airdrops := s.keeper.GetAllAirdrops(s.ctx)
	s.Require().Len(airdrops, 4)
}

func (s *KeeperTestSuite) TestAllClaimRecords() {
	airdrop := s.createAirdrop(
		1,
		s.addr(0),
		utils.ParseCoins("1000000000denom1"),
		[]types.ConditionType{
			types.ConditionTypeDeposit,
			types.ConditionTypeSwap,
			types.ConditionTypeLiquidStake,
			types.ConditionTypeVote,
		},
		s.ctx.BlockTime(),
		s.ctx.BlockTime().AddDate(0, 1, 0),
		true,
	)

	s.createClaimRecord(airdrop.Id, s.addr(0), utils.ParseCoins("300000000denom1"), utils.ParseCoins("300000000denom1"), []types.ConditionType{})
	s.createClaimRecord(airdrop.Id, s.addr(1), utils.ParseCoins("300000000denom1"), utils.ParseCoins("300000000denom1"), []types.ConditionType{})
	s.createClaimRecord(airdrop.Id, s.addr(2), utils.ParseCoins("400000000denom1"), utils.ParseCoins("400000000denom1"), []types.ConditionType{})

	records := s.keeper.GetAllClaimRecordsByAirdropId(s.ctx, airdrop.Id)
	s.Require().Len(records, 3)
}
