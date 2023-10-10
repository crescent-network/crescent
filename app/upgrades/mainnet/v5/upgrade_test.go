package v5_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	v5 "github.com/crescent-network/crescent/v5/app/upgrades/mainnet/v5"
	utils "github.com/crescent-network/crescent/v5/types"
	liquiditytypes "github.com/crescent-network/crescent/v5/x/liquidity/types"
	lpfarmtypes "github.com/crescent-network/crescent/v5/x/lpfarm/types"
)

type UpgradeTestSuite struct {
	testutil.TestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgradeV5() {
	enoughCoins := utils.ParseCoins(
		"1000000000000000ucre,1000000000000000uusd,1000000000000000uatom,1000000000000000stake")
	creatorAddr := s.FundedAccount(1, enoughCoins)

	pair, err := s.App.LiquidityKeeper.CreatePair(s.Ctx, liquiditytypes.NewMsgCreatePair(
		creatorAddr, "ucre", "uusd"))
	s.Require().NoError(err)
	pair.LastPrice = utils.ParseDecP("5")
	s.App.LiquidityKeeper.SetPair(s.Ctx, pair)
	oldPool1, err := s.App.LiquidityKeeper.CreatePool(s.Ctx, liquiditytypes.NewMsgCreatePool(
		creatorAddr, pair.Id, utils.ParseCoins("100_000000ucre,500_000000uusd")))
	s.Require().NoError(err)
	s.AssertEqual(utils.ParseDec("223606797.749978969640917367"), s.App.LPFarmKeeper.PoolRewardWeight(s.Ctx, oldPool1, pair))
	oldPool2, err := s.App.LiquidityKeeper.CreateRangedPool(s.Ctx, liquiditytypes.NewMsgCreateRangedPool(
		creatorAddr, pair.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"),
		utils.ParseDec("4"), utils.ParseDec("6"), utils.ParseDec("5")))
	s.Require().NoError(err)
	s.AssertEqual(utils.ParseDec("2118033995.149877930999785779"), s.App.LPFarmKeeper.PoolRewardWeight(s.Ctx, oldPool2, pair))
	lpfarmPlan, err := s.App.LPFarmKeeper.CreatePrivatePlan(s.Ctx, creatorAddr, "", []lpfarmtypes.RewardAllocation{
		lpfarmtypes.NewPairRewardAllocation(pair.Id, utils.ParseCoins("100_000000uatom")),
	}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))
	s.Require().NoError(err)
	s.FundAccount(lpfarmPlan.GetFarmingPoolAddress(), enoughCoins)
	s.NextBlock()

	// We do this so that the account is considered as a normal account,
	// not a module account.
	acc := s.App.AccountKeeper.GetAccount(s.Ctx, creatorAddr)
	_ = acc.SetSequence(1)
	_ = acc.SetPubKey(ed25519.GenPrivKey().PubKey())
	s.App.AccountKeeper.SetAccount(s.Ctx, acc)

	// Set the upgrade plan.
	upgradeHeight := s.Ctx.BlockHeight() + 1
	upgradePlan := upgradetypes.Plan{Name: v5.UpgradeName, Height: upgradeHeight}
	s.Require().NoError(s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, upgradePlan))
	_, havePlan := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(havePlan)

	// Let the upgrade happen.
	s.NextBlock()

	pool, found := s.App.AMMKeeper.GetPool(s.Ctx, 1)
	s.Require().True(found)
	poolState := s.App.AMMKeeper.MustGetPoolState(s.Ctx, pool.Id)
	s.AssertEqual(sdk.NewInt(2341640785), poolState.TotalLiquidity)
	s.AssertEqual(sdk.NewInt(2341640785), poolState.CurrentLiquidity)

	// Check if creating new market overwrites existing markets.
	s.Require().Equal(uint64(1), s.App.ExchangeKeeper.GetLastMarketId(s.Ctx))
	market := s.CreateMarket("uusd", "ucre")
	s.Require().Equal(uint64(2), market.Id)
	market, _ = s.App.ExchangeKeeper.GetMarket(s.Ctx, 1)
	s.Require().Equal("ucre", market.BaseDenom)
	s.Require().Equal("uusd", market.QuoteDenom)
}

