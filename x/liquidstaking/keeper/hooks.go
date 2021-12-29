package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/tendermint/farming/x/liquidstaking/types"
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

func (h Hooks) GetOtherVotes(ctx sdk.Context, votes *govtypes.Votes, otherVotes *govtypes.OtherVotes) {
	// TODO: WIP, add types for arg
	//(*votes)["testaddress"] = make(map[govtypes.VoteOption]sdk.Dec)
	liquidVals := h.k.GetActiveLiquidValidators(ctx)
	totalSupply := h.k.bankKeeper.GetSupply(ctx, types.LiquidBondDenom).Amount
	totalVotingPower := sdk.ZeroInt()
	// TODO: btoken balance to power conversion by netAmount
	for _, vote := range *votes {
		voter, err := sdk.AccAddressFromBech32(vote.Voter)
		if err != nil {
			panic(err)
			//continue
		}
		totalVotingPower = totalVotingPower.Add(h.k.bankKeeper.GetBalance(ctx, voter., types.LiquidBondDenom).Amount)
		(*otherVotes)[vote.Voter] = map[string]map[govtypes.VoteOption]sdk.Dec{}
	}
	for voter, ovote := range *otherVotes {
		// TODO: call by ref
		for _, val := range liquidVals {
			ovote[val.OperatorAddress] = map[govtypes.VoteOption]sdk.Dec{}
			ovote[val.OperatorAddress]
		}
		//for valAddrStr, optionMap := range ovote {
		//	for option, power := range optionMap {
		//
		//	}
		//}
	}
	//(*votes)["testaddress"] = map[govtypes.VoteOption]sdk.Dec{}
	//(*votes)["testaddress"][govtypes.OptionYes] = sdk.MustNewDecFromStr("99999999.9")
	fmt.Println("[GetOtherVotes on liquid-staking", *votes)
}
