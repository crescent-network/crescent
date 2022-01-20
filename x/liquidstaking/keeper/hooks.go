package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
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

// GetOtherVotes calculate the voting power of the person who participated in liquid staking.
func (h Hooks) GetOtherVotes(ctx sdk.Context, votes *govtypes.Votes, otherVotes *govtypes.OtherVotes) {
	liquidVals := h.k.GetActiveLiquidValidators(ctx)
	liquidBondDenom := h.k.LiquidBondDenom(ctx)
	totalSupply := h.k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount
	if totalSupply.IsPositive() {
		for _, vote := range *votes {
			voter, err := sdk.AccAddressFromBech32(vote.Voter)
			if err != nil {
				panic(err)
				//continue
			}
			// bToken balance
			bTokenBalance := h.k.bankKeeper.GetBalance(ctx, voter, liquidBondDenom).Amount
			nativeValue := sdk.ZeroDec()
			// native token value = BTokenAmount * NetAmount / TotalSupply
			if bTokenBalance.IsPositive() {
				nativeValue = types.BTokenToNativeToken(bTokenBalance, totalSupply, h.k.NetAmount(ctx), sdk.ZeroDec())
			}
			if nativeValue.IsPositive() {
				(*otherVotes)[vote.Voter] = map[string]sdk.Dec{}
				// TODO: ValidateUnbondAmount, delegation shares * bonded / total shares
				// TODO: votingPower := delegation.GetShares().MulInt(val.BondedTokens).Quo(val.DelegatorShares)
				//sharesAmount, err := h.k.stakingKeeper.ValidateUnbondAmount(ctx, proxyAcc, valAddr, sharesAmount.TruncateInt())
				//if err != nil {
				//	return time.Time{}, stakingtypes.UnbondingDelegation{}, err
				//}
				dividedPowers, _ := types.DivideByCurrentWeightDec(liquidVals, nativeValue)
				for i, val := range liquidVals {
					if existed, ok := (*otherVotes)[vote.Voter][val.OperatorAddress]; ok {
						(*otherVotes)[vote.Voter][val.OperatorAddress] = existed.Add(dividedPowers[i])
					} else {
						(*otherVotes)[vote.Voter][val.OperatorAddress] = dividedPowers[i]
					}
				}
			}
			// TODO: farming staking position, liquidity pool
		}
	}
}
