package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func (s *KeeperTestSuite) TestPairIndexes() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	pair2, found := s.keeper.GetPairByDenoms(s.ctx, "denom1", "denom2")
	s.Require().True(found)
	s.Require().Equal(pair.Id, pair2.Id)

	resp, err := s.querier.Pairs(sdk.WrapSDKContext(s.ctx), &types.QueryPairsRequest{
		Denoms: []string{"denom2"},
	})
	s.Require().NoError(err)
	s.Require().Len(resp.Pairs, 1)
	s.Require().Equal(pair.Id, resp.Pairs[0].Id)
}
