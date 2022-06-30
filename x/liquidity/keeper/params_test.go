package keeper_test

import (
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func (s *KeeperTestSuite) TestGetBatchSize() {
	s.Require().EqualValues(types.DefaultBatchSize, s.keeper.GetBatchSize(s.ctx))
}

func (s *KeeperTestSuite) TestGetTickPrecision() {
	s.Require().EqualValues(types.DefaultTickPrecision, s.keeper.GetTickPrecision(s.ctx))
}

func (s *KeeperTestSuite) TestGetFeeCollector() {
	s.Require().EqualValues(types.DefaultFeeCollectorAddress, s.keeper.GetFeeCollector(s.ctx))
}

func (s *KeeperTestSuite) TestGetDustCollector() {
	s.Require().EqualValues(types.DefaultDustCollectorAddress, s.keeper.GetDustCollector(s.ctx))
}

func (s *KeeperTestSuite) TestGetMinInitialPoolCoinSupply() {
	s.Require().EqualValues(types.DefaultMinInitialPoolCoinSupply, s.keeper.GetMinInitialPoolCoinSupply(s.ctx))
}

func (s *KeeperTestSuite) TestGetPairCreationFee() {
	s.Require().EqualValues(types.DefaultPairCreationFee, s.keeper.GetPairCreationFee(s.ctx))
}

func (s *KeeperTestSuite) TestGetPoolCreationFee() {
	s.Require().EqualValues(types.DefaultPoolCreationFee, s.keeper.GetPoolCreationFee(s.ctx))
}

func (s *KeeperTestSuite) TestGetMinInitialDepositAmount() {
	s.Require().EqualValues(types.DefaultMinInitialDepositAmount, s.keeper.GetMinInitialDepositAmount(s.ctx))
}

func (s *KeeperTestSuite) TestGetMaxPriceLimitRatio() {
	s.Require().EqualValues(types.DefaultMaxPriceLimitRatio, s.keeper.GetMaxPriceLimitRatio(s.ctx))
}

func (s *KeeperTestSuite) TestGetMaxOrderLifespan() {
	s.Require().EqualValues(types.DefaultMaxOrderLifespan, s.keeper.GetMaxOrderLifespan(s.ctx))
}

func (s *KeeperTestSuite) TestGetWithdrawFeeRate() {
	s.Require().EqualValues(types.DefaultWithdrawFeeRate, s.keeper.GetWithdrawFeeRate(s.ctx))
}

func (s *KeeperTestSuite) TestGetDepositExtraGas() {
	s.Require().EqualValues(types.DefaultDepositExtraGas, s.keeper.GetDepositExtraGas(s.ctx))
}

func (s *KeeperTestSuite) TestGetWithdrawExtraGas() {
	s.Require().EqualValues(types.DefaultWithdrawExtraGas, s.keeper.GetWithdrawExtraGas(s.ctx))
}

func (s *KeeperTestSuite) TestGetOrderExtraGas() {
	s.Require().EqualValues(types.DefaultOrderExtraGas, s.keeper.GetOrderExtraGas(s.ctx))
}
