package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
	"github.com/cosmosquad-labs/squad/x/utils"
)

// GetVoterBalanceByDenom return map of balance amount of voter by denom
func (k Keeper) GetVoterBalanceByDenom(ctx sdk.Context, votes *govtypes.Votes) map[string]map[string]sdk.Int {
	denomAddrBalanceMap := make(map[string]map[string]sdk.Int)
	for _, vote := range *votes {
		voter, err := sdk.AccAddressFromBech32(vote.Voter)
		if err != nil {
			continue
		}
		balances := k.bankKeeper.GetAllBalances(ctx, voter)
		for _, coin := range balances {
			if _, ok := denomAddrBalanceMap[coin.Denom]; !ok {
				denomAddrBalanceMap[coin.Denom] = map[string]sdk.Int{}
			}
			if coin.Amount.IsPositive() {
				denomAddrBalanceMap[coin.Denom][vote.Voter] = coin.Amount
			}
		}
	}
	utils.PP(denomAddrBalanceMap)
	return denomAddrBalanceMap
}

func (k Keeper) LiquidGov(ctx sdk.Context, votes *govtypes.Votes, otherVotes *govtypes.OtherVotes) {
	// TODO: active or with delisting
	liquidVals := k.GetActiveLiquidValidators(ctx)
	liquidBondDenom := k.LiquidBondDenom(ctx)
	totalSupply := k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount
	bTokenValueMap := make(map[string]sdk.Int)
	// get the map of balance amount of voter by denom
	voterBalanceByDenom := k.GetVoterBalanceByDenom(ctx, votes)
	if !totalSupply.IsPositive() {
		return
	}
	// calculate btoken value of each voter
	for denom, balanceByVoter := range voterBalanceByDenom {
		// add balance of bToken value
		if denom == liquidBondDenom {
			for voter, balance := range balanceByVoter {
				if _, ok := bTokenValueMap[voter]; !ok {
					bTokenValueMap[voter] = balance
				} else {
					bTokenValueMap[voter] = bTokenValueMap[voter].Add(balance)
				}
			}
		}
		// add balance of PoolTokens including bToken value

		// add Farming Staking Position of PoolTokens including bToken

		// add Farming Queued Staking of PoolTokens including bToken

		// add Farming Staking Position of bToken

		// add Farming Queued Staking of bToken
	}

	for voter, btokenValue := range bTokenValueMap {
		nativeValue := sdk.ZeroDec()
		if btokenValue.IsPositive() {
			nativeValue = types.BTokenToNativeToken(btokenValue, totalSupply, k.NetAmount(ctx), sdk.ZeroDec())
		}
		if nativeValue.IsPositive() {
			(*otherVotes)[voter] = map[string]sdk.Dec{}
			// TODO: ValidateUnbondAmount, delegation shares * bonded / total shares
			// TODO: votingPower := delegation.GetShares().MulInt(val.BondedTokens).Quo(val.DelegatorShares)
			//sharesAmount, err := k.stakingKeeper.ValidateUnbondAmount(ctx, proxyAcc, valAddr, sharesAmount.TruncateInt())
			//if err != nil {
			//	return time.Time{}, stakingtypes.UnbondingDelegation{}, err
			//}
			dividedPowers, _ := types.DivideByCurrentWeightDec(liquidVals, nativeValue)
			for i, val := range liquidVals {
				if existed, ok := (*otherVotes)[voter][val.OperatorAddress]; ok {
					(*otherVotes)[voter][val.OperatorAddress] = existed.Add(dividedPowers[i])
				} else {
					(*otherVotes)[voter][val.OperatorAddress] = dividedPowers[i]
				}
			}
		}
	}
}
