package keeper

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	utils "github.com/crescent-network/crescent/v2/types"
	farmingtypes "github.com/crescent-network/crescent/v2/x/farming/types"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// GetVoterBalanceByDenom return map of balance amount of voter by denom
func (k Keeper) GetVoterBalanceByDenom(ctx sdk.Context, votes govtypes.Votes) map[string]map[string]sdk.Int {
	denomAddrBalanceMap := map[string]map[string]sdk.Int{}
	for _, vote := range votes {
		voter, err := sdk.AccAddressFromBech32(vote.Voter)
		if err != nil {
			continue
		}
		balances := k.bankKeeper.SpendableCoins(ctx, voter)
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

// GetBTokenSharePerPoolCoinMap creates bTokenSharePerPoolCoinMap of pool coins which is containing target denom to calculate target denom value of farming positions
func (k Keeper) GetBTokenSharePerPoolCoinMap(ctx sdk.Context, targetDenom string) map[string]sdk.Dec {
	bTokenSharePerPoolCoinMap := map[string]sdk.Dec{}
	_ = k.liquidityKeeper.IterateAllPools(ctx, func(pool liquiditytypes.Pool) (stop bool, err error) {
		bTokenSharePerPoolCoin := k.TokenSharePerPoolCoin(ctx, targetDenom, pool.PoolCoinDenom)
		if bTokenSharePerPoolCoin.IsPositive() {
			bTokenSharePerPoolCoinMap[pool.PoolCoinDenom] = bTokenSharePerPoolCoin
		}
		return false, nil
	})
	return bTokenSharePerPoolCoinMap
}

// TokenAmountFromFarmingPositions returns worth of staked tokens amount of exist farming positions including queued of the addr
func (k Keeper) TokenAmountFromFarmingPositions(ctx sdk.Context, addr sdk.AccAddress, targetDenom string, tokenSharePerPoolCoinMap map[string]sdk.Dec) sdk.Int {
	tokenAmount := sdk.ZeroInt()

	// add worth of staked amount of Farming Staking Position of bToken and PoolTokens including bToken
	k.farmingKeeper.IterateStakingsByFarmer(ctx, addr, func(stakingCoinDenom string, staking farmingtypes.Staking) (stop bool) {
		if stakingCoinDenom == targetDenom {
			tokenAmount = tokenAmount.Add(staking.Amount)
		} else if ratio, ok := tokenSharePerPoolCoinMap[stakingCoinDenom]; ok {
			tokenAmount = tokenAmount.Add(utils.GetShareValue(staking.Amount, ratio))
		}
		return false
	})

	// add worth of staked amount of Farming Queued Staking of bToken and PoolTokens including bToken
	k.farmingKeeper.IterateQueuedStakingsByFarmer(ctx, addr, func(stakingCoinDenom string, endTime time.Time, queuedStaking farmingtypes.QueuedStaking) (stop bool) {
		if !endTime.After(ctx.BlockTime()) { // sanity check
			return false
		}
		if stakingCoinDenom == targetDenom {
			tokenAmount = tokenAmount.Add(queuedStaking.Amount)
		} else if ratio, ok := tokenSharePerPoolCoinMap[stakingCoinDenom]; ok {
			tokenAmount = tokenAmount.Add(utils.GetShareValue(queuedStaking.Amount, ratio))
		}
		return false
	})

	return tokenAmount
}

// TokenSharePerPoolCoin returns token share of the target denom of a pool coin
func (k Keeper) TokenSharePerPoolCoin(ctx sdk.Context, targetDenom, poolCoinDenom string) sdk.Dec {
	poolId, err := liquiditytypes.ParsePoolCoinDenom(poolCoinDenom)
	if err != nil {
		// If poolCoinDenom is not a valid pool coin denom, just return zero.
		return sdk.ZeroDec()
	}
	pool, found := k.liquidityKeeper.GetPool(ctx, poolId)
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

func (k Keeper) GetVotingPower(ctx sdk.Context, addr sdk.AccAddress) types.VotingPower {
	val, found := k.stakingKeeper.GetValidator(ctx, addr.Bytes())
	validatorVotingPower := sdk.ZeroInt()
	if found {
		validatorVotingPower = val.BondedTokens()
	}
	return types.VotingPower{
		Voter:                    addr.String(),
		StakingVotingPower:       k.CalcStakingVotingPower(ctx, addr),
		LiquidStakingVotingPower: k.CalcLiquidStakingVotingPower(ctx, addr),
		ValidatorVotingPower:     validatorVotingPower,
	}
}

// CalcStakingVotingPower returns voting power of the addr by normal delegations except self-delegation
func (k Keeper) CalcStakingVotingPower(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	totalVotingPower := sdk.ZeroInt()
	k.stakingKeeper.IterateDelegations(
		ctx, addr,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(ctx, valAddr)
			delShares := del.GetShares()
			// if the validator not bonded, bonded token and voting power is zero, and except self-delegation power
			if delShares.IsPositive() && val.IsBonded() && !valAddr.Equals(addr) {
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

// CalcLiquidStakingVotingPower returns voting power of the addr by liquid bond denom
func (k Keeper) CalcLiquidStakingVotingPower(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
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

	bTokenAmount := sdk.ZeroInt()
	bTokenSharePerPoolCoinMap := k.GetBTokenSharePerPoolCoinMap(ctx, liquidBondDenom)

	balances := k.bankKeeper.SpendableCoins(ctx, addr)
	for _, coin := range balances {
		// add balance of bToken
		if coin.Denom == liquidBondDenom {
			bTokenAmount = bTokenAmount.Add(coin.Amount)
		}

		// check if the denom is pool coin
		if bTokenSharePerPoolCoin, ok := bTokenSharePerPoolCoinMap[coin.Denom]; ok && bTokenSharePerPoolCoin.IsPositive() {
			bTokenAmount = bTokenAmount.Add(utils.GetShareValue(coin.Amount, bTokenSharePerPoolCoin))
		}
	}

	tokenAmount := k.TokenAmountFromFarmingPositions(ctx, addr, liquidBondDenom, bTokenSharePerPoolCoinMap)
	if tokenAmount.IsPositive() {
		bTokenAmount = bTokenAmount.Add(tokenAmount)
	}

	if bTokenAmount.IsPositive() {
		return types.BTokenToNativeToken(bTokenAmount, bTokenTotalSupply, totalBondedLiquidTokens.ToDec()).TruncateInt()
	} else {
		return sdk.ZeroInt()
	}
}

func (k Keeper) SetLiquidStakingVotingPowers(ctx sdk.Context, votes govtypes.Votes, votingPowers *govtypes.AdditionalVotingPowers) {
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
	bTokenSharePerPoolCoinMap := map[string]sdk.Dec{}
	bTokenOwnMap := make(utils.StrIntMap)

	// sort denom keys of voterBalanceByDenom for deterministic iteration
	var denoms []string
	for denom := range voterBalanceByDenom {
		denoms = append(denoms, denom)
	}
	sort.Strings(denoms)

	// calculate owned btoken amount of each voter
	for _, denom := range denoms {

		// add balance of bToken
		if denom == liquidBondDenom {
			for voter, balance := range voterBalanceByDenom[denom] {
				bTokenOwnMap.AddOrSet(voter, balance)
			}
			continue
		}

		// if the denom is pool coin, calc btoken share and add owned btoken on bTokenOwnMap
		bTokenSharePerPoolCoin := k.TokenSharePerPoolCoin(ctx, liquidBondDenom, denom)
		if bTokenSharePerPoolCoin.IsPositive() {
			bTokenSharePerPoolCoinMap[denom] = bTokenSharePerPoolCoin
			for voter, balance := range voterBalanceByDenom[denom] {
				bTokenOwnMap.AddOrSet(voter, utils.GetShareValue(balance, bTokenSharePerPoolCoin))
			}
		}
	}

	// add owned btoken amount of farming positions on bTokenOwnMap
	for _, vote := range votes {
		voter, err := sdk.AccAddressFromBech32(vote.Voter)
		if err != nil {
			continue
		}
		tokenAmount := k.TokenAmountFromFarmingPositions(ctx, voter, liquidBondDenom, bTokenSharePerPoolCoinMap)
		if tokenAmount.IsPositive() {
			bTokenOwnMap.AddOrSet(vote.Voter, tokenAmount)
		}
	}

	for voter, bTokenAmount := range bTokenOwnMap {
		// calculate voting power of the voter, distribute voting power to liquid validators by current weight of bonded liquid tokens
		votingPower := sdk.ZeroDec()
		if bTokenAmount.IsPositive() {
			votingPower = types.BTokenToNativeToken(bTokenAmount, bTokenTotalSupply, totalBondedLiquidTokens.ToDec())
		}
		if votingPower.IsPositive() {
			(*votingPowers)[voter] = map[string]sdk.Dec{}
			// drop crumb for defensive policy about delShares decimal errors
			dividedPowers, _ := types.DivideByCurrentWeight(liquidVals, votingPower, totalBondedLiquidTokens, bondedLiquidTokenMap)
			for i, val := range liquidVals {
				if !dividedPowers[i].IsPositive() {
					continue
				}
				if existed, ok := (*votingPowers)[voter][val.OperatorAddress]; ok {
					(*votingPowers)[voter][val.OperatorAddress] = existed.Add(dividedPowers[i])
				} else {
					(*votingPowers)[voter][val.OperatorAddress] = dividedPowers[i]
				}
			}
		}
	}
}
