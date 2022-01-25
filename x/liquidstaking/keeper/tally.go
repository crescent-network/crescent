package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	farmingtypes "github.com/cosmosquad-labs/squad/x/farming/types"
	liquiditytypes "github.com/cosmosquad-labs/squad/x/liquidity/types"
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
	bTokenValueMap := make(utils.StrIntMap)
	// get the map of balance amount of voter by denom
	voterBalanceByDenom := k.GetVoterBalanceByDenom(ctx, votes)
	bTokenSharePerPoolCoinMap := make(map[string]sdk.Dec)
	//bTokenSharePerPoolCoinMap[liquidBondDenom] = sdk.OneDec()
	if !totalSupply.IsPositive() {
		return
	}
	// calculate btoken value of each voter
	for denom, balanceByVoter := range voterBalanceByDenom {

		// add balance of bToken value
		if denom == liquidBondDenom {
			for voter, balance := range balanceByVoter {
				bTokenValueMap.AddOrInit(voter, balance)
			}
		}

		// add balance of PoolTokens including bToken value
		if pool, found := k.liquidityKeeper.GetPool(ctx, liquiditytypes.ParsePoolCoinDenom(denom)); !found {
			if pair, found := k.liquidityKeeper.GetPair(ctx, pool.PairId); found {
				rx, ry := k.liquidityKeeper.GetPoolBalance(ctx, pool, pair)
				poolCoinSupply := k.liquidityKeeper.GetPoolCoinSupply(ctx, pool)
				bTokenSharePerPoolCoin := sdk.ZeroDec()
				if pair.QuoteCoinDenom == liquidBondDenom {
					bTokenSharePerPoolCoin = rx.ToDec().Quo(poolCoinSupply.ToDec())
				}
				if pair.BaseCoinDenom == liquidBondDenom {
					bTokenSharePerPoolCoin = ry.ToDec().Quo(poolCoinSupply.ToDec())
				}
				if !bTokenSharePerPoolCoin.IsPositive() {
					continue
				}
				bTokenSharePerPoolCoinMap[denom] = bTokenSharePerPoolCoin
				for voter, balance := range balanceByVoter {
					bTokenValueMap.AddOrInit(voter, utils.GetShareValue(balance, bTokenSharePerPoolCoin))
				}
			}
		}
	}

	for _, vote := range *votes {
		voter, err := sdk.AccAddressFromBech32(vote.Voter)
		if err != nil {
			continue
		}
		// add Farming Staking Position of bToken and PoolTokens including bToken
		k.farmingKeeper.IterateStakingsByFarmer(ctx, voter, func(stakingCoinDenom string, staking farmingtypes.Staking) (stop bool) {
			if stakingCoinDenom == liquidBondDenom {
				bTokenValueMap.AddOrInit(vote.Voter, staking.Amount)
			} else if ratio, ok := bTokenSharePerPoolCoinMap[stakingCoinDenom]; ok {
				bTokenValueMap.AddOrInit(vote.Voter, utils.GetShareValue(staking.Amount, ratio))
			}
			return false
		})

		// add Farming Queued Staking of bToken and PoolTokens including bToken
		k.farmingKeeper.IterateQueuedStakingsByFarmer(ctx, voter, func(stakingCoinDenom string, queuedStaking farmingtypes.QueuedStaking) (stop bool) {
			if stakingCoinDenom == liquidBondDenom {
				bTokenValueMap.AddOrInit(vote.Voter, queuedStaking.Amount)
			} else if ratio, ok := bTokenSharePerPoolCoinMap[stakingCoinDenom]; ok {
				bTokenValueMap.AddOrInit(vote.Voter, utils.GetShareValue(queuedStaking.Amount, ratio))
			}
			return false
		})
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
