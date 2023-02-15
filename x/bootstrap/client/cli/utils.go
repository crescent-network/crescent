package cli

import (
	"fmt"
	"os"
	"strings"

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

// parseOrderDirection parses order direction string and returns
// types.OrderDirection.
func parseOrderDirection(s string) (types.OrderDirection, error) {
	switch strings.ToLower(s) {
	case "buy", "b":
		return types.OrderDirectionBuy, nil
	case "sell", "s":
		return types.OrderDirectionSell, nil
	}
	return 0, fmt.Errorf("invalid order direction: %s", s)
}
