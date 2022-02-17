package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	squadtypes "github.com/cosmosquad-labs/squad/types"
	farmingtypes "github.com/cosmosquad-labs/squad/x/farming/types"
	liquiditytypes "github.com/cosmosquad-labs/squad/x/liquidity/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
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
	return denomAddrBalanceMap
}

// TokenValueFromFarmingPositions returns TokenValue of exist farming positions including queued of the addr
func (k Keeper) TokenValueFromFarmingPositions(ctx sdk.Context, addr sdk.AccAddress, targetDenom string, tokenSharePerPoolCoinMap map[string]sdk.Dec) sdk.Int {
	tokenValue := sdk.ZeroInt()

	// add value of Farming Staking Position of bToken and PoolTokens including bToken
	k.farmingKeeper.IterateStakingsByFarmer(ctx, addr, func(stakingCoinDenom string, staking farmingtypes.Staking) (stop bool) {
		if stakingCoinDenom == targetDenom {
			tokenValue = tokenValue.Add(staking.Amount)
		} else if ratio, ok := tokenSharePerPoolCoinMap[stakingCoinDenom]; ok {
			tokenValue = tokenValue.Add(squadtypes.GetShareValue(staking.Amount, ratio))
		}
		return false
	})

	// add value of Farming Queued Staking of bToken and PoolTokens including bToken
	k.farmingKeeper.IterateQueuedStakingsByFarmer(ctx, addr, func(stakingCoinDenom string, queuedStaking farmingtypes.QueuedStaking) (stop bool) {
		if stakingCoinDenom == targetDenom {
			tokenValue = tokenValue.Add(queuedStaking.Amount)
		} else if ratio, ok := tokenSharePerPoolCoinMap[stakingCoinDenom]; ok {
			tokenValue = tokenValue.Add(squadtypes.GetShareValue(queuedStaking.Amount, ratio))
		}
		return false
	})

	return tokenValue
}

// TokenSharePerPoolCoin returns token share of the target denom of a pool coin
func (k Keeper) TokenSharePerPoolCoin(ctx sdk.Context, targetDenom, poolCoinDenom string) sdk.Dec {
	pool, found := k.liquidityKeeper.GetPool(ctx, liquiditytypes.ParsePoolCoinDenom(poolCoinDenom))
	if !found {
		return sdk.ZeroDec()
	}

	pair, found := k.liquidityKeeper.GetPair(ctx, pool.PairId)
	if !found {
		return sdk.ZeroDec()
	}

	rx, ry := k.liquidityKeeper.GetPoolBalances(ctx, pool)
	poolCoinSupply := k.liquidityKeeper.GetPoolCoinSupply(ctx, pool)
	if !poolCoinSupply.IsPositive() {
		return sdk.ZeroDec()
	}
	bTokenSharePerPoolCoin := sdk.ZeroDec()
	if pair.QuoteCoinDenom == targetDenom {
		bTokenSharePerPoolCoin = rx.Amount.ToDec().QuoTruncate(poolCoinSupply.ToDec())
	} else if pair.BaseCoinDenom == targetDenom {
		bTokenSharePerPoolCoin = ry.Amount.ToDec().QuoTruncate(poolCoinSupply.ToDec())
	}
	if !bTokenSharePerPoolCoin.IsPositive() {
		return sdk.ZeroDec()
	}
	return bTokenSharePerPoolCoin
}

// CalcVotingPower returns voting power of the addr by normal delegations
func (k Keeper) CalcVotingPower(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	totalVotingPower := sdk.ZeroInt()
	k.stakingKeeper.IterateDelegations(
		ctx, addr,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(ctx, valAddr)
			delShares := del.GetShares()
			// if the validator not bonded, bonded token and voting power is zero
			if delShares.IsPositive() && val.IsBonded() {
				votingPower := val.TokensFromSharesTruncated(delShares).TruncateInt()
				if votingPower.IsPositive() {
					totalVotingPower = totalVotingPower.Add(votingPower)
				}
			}
			return false
		},
	)
	return totalVotingPower
}

