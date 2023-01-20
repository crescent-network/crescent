package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/crescent-network/crescent/v4/x/bootstrap/client/cli"
	"github.com/crescent-network/crescent/v4/x/bootstrap/client/rest"
)

// ProposalHandler is the market maker proposal command handler.
// Note that rest.ProposalRESTHandler will be deprecated in the future.
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitBootstrapProposal, rest.ProposalRESTHandler)
)
