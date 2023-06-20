package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/client/cli"
)

func dummyRESTHandler(client.Context) rest.ProposalRESTHandler {
	return rest.ProposalRESTHandler{
		SubRoute: "dummy_liquidfarming",
		Handler:  func(http.ResponseWriter, *http.Request) {},
	}
}

var (
	LiquidFarmCreateProposalHandler          = govclient.NewProposalHandler(cli.NewCmdSubmitLiquidFarmCreateProposal, dummyRESTHandler)
	LiquidFarmParameterChangeProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitLiquidFarmParameterChangeProposal, dummyRESTHandler)
)
