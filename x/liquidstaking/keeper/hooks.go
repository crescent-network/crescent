package keeper

import (
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

func (h Hooks) GetOtherVotes(ctx sdk.Context, votes *govtypes.Votes, otherVotes *govtypes.OtherVotes) {
	liquidVals, _ := h.k.GetActiveLiquidValidators(ctx)
	lenLiquidVals := len(liquidVals)
	liquidBondDenom := h.k.LiquidBondDenom(ctx)
	totalSupply := h.k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount.ToDec()
	if totalSupply.IsPositive() {
		for _, vote := range *votes {
			voter, err := sdk.AccAddressFromBech32(vote.Voter)
			if err != nil {
				panic(err)
				//continue
			}
			// lToken balance
			lTokenBalance := h.k.bankKeeper.GetBalance(ctx, voter, liquidBondDenom).Amount.ToDec()
			// TODO: exchange rate for native token, netAmount function
			if lTokenBalance.IsPositive() {
				(*otherVotes)[vote.Voter] = map[string]sdk.Dec{}
				dividedPower := lTokenBalance.QuoTruncate(sdk.NewDec(int64(lenLiquidVals)))
				for _, val := range liquidVals {
					if existed, ok := (*otherVotes)[vote.Voter][val.OperatorAddress]; ok {
						(*otherVotes)[vote.Voter][val.OperatorAddress] = existed.Add(dividedPower)
					} else {
						(*otherVotes)[vote.Voter][val.OperatorAddress] = dividedPower
					}
				}
			}
			// TODO: farming staking position, liquidity pool
		}
	}
	//// TODO: remove debug logging
	//for _, vote := range *votes {
	//	pp.Print("[GetOtherVotes on liquid-staking votes]", vote.Voter)
	//	for _, option := range vote.Options {
	//		pp.Println(option.Option, option.Weight.String(), option.Option)
	//	}
	//}
	//for voter, voteMap := range *otherVotes {
	//	pp.Println("[GetOtherVotes on liquid-staking otherVotes]", voter)
	//	for vali, option := range voteMap {
	//		pp.Println(vali, option.String())
	//	}
	//}
	////if totalSupply.IsPositive() && totalVotingPower.IsPositive() {
	//if totalSupply.IsPositive() {
	//	//powerRate := sdk.OneDec()
	//	//powerRate := totalVotingPower.QuoTruncate(totalSupply)
	//	//pp.Println(powerRate.String(), totalVotingPower.String(), totalSupply.String())
	//	for voter, vals := range *otherVotes {
	//		for val, power := range vals {
	//			// TODO: decimal correction
	//			//(*otherVotes)[voter][val] = power.MulTruncate(powerRate).QuoTruncate(sdk.NewDec(int64(lenLiquidVals)))
	//			(*otherVotes)[voter][val] = power.QuoTruncate(sdk.NewDec(int64(lenLiquidVals)))
	//		}
	//	}
}