// CalcLiquidVotingPower returns voting power of the addr by liquid bond denom
// TODO: refactor votingPowerStruct (delShares, btoken, poolCoin, farming)
func (k Keeper) CalcLiquidVotingPower(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	liquidBondDenom := k.LiquidBondDenom(ctx)

	// skip when no liquid bond token supply
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount
	if !bTokenTotalSupply.IsPositive() {
		return sdk.ZeroInt()
	}

	// skip when no active validators, liquid tokens
	liquidVals := k.GetAllLiquidValidators(ctx)
	if len(liquidVals) == 0 {
		return sdk.ZeroInt()
	}

	// using only liquid tokens of bonded liquid validators to ensure voting power doesn't exceed delegation shares on x/gov tally
	totalBondedLiquidTokens, _ := liquidVals.TotalLiquidTokens(ctx, k.stakingKeeper, true)
	if !totalBondedLiquidTokens.IsPositive() {
		return sdk.ZeroInt()
	}

	bTokenValue := sdk.ZeroInt()
	bTokenSharePerPoolCoinMap := make(map[string]sdk.Dec)
	balances := k.bankKeeper.GetAllBalances(ctx, addr)
	for _, coin := range balances {
		// add balance of bToken value
		if coin.Denom == liquidBondDenom {
			bTokenValue = bTokenValue.Add(coin.Amount)
		}

		// check if the denom is pool coin
		bTokenSharePerPoolCoin := k.TokenSharePerPoolCoin(ctx, liquidBondDenom, coin.Denom)
		if bTokenSharePerPoolCoin.IsPositive() {
			bTokenSharePerPoolCoinMap[coin.Denom] = bTokenSharePerPoolCoin
			bTokenValue = bTokenValue.Add(squadtypes.GetShareValue(coin.Amount, bTokenSharePerPoolCoin))
		}
	}

	tokenValue := k.TokenValueFromFarmingPositions(ctx, addr, liquidBondDenom, bTokenSharePerPoolCoinMap)
	if tokenValue.IsPositive() {
		bTokenValue = bTokenValue.Add(tokenValue)
	}

	if bTokenValue.IsPositive() {
		return types.BTokenToNativeToken(bTokenValue, bTokenTotalSupply, totalBondedLiquidTokens.ToDec()).TruncateInt()
	} else {
		return sdk.ZeroInt()
	}
}

func (k Keeper) TallyLiquidGov(ctx sdk.Context, votes *govtypes.Votes, otherVotes *govtypes.OtherVotes) {
	liquidBondDenom := k.LiquidBondDenom(ctx)

	// skip when no liquid bond token supply
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount
	if !bTokenTotalSupply.IsPositive() {
		return
	}

	// skip when no active validators, liquid tokens
	liquidVals := k.GetAllLiquidValidators(ctx)
	if len(liquidVals) == 0 {
		return
	}
	// using only liquid tokens of bonded liquid validators to ensure voting power doesn't exceed delegation shares on x/gov tally
	totalBondedLiquidTokens, bondedLiquidTokenMap := liquidVals.TotalLiquidTokens(ctx, k.stakingKeeper, true)
	if !totalBondedLiquidTokens.IsPositive() {
		return
	}

	// get the map of balance amount of voter by denom
	voterBalanceByDenom := k.GetVoterBalanceByDenom(ctx, votes)
	bTokenSharePerPoolCoinMap := make(map[string]sdk.Dec)
	bTokenValueMap := make(squadtypes.StrIntMap)

	// calculate btoken value of each voter
	for denom, balanceByVoter := range voterBalanceByDenom {

		// add balance of bToken value
		if denom == liquidBondDenom {
			for voter, balance := range balanceByVoter {
				bTokenValueMap.AddOrSet(voter, balance)
			}
		}

		// if the denom is pool coin, calc btoken share and add btoken value on bTokenValueMap
		bTokenSharePerPoolCoin := k.TokenSharePerPoolCoin(ctx, liquidBondDenom, denom)
		if bTokenSharePerPoolCoin.IsPositive() {
			bTokenSharePerPoolCoinMap[denom] = bTokenSharePerPoolCoin
			for voter, balance := range balanceByVoter {
				bTokenValueMap.AddOrSet(voter, squadtypes.GetShareValue(balance, bTokenSharePerPoolCoin))
			}
		}
	}

	// add btoken value of farming positions on bTokenValueMap
	for _, vote := range *votes {
		voter, err := sdk.AccAddressFromBech32(vote.Voter)
		if err != nil {
			continue
		}
		tokenValue := k.TokenValueFromFarmingPositions(ctx, voter, liquidBondDenom, bTokenSharePerPoolCoinMap)
		if tokenValue.IsPositive() {
			bTokenValueMap.AddOrSet(vote.Voter, tokenValue)
		}
	}

	for voter, bTokenValue := range bTokenValueMap {
		// caclulate voting power of the voter, distribute voting power to liquid validators by current weight
		votingPower := sdk.ZeroDec()
		if bTokenValue.IsPositive() {
			votingPower = types.BTokenToNativeToken(bTokenValue, bTokenTotalSupply, totalBondedLiquidTokens.ToDec())
		}
		if votingPower.IsPositive() {
			(*otherVotes)[voter] = map[string]sdk.Dec{}
			// drop crumb for defensive policy about delShares decimal errors
			dividedPowers, _ := types.DivideByCurrentWeight((types.ActiveLiquidValidators)(liquidVals), votingPower, totalBondedLiquidTokens, bondedLiquidTokenMap)
			if len(dividedPowers) == 0 {
				continue
			}
			for i, val := range liquidVals {
				if !dividedPowers[i].IsPositive() {
					continue
				}
				if existed, ok := (*otherVotes)[voter][val.OperatorAddress]; ok {
					(*otherVotes)[voter][val.OperatorAddress] = existed.Add(dividedPowers[i])
				} else {
					(*otherVotes)[voter][val.OperatorAddress] = dividedPowers[i]
				}
			}
		}
	}
	// TODO: consider return bTokenValueMap, bTokenSharePerPoolCoinMap for assertion and query
}