func (s *UpgradeTestSuite) TestUpgradeV5Params() {
	creatorAddr := s.FundedAccount(1, utils.ParseCoins("1000_000000ucre,1000_000000stake"))

	var denoms []string
	for i := 0; i < 60; i++ {
		denom := fmt.Sprintf("denom%d", i+1)
		denoms = append(denoms, denom)
		s.FundAccount(creatorAddr, sdk.NewCoins(sdk.NewInt64Coin(denom, 1000_000000)))
	}
	// Dummy pair to make pairId != poolId
	_, err := s.App.LiquidityKeeper.CreatePair(s.Ctx, liquiditytypes.NewMsgCreatePair(
		creatorAddr, denoms[59], denoms[0]))
	s.Require().NoError(err)

	for i := 0; i < 59; i++ {
		pair, err := s.App.LiquidityKeeper.CreatePair(s.Ctx, liquiditytypes.NewMsgCreatePair(
			creatorAddr, denoms[i], denoms[i+1]))
		s.Require().NoError(err)
		_, err = s.App.LiquidityKeeper.CreatePool(s.Ctx, liquiditytypes.NewMsgCreatePool(
			creatorAddr, pair.Id,
			sdk.NewCoins(sdk.NewInt64Coin(denoms[i], 10_000000), sdk.NewInt64Coin(denoms[i+1], 10_000000))))
		s.Require().NoError(err)
	}

	// We do this so that the account is considered as a normal account,
	// not a module account.
	acc := s.App.AccountKeeper.GetAccount(s.Ctx, creatorAddr)
	_ = acc.SetSequence(1)
	_ = acc.SetPubKey(ed25519.GenPrivKey().PubKey())
	s.App.AccountKeeper.SetAccount(s.Ctx, acc)

	// Set the upgrade plan.
	upgradeHeight := s.Ctx.BlockHeight() + 1
	upgradePlan := upgradetypes.Plan{Name: v5.UpgradeName, Height: upgradeHeight}
	s.Require().NoError(s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, upgradePlan))
	_, havePlan := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(havePlan)

	// Let the upgrade happen.
	s.NextBlock()

	changedPairIds := maps.Keys(v5.ParamChanges)
	slices.Sort(changedPairIds)
	for _, pairId := range changedPairIds {
		change := v5.ParamChanges[pairId]
		market := s.App.ExchangeKeeper.MustGetMarket(s.Ctx, pairId)
		if change.MakerFeeRate != nil {
			s.AssertEqual(*change.MakerFeeRate, market.Fees.MakerFeeRate)
		} else {
			s.AssertEqual(sdk.NewDecWithPrec(1, 3), market.Fees.MakerFeeRate) // default
		}
		if change.TakerFeeRate != nil {
			s.AssertEqual(*change.TakerFeeRate, market.Fees.TakerFeeRate)
		} else {
			s.AssertEqual(sdk.NewDecWithPrec(2, 3), market.Fees.TakerFeeRate) // default
		}
		if change.MinOrderQuantity != nil {
			s.AssertEqual(*change.MinOrderQuantity, market.OrderQuantityLimits.Min)
		} else {
			s.AssertEqual(sdk.NewInt(10000), market.OrderQuantityLimits.Min) // default
		}
		if change.MinOrderQuote != nil {
			s.AssertEqual(*change.MinOrderQuote, market.OrderQuoteLimits.Min)
		} else {
			s.AssertEqual(sdk.NewInt(10000), market.OrderQuoteLimits.Min) // default
		}
		pool, found := s.App.AMMKeeper.GetPoolByMarket(s.Ctx, pairId)
		s.Require().True(found)
		if change.TickSpacing != nil {
			s.Require().EqualValues(*change.TickSpacing, pool.TickSpacing)
		} else {
			s.Require().EqualValues(uint32(50), pool.TickSpacing) // default
		}
	}
}
