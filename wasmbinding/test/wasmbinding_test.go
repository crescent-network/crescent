package wasmbinding_test

import (
	"encoding/binary"
	"os"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v4/app"
	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
	minttypes "github.com/crescent-network/crescent/v4/x/mint/types"

	"github.com/stretchr/testify/suite"
)

type WasmBindingTestSuite struct {
	suite.Suite

	app *chain.App
	ctx sdk.Context
}

func TestWasmBindingTestSuite(t *testing.T) {
	suite.Run(t, new(WasmBindingTestSuite))
}

func (s *WasmBindingTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{
		Height:  1,
		ChainID: "crescent-1",
		Time:    time.Now().UTC(),
	})
}

// Below are useful helpers to write test code easily.
func (s *WasmBindingTestSuite) addr(addrNum int) sdk.AccAddress {
	addr := make(sdk.AccAddress, 20)
	binary.PutVarint(addr, int64(addrNum))
	return addr
}

func (s *WasmBindingTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	s.T().Helper()
	err := s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr, amt)
	s.Require().NoError(err)
}

func (s *WasmBindingTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string, fund bool) liquiditytypes.Pair {
	s.T().Helper()
	if fund {
		s.fundAddr(creator, s.app.LiquidityKeeper.GetPairCreationFee(s.ctx))
	}
	msg := liquiditytypes.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom)
	s.Require().NoError(msg.ValidateBasic())
	pair, err := s.app.LiquidityKeeper.CreatePair(s.ctx, msg)
	s.Require().NoError(err)
	return pair
}

func (s *WasmBindingTestSuite) storeCode(creator sdk.AccAddress, filePath string) {
	s.T().Helper()
	wasmCode, err := os.ReadFile(filePath)
	s.Require().NoError(err)

	content := wasmtypes.StoreCodeProposalFixture(func(p *wasmtypes.StoreCodeProposal) {
		p.Title = "Store Sample Contract"
		p.Description = "Store Sample Contract Description"
		p.RunAs = creator.String()
		p.WASMByteCode = wasmCode
	})

	storedProposal, err := s.app.GovKeeper.SubmitProposal(s.ctx, content)
	s.Require().NoError(err)

	handler := s.app.GovKeeper.Router().GetRoute(storedProposal.ProposalRoute())
	err = handler(s.ctx, storedProposal.GetContent())
	s.Require().NoError(err)
}

func (s *WasmBindingTestSuite) instantiateEmptyContract(creator, admin sdk.AccAddress) sdk.AccAddress {
	s.T().Helper()
	codeID := uint64(1)
	initMsgBz := []byte("{}")
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(s.app.WasmKeeper)
	addr, _, err := contractKeeper.Instantiate(s.ctx, codeID, creator, admin, initMsgBz, "demo contract", nil)
	s.Require().NoError(err)
	return addr
}
