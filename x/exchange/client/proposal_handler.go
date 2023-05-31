package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	"github.com/crescent-network/crescent/v5/x/exchange/client/cli"
)

func dummyRESTHandler(client.Context) rest.ProposalRESTHandler {
	return rest.ProposalRESTHandler{
		SubRoute: "dummy_exchange",
		Handler:  func(http.ResponseWriter, *http.Request) {},
	}
}

var (
	MarketParameterChangeProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitMarketParameterChangeProposal, dummyRESTHandler)
)
