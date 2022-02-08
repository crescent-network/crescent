package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
	squadtypes.PP(denomAddrBalanceMap)
	return denomAddrBalanceMap
}

func (k Keeper) TallyLiquidGov(ctx sdk.Context, votes *govtypes.Votes, otherVotes *govtypes.OtherVotes) {
	params := k.GetParams(ctx)
	activeVals := k.GetActiveLiquidValidators(ctx, params.WhitelistedValMap())
	bondedBondDenom := k.BondedBondDenom(ctx)
	totalSupply := k.bankKeeper.GetSupply(ctx, bondedBondDenom).Amount
	bTokenValueMap := make(squadtypes.StrIntMap)
	// get the map of balance amount of voter by denom
	voterBalanceByDenom := k.GetVoterBalanceByDenom(ctx, votes)
	bTokenSharePerPoolCoinMap := make(map[string]sdk.Dec)
	if !totalSupply.IsPositive() {
		return
	}
	// calculate btoken value of each voter
	for denom, balanceByVoter := range voterBalanceByDenom {

		// add balance of bToken value
		if denom == bondedBondDenom {
			for voter, balance := range balanceByVoter {
				bTokenValueMap.AddOrSet(voter, balance)
			}
		}

		// add balance of PoolTokens including bToken value
		if pool, found := k.liquidityKeeper.GetPool(ctx, liquiditytypes.ParsePoolCoinDenom(denom)); found {
			if pair, found := k.liquidityKeeper.GetPair(ctx, pool.PairId); found {
				rx, ry := k.liquidityKeeper.GetPoolBalance(ctx, pool, pair)
				poolCoinSupply := k.liquidityKeeper.GetPoolCoinSupply(ctx, pool)
				bTokenSharePerPoolCoin := sdk.ZeroDec()
				if pair.QuoteCoinDenom == bondedBondDenom {
					bTokenSharePerPoolCoin = rx.ToDec().Quo(poolCoinSupply.ToDec())
				}
				if pair.BaseCoinDenom == bondedBondDenom {
					bTokenSharePerPoolCoin = ry.ToDec().Quo(poolCoinSupply.ToDec())
				}
				if !bTokenSharePerPoolCoin.IsPositive() {
					continue
				}
				bTokenSharePerPoolCoinMap[denom] = bTokenSharePerPoolCoin
				for voter, balance := range balanceByVoter {
					bTokenValueMap.AddOrSet(voter, squadtypes.GetShareValue(balance, bTokenSharePerPoolCoin))
				}
			}
		}
	}

	for _, vote := range *votes {
		voter, err := sdk.AccAddressFromBech32(vote.Voter)
		if err != nil {
			continue
		}
		// add value of Farming Staking Position of bToken and PoolTokens including bToken
		k.farmingKeeper.IterateStakingsByFarmer(ctx, voter, func(stakingCoinDenom string, staking farmingtypes.Staking) (stop bool) {
			if stakingCoinDenom == bondedBondDenom {
				bTokenValueMap.AddOrSet(vote.Voter, staking.Amount)
			} else if ratio, ok := bTokenSharePerPoolCoinMap[stakingCoinDenom]; ok {
				bTokenValueMap.AddOrSet(vote.Voter, squadtypes.GetShareValue(staking.Amount, ratio))
			}
			return false
		})

		// add value of Farming Queued Staking of bToken and PoolTokens including bToken
		k.farmingKeeper.IterateQueuedStakingsByFarmer(ctx, voter, func(stakingCoinDenom string, queuedStaking farmingtypes.QueuedStaking) (stop bool) {
			if stakingCoinDenom == bondedBondDenom {
				bTokenValueMap.AddOrSet(vote.Voter, queuedStaking.Amount)
			} else if ratio, ok := bTokenSharePerPoolCoinMap[stakingCoinDenom]; ok {
				bTokenValueMap.AddOrSet(vote.Voter, squadtypes.GetShareValue(queuedStaking.Amount, ratio))
			}
			return false
		})
	}

	for voter, bTokenValue := range bTokenValueMap {
		delShares := sdk.ZeroDec()
		if bTokenValue.IsPositive() {
			delShares = types.BTokenToNativeToken(bTokenValue, totalSupply, k.NetAmount(ctx), sdk.ZeroDec())
		}
		if delShares.IsPositive() {
			(*otherVotes)[voter] = map[string]sdk.Dec{}
			dividedPowers, _ := k.DivideByCurrentWeight(ctx, activeVals, delShares)
			for i, val := range activeVals {
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
