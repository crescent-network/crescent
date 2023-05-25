package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/crescent-network/crescent/v5/x/amm/client/cli"
)

var (
	PoolParameterChangeProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitPoolParameterChangeProposal, nil)
	PublicFarmingPlanProposalHandler   = govclient.NewProposalHandler(cli.NewCmdSubmitPublicFarmingPlanProposal, nil)
)
