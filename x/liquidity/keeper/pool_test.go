package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func (s *KeeperTestSuite) TestExecuteDepositRequestWithDisabledPool() {
	k, ctx := s.keeper, s.ctx

	// Create a pool.
	depositXCoin, depositYCoin := parseCoin("1000000denom1"), parseCoin("1000000denom2")
	depositCoins := sdk.NewCoins(depositXCoin, depositYCoin)
	pool := s.createPool(s.addr(0), depositXCoin, depositYCoin)

	// Send the pool's balance to an attacker's address.
	// In normal situations this wouldn't happen but there's still such threat.
	err := s.app.BankKeeper.SendCoins(ctx, pool.GetReserveAddress(), s.addr(10), s.getBalances(pool.GetReserveAddress()))
	s.Require().NoError(err)

	// Create a depositor.
	depositor := s.addr(1)
	depositXCoin, depositYCoin = parseCoin("1000000denom1"), parseCoin("10000000denom2")
	depositCoins = sdk.NewCoins(depositXCoin, depositYCoin)
	s.fundAddr(depositor, depositCoins)

	// The depositor deposits coins but this will fail because the pool
	// is treated as disabled.
	msg := types.NewMsgDepositBatch(depositor, pool.Id, depositXCoin, depositYCoin)
	req, err := k.DepositBatch(ctx, msg)
	s.Require().NoError(err)
	err = k.ExecuteDepositRequest(ctx, req)
	s.Require().NoError(err)
	req, _ = k.GetDepositRequest(ctx, req.PoolId, req.Id)
	s.Require().False(req.Succeeded)

	// After refunding coins, the depositor will get back his coins as-is.
	k.RefundAndDeleteDepositRequestsToBeDeleted(ctx)
	s.Require().True(coinsEq(depositCoins, s.getBalances(depositor)))
}
