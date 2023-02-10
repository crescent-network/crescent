package antehandlers_test

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	bootstraptypes "github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

// Test logic around account number checking with one signer and many signers.
func (suite *AnteTestSuite) TestAnteHandlerValidateProposer() {
	suite.SetupTest(false) // reset

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(2)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	baseContent := bootstraptypes.NewBootstrapProposal(
		"test title",
		"test description",
		"",
		sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(1))),
		"quote",
		sdk.MustNewDecFromStr("0.1"),
		sdk.MustNewDecFromStr("0.1"),
		uint64(1),
		uint64(1),
		[]bootstraptypes.InitialOrder{},
	)

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []TestCase{
		{
			"valid proposer",
			func() {
				baseContent.ProposerAddress = accounts[0].acc.GetAddress().String()
				msg, err := govtypes.NewMsgSubmitProposal(baseContent, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), accounts[0].acc.GetAddress())
				suite.Require().NoError(err)
				msgs = []sdk.Msg{msg}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			true,
			nil,
		},
		{
			"invalid proposer",
			func() {
				baseContent.ProposerAddress = accounts[0].acc.GetAddress().String()
				msg, err := govtypes.NewMsgSubmitProposal(baseContent, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), accounts[1].acc.GetAddress())
				suite.Require().NoError(err)
				msgs = []sdk.Msg{msg}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrInvalidRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate()

			suite.RunTestCase(privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}
