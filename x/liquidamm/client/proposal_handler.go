package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	"github.com/crescent-network/crescent/v5/x/liquidamm/client/cli"
)

func dummyRESTHandler(client.Context) rest.ProposalRESTHandler {
	return rest.ProposalRESTHandler{
		SubRoute: "dummy_liquidamm",
		Handler:  func(http.ResponseWriter, *http.Request) {},
	}
}

var (
	PublicPositionCreateProposalHandler = govclient.NewProposalHandler(
		cli.NewCmdSubmitPublicPositionCreateProposal, dummyRESTHandler)
	PublicPositionParameterChangeProposalHandler = govclient.NewProposalHandler(
		cli.NewCmdSubmitPublicPositionParameterChangeProposal, dummyRESTHandler)
)
