package keeper_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v2/app"
	"github.com/crescent-network/crescent/v2/x/claim/keeper"
	"github.com/crescent-network/crescent/v2/x/claim/types"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
	liquidstakingtypes "github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app       *chain.App
	ctx       sdk.Context
	keeper    keeper.Keeper
	querier   keeper.Querier
	msgServer types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.keeper = s.app.ClaimKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
}

//
// Below are just shortcuts to internal functions.
//

func (s *KeeperTestSuite) createAirdrop(
	id uint64,
	sourceAddr sdk.AccAddress,
	sourceCoins sdk.Coins,
	conditions []types.ConditionType,
	startTime time.Time,
	endTime time.Time,
	fund bool,
) types.Airdrop {
	if fund {
		s.fundAddr(sourceAddr, sourceCoins)
	}

	s.keeper.SetAirdrop(s.ctx, types.Airdrop{
		Id:            id,
		SourceAddress: sourceAddr.String(),
		Conditions:    conditions,
		StartTime:     startTime,
		EndTime:       endTime,
	})

	airdrop, found := s.keeper.GetAirdrop(s.ctx, id)
	s.Require().True(found)

	return airdrop
}

func (s *KeeperTestSuite) createClaimRecord(
	airdropId uint64,
	recipient sdk.AccAddress,
	initialClaimableCoins sdk.Coins,
	claimableCoins sdk.Coins,
	claimedConditions []types.ConditionType,
) types.ClaimRecord {
	s.keeper.SetClaimRecord(s.ctx, types.ClaimRecord{
		AirdropId:             airdropId,
		Recipient:             recipient.String(),
		InitialClaimableCoins: initialClaimableCoins,
		ClaimableCoins:        claimableCoins,
		ClaimedConditions:     claimedConditions,
	})

	r, found := s.keeper.GetClaimRecordByRecipient(s.ctx, airdropId, recipient)
	s.Require().True(found)

	return r
}

func (s *KeeperTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string, fund bool) liquiditytypes.Pair {
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, params.PairCreationFee)
	}
	pair, err := s.app.LiquidityKeeper.CreatePair(s.ctx, liquiditytypes.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

func (s *KeeperTestSuite) createPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins, fund bool) liquiditytypes.Pool {
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, depositCoins.Add(params.PoolCreationFee...))
	}
	pool, err := s.app.LiquidityKeeper.CreatePool(s.ctx, liquiditytypes.NewMsgCreatePool(creator, pairId, depositCoins))
	s.Require().NoError(err)
	return pool
}

func (s *KeeperTestSuite) deposit(depositor sdk.AccAddress, poolId uint64, depositCoins sdk.Coins, fund bool) liquiditytypes.DepositRequest {
	if fund {
		s.fundAddr(depositor, depositCoins)
	}
	req, err := s.app.LiquidityKeeper.Deposit(s.ctx, liquiditytypes.NewMsgDeposit(depositor, poolId, depositCoins))
	s.Require().NoError(err)
	return req
}

func (s *KeeperTestSuite) limitOrder(
	orderer sdk.AccAddress, pairId uint64, dir liquiditytypes.OrderDirection,
	price sdk.Dec, amt sdk.Int, orderLifespan time.Duration, fund bool) liquiditytypes.Order {
	pair, found := s.app.LiquidityKeeper.GetPair(s.ctx, pairId)
	s.Require().True(found)

	var offerCoin sdk.Coin
	var demandCoinDenom string
	switch dir {
	case liquiditytypes.OrderDirectionBuy:
		offerCoin = sdk.NewCoin(pair.QuoteCoinDenom, price.MulInt(amt).Ceil().TruncateInt())
		demandCoinDenom = pair.BaseCoinDenom
	case liquiditytypes.OrderDirectionSell:
		offerCoin = sdk.NewCoin(pair.BaseCoinDenom, amt)
		demandCoinDenom = pair.QuoteCoinDenom
	}

	if fund {
		s.fundAddr(orderer, sdk.NewCoins(offerCoin))
	}

	req, err := s.app.LiquidityKeeper.LimitOrder(s.ctx, liquiditytypes.NewMsgLimitOrder(
		orderer, pairId, dir, offerCoin, demandCoinDenom,
		price, amt, orderLifespan),
	)
	s.Require().NoError(err)

	return req
}

func (s *KeeperTestSuite) sellLimitOrder(
	orderer sdk.AccAddress, pairId uint64, price sdk.Dec,
	amt sdk.Int, orderLifespan time.Duration, fund bool) liquiditytypes.Order {
	return s.limitOrder(
		orderer, pairId, liquiditytypes.OrderDirectionSell, price, amt, orderLifespan, fund)
}

