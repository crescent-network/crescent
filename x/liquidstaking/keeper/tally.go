package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	liquidammtypes "github.com/crescent-network/crescent/v5/x/liquidamm/types"
	"github.com/crescent-network/crescent/v5/x/liquidstaking/types"
	lpfarmtypes "github.com/crescent-network/crescent/v5/x/lpfarm/types"
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

func (k Keeper) TokenAmountFromAMMPositions(ctx sdk.Context, addr sdk.AccAddress, targetDenom string) sdk.Int {
	tokenAmount := utils.ZeroInt

	k.ammKeeper.IteratePositionsByOwner(ctx, addr, func(position ammtypes.Position) (stop bool) {
		coin0, coin1, err := k.ammKeeper.PositionAssets(ctx, position.Id)
		if err != nil {
			k.Logger(ctx).Error("failed to get position assets", "error", err)
			return false
		}
		if coin0.Denom == targetDenom {
			tokenAmount = tokenAmount.Add(coin0.Amount)
		} else if coin1.Denom == targetDenom {
			tokenAmount = tokenAmount.Add(coin1.Amount)
		}
		return false
	})

	return tokenAmount
}

// GetBTokenSharePerPublicPositionShareMap creates a map of public position
// share which contains targetDenom to calculate target denom value of public positions
func (k Keeper) GetBTokenSharePerPublicPositionShareMap(ctx sdk.Context, targetDenom string) map[string]sdk.Dec {
	m := map[string]sdk.Dec{}
	k.liquidAMMKeeper.IterateAllPublicPositions(ctx, func(publicPosition liquidammtypes.PublicPosition) (stop bool) {
		shareDenom := liquidammtypes.ShareDenom(publicPosition.Id)
		bTokenSharePerPublicPositionShare := k.BTokenSharePerPublicPositionShare(ctx, publicPosition, targetDenom)
		if bTokenSharePerPublicPositionShare.IsPositive() {
			m[shareDenom] = bTokenSharePerPublicPositionShare
		}
		return false
	})
	return m
}

func (k Keeper) BTokenSharePerPublicPositionShare(ctx sdk.Context, publicPosition liquidammtypes.PublicPosition, targetDenom string) sdk.Dec {
	shareDenom := liquidammtypes.ShareDenom(publicPosition.Id)

	ammPosition := k.liquidAMMKeeper.MustGetAMMPosition(ctx, publicPosition)
	coin0, coin1, err := k.ammKeeper.PositionAssets(ctx, ammPosition.Id)
	if err != nil {
		// TODO: panic instead?
		return utils.ZeroDec
	}
	if coin0.Denom != targetDenom && coin1.Denom != targetDenom {
		// The pool doesn't have target denom in its assets.
		return utils.ZeroDec
	}

	publicPositionShareSupply := k.bankKeeper.GetSupply(ctx, shareDenom).Amount
	if !publicPositionShareSupply.IsPositive() {
		// No liquidity in the public position.
		return utils.ZeroDec
	}

	bTokenSharePerPublicPositionShare := utils.ZeroDec
	if coin0.Denom == targetDenom {
		bTokenSharePerPublicPositionShare = coin0.Amount.ToDec().
			QuoTruncate(publicPositionShareSupply.ToDec())
	} else if coin1.Denom == targetDenom {
		bTokenSharePerPublicPositionShare = coin1.Amount.ToDec().
			QuoTruncate(publicPositionShareSupply.ToDec())
	}
	return bTokenSharePerPublicPositionShare
}

// TokenAmountFromFarmingPositions returns worth of staked tokens amount of exist farming positions including queued of the addr
func (k Keeper) TokenAmountFromFarmingPositions(ctx sdk.Context, addr sdk.AccAddress, targetDenom string, tokenSharePerPoolCoinMap map[string]sdk.Dec) sdk.Int {
	tokenAmount := sdk.ZeroInt()

	k.lpfarmKeeper.GetPosition(ctx, addr, targetDenom)
	k.lpfarmKeeper.IteratePositionsByFarmer(ctx, addr, func(position lpfarmtypes.Position) bool {
		if position.Denom == targetDenom {
			tokenAmount = tokenAmount.Add(position.FarmingAmount)
		} else if ratio, ok := tokenSharePerPoolCoinMap[position.Denom]; ok {
			tokenAmount = tokenAmount.Add(utils.GetShareValue(position.FarmingAmount, ratio))
		}
		return false
	})

	return tokenAmount
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
	bTokenSharePerPublicPositionShareMap := k.GetBTokenSharePerPublicPositionShareMap(ctx, liquidBondDenom)

	balances := k.bankKeeper.SpendableCoins(ctx, addr)
	for _, coin := range balances {
		// add balance of bToken
		if coin.Denom == liquidBondDenom {
			bTokenAmount = bTokenAmount.Add(coin.Amount)
		}

		// check if the denom is pool coin
		if bTokenSharePerPublicPositionShare, ok := bTokenSharePerPublicPositionShareMap[coin.Denom]; ok {
			bTokenAmount = bTokenAmount.Add(utils.GetShareValue(coin.Amount, bTokenSharePerPublicPositionShare))
		}
	}

	tokenAmount := k.TokenAmountFromAMMPositions(ctx, addr, liquidBondDenom)
	if tokenAmount.IsPositive() {
		bTokenAmount = bTokenAmount.Add(tokenAmount)
	}

	tokenAmount = k.TokenAmountFromFarmingPositions(ctx, addr, liquidBondDenom, bTokenSharePerPublicPositionShareMap)
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
	bTokenSharePerPublicPositionShareMap := k.GetBTokenSharePerPublicPositionShareMap(ctx, liquidBondDenom)
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

		// if the denom is public position share, get bToken share and add owned bToken on bTokenOwnMap
		if bTokenSharePerPublicPositionShare, ok := bTokenSharePerPublicPositionShareMap[denom]; ok {
			for voter, balance := range voterBalanceByDenom[denom] {
				bTokenOwnMap.AddOrSet(voter, utils.GetShareValue(balance, bTokenSharePerPublicPositionShare))
			}
		}
	}

	// add owned btoken amount of farming positions on bTokenOwnMap
	for _, vote := range votes {
		voter, err := sdk.AccAddressFromBech32(vote.Voter)
		if err != nil {
			continue
		}
		tokenAmount := k.TokenAmountFromAMMPositions(ctx, voter, liquidBondDenom)
		if tokenAmount.IsPositive() {
			bTokenOwnMap.AddOrSet(vote.Voter, tokenAmount)
		}
		tokenAmount = k.TokenAmountFromFarmingPositions(ctx, voter, liquidBondDenom, bTokenSharePerPublicPositionShareMap)
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
