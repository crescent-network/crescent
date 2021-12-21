package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ govtypes.GovHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

func (h Hooks) AfterProposalSubmission(_ sdk.Context, _ uint64)                {}
func (h Hooks) AfterProposalDeposit(_ sdk.Context, _ uint64, _ sdk.AccAddress) {}
func (h Hooks) AfterProposalVote(_ sdk.Context, _ uint64, _ sdk.AccAddress)    {}
func (h Hooks) AfterProposalFailedMinDeposit(_ sdk.Context, _ uint64)          {}
func (h Hooks) AfterProposalVotingPeriodEnded(_ sdk.Context, _ uint64)         {}

func (h Hooks) GetOtherVotes(ctx sdk.Context, votes *govtypes.OtherVotes) {
	// TODO: WIP
	//(*votes)["testaddress"] = make(map[govtypes.VoteOption]sdk.Dec)
	(*votes)["testaddress"] = map[govtypes.VoteOption]sdk.Dec{}
	(*votes)["testaddress"][govtypes.OptionYes] = sdk.MustNewDecFromStr("99999999.9")
	fmt.Println("[GetOtherVotes on liquid-staking", *votes)
}
