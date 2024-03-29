package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var enoughCoins = utils.ParseCoins(
	"1000000_000000000000000000ucre,1000000_000000000000000000uusd,1000000_000000000000000000uatom,1000000_000000000000000000stake")

type KeeperTestSuite struct {
	testutil.TestSuite
	keeper    keeper.Keeper
	msgServer types.MsgServer
	querier   keeper.Querier
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.TestSuite.SetupTest()
	s.keeper = s.App.AMMKeeper
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1ucre,1uusd,1uatom")) // make positive supplies
}

func (s *KeeperTestSuite) CreateMarketAndPool(baseDenom, quoteDenom string, price sdk.Dec) (market exchangetypes.Market, pool types.Pool) {
	market = s.CreateMarket(baseDenom, quoteDenom)
	pool = s.CreatePool(market.Id, price)
	return market, pool
}
