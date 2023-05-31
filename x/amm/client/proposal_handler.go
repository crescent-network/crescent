package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	"github.com/crescent-network/crescent/v5/x/amm/client/cli"
)

func dummyRESTHandler(client.Context) rest.ProposalRESTHandler {
	return rest.ProposalRESTHandler{
		SubRoute: "dummy_amm",
		Handler:  func(http.ResponseWriter, *http.Request) {},
	}
}

var (
	PoolParameterChangeProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitPoolParameterChangeProposal, dummyRESTHandler)
	PublicFarmingPlanProposalHandler   = govclient.NewProposalHandler(cli.NewCmdSubmitPublicFarmingPlanProposal, dummyRESTHandler)
)
