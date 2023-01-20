package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

// ParseBootstrapProposal reads and parses a BootstrapProposal from a file.
func ParseBootstrapProposal(cdc codec.JSONCodec, proposalFile string) (types.BootstrapProposal, error) {
	proposal := types.BootstrapProposal{}

	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
