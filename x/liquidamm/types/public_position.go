package types

import (
	"fmt"
	"regexp"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v5/types"
)

const (
	ShareDenomPrefix = "sb"
)

var (
	shareDenomRe = regexp.MustCompile(`^sb([1-9]\d*)$`)
)

// DeriveBidReserveAddress creates the reserve address for bids
// with the given public position id.
func DeriveBidReserveAddress(positionId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("BidReserveAddress/%d", positionId)))
}

// NewPublicPosition returns a new PublicPosition.
func NewPublicPosition(
	id, poolId uint64, lowerTick, upperTick int32, feeRate sdk.Dec) PublicPosition {
	return PublicPosition{
		Id:                   id,
		PoolId:               poolId,
		LowerTick:            lowerTick,
		UpperTick:            upperTick,
		BidReserveAddress:    DeriveBidReserveAddress(id).String(),
		FeeRate:              feeRate,
		LastRewardsAuctionId: 0,
	}
}

// Validate validates PublicPosition.
func (publicPosition PublicPosition) Validate() error {
	if publicPosition.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if publicPosition.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if publicPosition.LowerTick >= publicPosition.UpperTick {
		return fmt.Errorf("lower tick must be lower than upper tick")
	}
	if _, err := sdk.AccAddressFromBech32(publicPosition.BidReserveAddress); err != nil {
		return fmt.Errorf("invalid bid reserve address %w", err)
	}
	if publicPosition.FeeRate.GT(utils.OneDec) || publicPosition.FeeRate.IsNegative() {
		return fmt.Errorf("fee rate must be in range [0, 1]: %s", publicPosition.FeeRate)
	}
	return nil
}

func (publicPosition PublicPosition) MustGetBidReserveAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(publicPosition.BidReserveAddress)
}

// ShareDenom returns a unique public position share denom.
func ShareDenom(publicPositionId uint64) string {
	return fmt.Sprintf("%s%d", ShareDenomPrefix, publicPositionId)
}

// ParseShareDenom parses a public position share denom and returns the position's id.
func ParseShareDenom(denom string) (publicPositionId uint64, err error) {
	chunks := shareDenomRe.FindStringSubmatch(denom)
	if len(chunks) == 0 {
		return 0, fmt.Errorf("invalid share denom: %s", denom)
	}
	publicPositionId, err = strconv.ParseUint(chunks[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid public position id: %s", chunks[1])
	}
	return publicPositionId, nil
}

func CalculateMintRate(totalLiquidity, shareSupply sdk.Int) sdk.Dec {
	if shareSupply.IsZero() { // initial minting
		return utils.OneDec
	}
	return shareSupply.ToDec().QuoTruncate(totalLiquidity.ToDec())
}

// CalculateMintedShareAmount calculates minting public position share amount.
// mintedShareAmt = shareSupply * (addedLiquidity / totalLiquidity)
func CalculateMintedShareAmount(addedLiquidity, totalLiquidity, shareSupply sdk.Int) sdk.Int {
	if shareSupply.IsZero() { // initial minting
		return addedLiquidity
	}
	return shareSupply.Mul(addedLiquidity).Quo(totalLiquidity)
}

func CalculateBurnRate(shareSupply, totalLiquidity, prevWinningBidShareAmt sdk.Int) sdk.Dec {
	if shareSupply.Add(prevWinningBidShareAmt).IsZero() { // no more share
		return utils.ZeroDec
	}
	return totalLiquidity.ToDec().QuoTruncate(shareSupply.Add(prevWinningBidShareAmt).ToDec())
}

// CalculateRemovedLiquidity calculates liquidity amount to be removed when
// burning public position share.
// removedLiquidity = totalLiquidity * (burnedShareAmt / (shareSupply + prevWinningBidShareAmt))
func CalculateRemovedLiquidity(
	burnedShareAmt, shareSupply, totalLiquidity, prevWinningBidShareAmt sdk.Int) sdk.Int {
	if burnedShareAmt.Equal(shareSupply) { // last one to burn
		return totalLiquidity
	}
	if shareSupply.Add(prevWinningBidShareAmt).IsZero() {
		return sdk.ZeroInt()
	}
	return totalLiquidity.Mul(burnedShareAmt).Quo(shareSupply.Add(prevWinningBidShareAmt))
}

// DeductFees deducts fees from rewards by the fee rate.
func DeductFees(rewards sdk.Coins, feeRate sdk.Dec) (deductedRewards sdk.Coins, fees sdk.Coins) {
	deductedRewards, _ = sdk.NewDecCoinsFromCoins(rewards...).MulDecTruncate(utils.OneDec.Sub(feeRate)).TruncateDecimal()
	fees = rewards.Sub(deductedRewards)
	return
}
