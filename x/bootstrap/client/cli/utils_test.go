package cli_test

//
//import (
//	"testing"
//
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	"github.com/stretchr/testify/require"
//
//	"github.com/cosmos/cosmos-sdk/testutil"
//
//	"github.com/crescent-network/crescent/v4/app/params"
//
//	"github.com/crescent-network/crescent/v4/x/bootstrap/client/cli"
//)
//
//func TestParseBootstrapProposal(t *testing.T) {
//	encodingConfig := params.MakeTestEncodingConfig()
//
//	okJSON := testutil.WriteToNewTempFile(t, `
//{
//  "title": "Market Maker Proposal",
//  "description": "Are you ready to market making?",
//  "inclusions": [
//    {
//      "address": "cosmos1vqac3p8fl4kez7ehjz8eltugd2fm67pckpl7pn",
//      "pair_id": "1"
//    }
//  ],
//  "exclusions": [],
//  "rejections": [
//    {
//      "address": "cosmos1vqac3p8fl4kez7ehjz8eltugd2fm67pckpl7pn",
//      "pair_id": "2"
//    }
//  ],
//  "distributions": [
//    {
//      "address": "cosmos1vqac3p8fl4kez7ehjz8eltugd2fm67pckpl7pn",
//      "pair_id": "1",
//      "amount": [
//        {
//          "denom": "stake",
//          "amount": "100000000"
//        }
//      ]
//    }
//  ]
//}
//`)
//
//	proposal, err := cli.ParseBootstrapProposal(encodingConfig.Marshaler, okJSON.Name())
//	require.NoError(t, err)
//
//	require.Equal(t, "Market Maker Proposal", proposal.Title)
//	require.Equal(t, "Are you ready to market making?", proposal.Description)
//	require.Equal(t, uint64(1), proposal.Inclusions[0].PairId)
//	require.Equal(t, "cosmos1vqac3p8fl4kez7ehjz8eltugd2fm67pckpl7pn", proposal.Inclusions[0].Address)
//	require.Equal(t, uint64(2), proposal.Rejections[0].PairId)
//	require.Equal(t, "cosmos1vqac3p8fl4kez7ehjz8eltugd2fm67pckpl7pn", proposal.Rejections[0].Address)
//	require.Equal(t, uint64(1), proposal.Distributions[0].PairId)
//	require.Equal(t, "cosmos1vqac3p8fl4kez7ehjz8eltugd2fm67pckpl7pn", proposal.Distributions[0].Address)
//	require.Equal(t, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000))), proposal.Distributions[0].Amount)
//}
