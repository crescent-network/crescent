package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/crescent-network/crescent/v5/x/exchange/client/cli"
)

var (
	MarketParameterChangeProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitMarketParameterChangeProposal, nil)
)
