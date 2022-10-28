package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

func ProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "farming_plan",
		Handler:  postProposalHandlerFn(clientCtx),
	}
}

func postProposalHandlerFn(_ client.Context) http.HandlerFunc {
	return func(_ http.ResponseWriter, _ *http.Request) {
	}
}
