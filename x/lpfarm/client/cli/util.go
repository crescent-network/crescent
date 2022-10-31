package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/crescent-network/crescent/v3/x/lpfarm/types"
)

func ParseFarmingPlanProposal(cdc codec.JSONCodec, proposalFile string) (types.FarmingPlanProposal, error) {
	proposal := types.FarmingPlanProposal{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
