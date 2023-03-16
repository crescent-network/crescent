package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/crescent-network/crescent/v5/x/marketmaker/client/cli"
	"github.com/crescent-network/crescent/v5/x/marketmaker/client/rest"
)

// ProposalHandler is the market maker proposal command handler.
// Note that rest.ProposalRESTHandler will be deprecated in the future.
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitMarketMakerProposal, rest.ProposalRESTHandler)
)
