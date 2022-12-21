package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

// ParseMarketMakerProposal reads and parses a MarketMakerProposal from a file.
func ParseMarketMakerProposal(cdc codec.JSONCodec, proposalFile string) (types.MarketMakerProposal, error) {
	proposal := types.MarketMakerProposal{}

	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
