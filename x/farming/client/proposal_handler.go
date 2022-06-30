package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/crescent-network/crescent/v2/x/farming/client/cli"
	"github.com/crescent-network/crescent/v2/x/farming/client/rest"
)

// ProposalHandler is the public plan command handler.
// Note that rest.ProposalRESTHandler will be deprecated in the future.
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitPublicPlanProposal, rest.ProposalRESTHandler)
)
