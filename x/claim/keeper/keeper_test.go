package keeper_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	squadapp "github.com/cosmosquad-labs/squad/app"
	"github.com/cosmosquad-labs/squad/x/claim/keeper"
	"github.com/cosmosquad-labs/squad/x/claim/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app       *squadapp.SquadApp
	ctx       sdk.Context
	keeper    keeper.Keeper
	querier   keeper.Querier
	msgServer types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = squadapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.keeper = s.app.ClaimKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
}

//
// Below are just shortcuts to internal functions.
//

func (s *KeeperTestSuite) createAirdrop(
	airdropId uint64,
	sourceCoins sdk.Coins,
	startTime time.Time,
	endTime time.Time,
	fund bool,
) types.Airdrop {
	sourceAddr := types.SourceAddress(airdropId)

	if fund {
		s.fundAddr(sourceAddr, sourceCoins)
	}

	s.keeper.SetAirdrop(s.ctx, types.Airdrop{
		AirdropId:          airdropId,
		SourceAddress:      sourceAddr.String(),
		TerminationAddress: s.addr(6).String(),
		StartTime:          startTime,
		EndTime:            endTime,
	})

	airdrop, found := s.keeper.GetAirdrop(s.ctx, airdropId)
	s.Require().True(found)

	return airdrop
}

func (s *KeeperTestSuite) createClaimRecord(
	airdropId uint64,
	recipient sdk.AccAddress,
	initialClaimableCoins sdk.Coins,
	claimableCoins sdk.Coins,
	actions []types.Action,
) types.ClaimRecord {
	s.keeper.SetClaimRecord(s.ctx, types.ClaimRecord{
		AirdropId:             airdropId,
		Recipient:             recipient.String(),
		InitialClaimableCoins: initialClaimableCoins,
		ClaimableCoins:        claimableCoins,
		Actions:               actions,
	})

	r, found := s.keeper.GetClaimRecordByRecipient(s.ctx, airdropId, recipient)
	s.Require().True(found)

	return r
}

//
// Below are useful helpers to write test code easily.
//

func (s *KeeperTestSuite) getBalance(addr sdk.AccAddress, denom string) sdk.Coin {
	return s.app.BankKeeper.GetBalance(s.ctx, addr, denom)
}

func (s *KeeperTestSuite) addr(addrNum int) sdk.AccAddress {
	addr := make(sdk.AccAddress, 20)
	binary.PutVarint(addr, int64(addrNum))
	return addr
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, coins sdk.Coins) {
	err := squadapp.FundAccount(s.app.BankKeeper, s.ctx, addr, coins)
	s.Require().NoError(err)
}

func parseCoins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return coins
}

func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
