package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/k0kubun/pp"
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
	liquidVals := h.k.GetActiveLiquidValidators(ctx)
	lenLiquidVals := len(liquidVals)
	totalSupply := h.k.bankKeeper.GetSupply(ctx, types.LiquidBondDenom).Amount.ToDec()
	totalVotingPower := sdk.ZeroDec()
	for _, vote := range *votes {
		voter, err := sdk.AccAddressFromBech32(vote.Voter)
		if err != nil {
			panic(err)
			//continue
		}
		lTokenBalance := h.k.bankKeeper.GetBalance(ctx, voter, types.LiquidBondDenom).Amount.ToDec()
		totalVotingPower = totalVotingPower.Add(lTokenBalance)
		(*otherVotes)[vote.Voter] = map[string]sdk.Dec{}
		for _, val := range liquidVals {
			(*otherVotes)[vote.Voter][val.OperatorAddress] = lTokenBalance
		}
	}
	if totalSupply.IsPositive() && totalVotingPower.IsPositive() {
		powerRate := totalVotingPower.QuoTruncate(totalSupply)
		pp.Println(powerRate.String(), totalVotingPower.String(), totalSupply.String())
		for voter, vals := range *otherVotes {
			for val, power := range vals {
				// TODO: decimal correction
				(*otherVotes)[voter][val] = power.MulTruncate(powerRate).QuoTruncate(sdk.NewDec(int64(lenLiquidVals)))
			}
		}
		pp.Print("[GetOtherVotes on liquid-staking votes", *votes)
		pp.Print("[GetOtherVotes on liquid-staking otherVotes", *otherVotes)
	}
}
