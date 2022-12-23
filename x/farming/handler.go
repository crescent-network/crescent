package farming

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/crescent-network/crescent/v4/x/farming/keeper"
	"github.com/crescent-network/crescent/v4/x/farming/types"
)

func NewHandler(_ keeper.Keeper) sdk.Handler {
	return func(_ sdk.Context, _ sdk.Msg) (*sdk.Result, error) {
		return nil, types.ErrModuleDisabled
	}
}

// NewPublicPlanProposalHandler creates a governance handler to manage new proposal types.
// It enables PublicPlanProposal to propose a plan creation / modification / deletion.
func NewPublicPlanProposalHandler(_ keeper.Keeper) govtypes.Handler {
	return func(_ sdk.Context, _ govtypes.Content) error {
		return types.ErrModuleDisabled
	}
}