func (s *KeeperTestSuite) createWhitelistedValidators(powers []int64) ([]sdk.AccAddress, []sdk.ValAddress, []cryptotypes.PubKey) {
	params := s.app.LiquidStakingKeeper.GetParams(s.ctx)

	num := len(powers)
	addrs := chain.AddTestAddrsIncremental(s.app, s.ctx, num, sdk.NewInt(1000000000))
	valAddrs := chain.ConvertAddrsToValAddrs(addrs)
	pks := chain.CreateTestPubKeys(num)

	for i, power := range powers {
		val, err := stakingtypes.NewValidator(valAddrs[i], pks[i], stakingtypes.Description{})
		s.Require().NoError(err)

		s.app.StakingKeeper.SetValidator(s.ctx, val)
		err = s.app.StakingKeeper.SetValidatorByConsAddr(s.ctx, val)
		s.Require().NoError(err)

		s.app.StakingKeeper.SetNewValidatorByPowerIndex(s.ctx, val)
		s.app.StakingKeeper.AfterValidatorCreated(s.ctx, val.GetOperator())
		newShares, err := s.app.StakingKeeper.Delegate(s.ctx, addrs[i], sdk.NewInt(power), stakingtypes.Unbonded, val, true)
		s.Require().NoError(err)
		s.Require().Equal(newShares.TruncateInt(), sdk.NewInt(power))
	}

	whitelistedVals := []liquidstakingtypes.WhitelistedValidator{}

	// Add active validator
	for _, valAddr := range valAddrs {
		whitelistedVals = append(whitelistedVals, liquidstakingtypes.WhitelistedValidator{
			ValidatorAddress: valAddr.String(),
			TargetWeight:     sdk.NewInt(1),
		})
	}
	params.WhitelistedValidators = whitelistedVals

	s.app.LiquidStakingKeeper.SetParams(s.ctx, params)
	s.app.LiquidStakingKeeper.UpdateLiquidValidatorSet(s.ctx)

	return addrs, valAddrs, pks
}

func (s *KeeperTestSuite) liquidStaking(liquidStaker sdk.AccAddress, stakingAmt sdk.Int, fund bool) {
	if fund {
		fundCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
		s.fundAddr(liquidStaker, fundCoins)
	}

	ctx, writeCache := s.ctx.CacheContext()
	lsKeeper := s.app.LiquidStakingKeeper

	params := lsKeeper.GetParams(ctx)
	btokenBalanceBefore := s.app.BankKeeper.GetBalance(ctx, liquidStaker, params.LiquidBondDenom).Amount
	newShares, bTokenMintAmt, err := lsKeeper.LiquidStake(
		ctx,
		liquidstakingtypes.LiquidStakingProxyAcc,
		liquidStaker,
		sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt),
	)
	s.Require().NoError(err)

	btokenBalanceAfter := s.app.BankKeeper.GetBalance(ctx, liquidStaker, params.LiquidBondDenom).Amount
	s.Require().NoError(err)
	s.NotEqualValues(newShares, sdk.ZeroDec())
	s.Require().EqualValues(bTokenMintAmt, btokenBalanceAfter.Sub(btokenBalanceBefore))

	writeCache()
}

func (s *KeeperTestSuite) createTextProposal(proposer sdk.AccAddress, title string, description string) govtypes.Proposal {
	content := govtypes.NewTextProposal(title, description)
	proposal, err := s.app.GovKeeper.SubmitProposal(s.ctx, content)
	s.Require().NoError(err)

	proposal.Status = govtypes.StatusVotingPeriod
	s.app.GovKeeper.SetProposal(s.ctx, proposal)

	proposal, found := s.app.GovKeeper.GetProposal(s.ctx, 1)
	s.Require().True(found)

	return proposal
}

func (s *KeeperTestSuite) vote(voter sdk.AccAddress, proposalId uint64, option govtypes.VoteOption) govtypes.Vote {
	err := s.app.GovKeeper.AddVote(s.ctx, proposalId, voter, govtypes.NewNonSplitVoteOption(option))
	s.Require().NoError(err)

	vote, found := s.app.GovKeeper.GetVote(s.ctx, proposalId, voter)
	s.Require().True(found)

	return vote
}

//
// Below are useful helpers to write test code easily.
//

func (s *KeeperTestSuite) getBalance(addr sdk.AccAddress, denom string) sdk.Coin {
	return s.app.BankKeeper.GetBalance(s.ctx, addr, denom)
}

func (s *KeeperTestSuite) getAllBalances(addr sdk.AccAddress) sdk.Coins {
	return s.app.BankKeeper.GetAllBalances(s.ctx, addr)
}

func (s *KeeperTestSuite) addr(addrNum int) sdk.AccAddress {
	addr := make(sdk.AccAddress, 20)
	binary.PutVarint(addr, int64(addrNum))
	return addr
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, coins sdk.Coins) {
	err := chain.FundAccount(s.app.BankKeeper, s.ctx, addr, coins)
	s.Require().NoError(err)
}

func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
