package types

import (
	fmt "fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingtypes "github.com/crescent-network/crescent/v2/x/farming/types"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

const (
	LiquidFarmReserveAccPrefix    string = "LiquidFarmReserveAcc"
	LiquidFarmFeeCollectorAccName string = "LiquidFarmFeeCollectorAcc"
)

var (
	liquidFarmCoinDenomRegexp = regexp.MustCompile(`^lf([1-9]\d*)$`)

	// LiquidFarmFeeCollectorAcc is the fee collector address for liquid farms that collects all fees generated from the module.
	LiquidFarmFeeCollectorAcc = farmingtypes.DeriveAddress(farmingtypes.ReserveAddressType, types.ModuleName, LiquidFarmFeeCollectorAccName)
)

// NewLiquidFarm returns a new LiquidFarm.
func NewLiquidFarm(poolId uint64, minFarmAmt, minBidAmount sdk.Int, feeRate sdk.Dec) LiquidFarm {
	return LiquidFarm{
		PoolId:        poolId,
		MinFarmAmount: minFarmAmt,
		MinBidAmount:  minBidAmount,
		FeeRate:       feeRate,
	}
}

// Validate validates LiquidFarm.
func (l LiquidFarm) Validate() error {
	if l.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if l.MinBidAmount.IsNegative() {
		return fmt.Errorf("minimum bid amount must be 0 or positive value: %s", l.MinBidAmount)
	}
	if l.MinFarmAmount.IsNegative() {
		return fmt.Errorf("minimum farm amount must be 0 or positive value: %s", l.MinFarmAmount)
	}
	if l.FeeRate.IsNegative() {
		return fmt.Errorf("fee rate must be 0 or positive value: %s", l.FeeRate)
	}
	return nil
}

// String returns a human-readable string representation of the LiquidFarm.
func (l LiquidFarm) String() string {
	out, _ := l.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a LiquidFarm.
func (l LiquidFarm) MarshalYAML() (interface{}, error) {
	bz, err := codec.MarshalYAML(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), &l)
	if err != nil {
		return nil, err
	}
	return string(bz), err
}

// LiquidFarmCoinDenom returns a unique liquid farming coin denom for a LiquidFarm.
func LiquidFarmCoinDenom(poolId uint64) string {
	return fmt.Sprintf("lf%d", poolId)
}

// ParseLiquidFarmCoinDenom parses a LF coin denom and returns its pool id.
func ParseLiquidFarmCoinDenom(denom string) (poolId uint64, err error) {
	chunks := liquidFarmCoinDenomRegexp.FindStringSubmatch(denom)
	if len(chunks) == 0 {
		return 0, fmt.Errorf("%s is not a liquid farm coin denom", denom)
	}
	poolId, err = strconv.ParseUint(chunks[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse pool id: %w", err)
	}
	return poolId, nil
}

// LiquidFarmReserveAddress returns the reserve address for a liquid farm with the given pool id.
func LiquidFarmReserveAddress(poolId uint64) sdk.AccAddress {
	return farmingtypes.DeriveAddress(
		ReserveAddressType,
		ModuleName,
		strings.Join([]string{LiquidFarmReserveAccPrefix, strconv.FormatUint(poolId, 10)}, ModuleAddressNameSplitter),
	)
}

// CalculateFarmMintingAmount calculates minting liquid farm amount.
// MintingAmt = LFCoinTotalSupply / (LPCoinTotalStaked + LPCoinTotalQueued) * LPCoinFarmingAmount
func CalculateFarmMintingAmount(
	lfCoinTotalSupplyAmt sdk.Int,
	lpCoinTotalQueuedAmt sdk.Int,
	lpCoinTotalStakedAmt sdk.Int,
	newFarmingAmt sdk.Int,
) sdk.Int {
	if lfCoinTotalSupplyAmt.IsZero() { // initial minting
		return newFarmingAmt
	}
	totalFarmingAmt := lpCoinTotalStakedAmt.Add(lpCoinTotalQueuedAmt)
	return lfCoinTotalSupplyAmt.Mul(newFarmingAmt).Quo(totalFarmingAmt)
}

// CalculateUnfarmedAmount calculates unfarmed amount.
// UnfarmAmount = LPCoinTotalStaked + LPCoinTotalQueued - CompoundingRewards / LFCoinTotalSupply * LFCoinUnfarmingAmount
func CalculateUnfarmedAmount(
	lfCoinTotalSupplyAmt sdk.Int,
	lpCoinTotalStakedAmt sdk.Int,
	lpCoinTotalQueuedAmt sdk.Int,
	unfarmingAmt sdk.Int,
	compoundingRewards sdk.Int,
) sdk.Int {
	if lfCoinTotalSupplyAmt.Equal(unfarmingAmt) {
		return lpCoinTotalStakedAmt.Add(lpCoinTotalQueuedAmt)
	}
	totalFarmingAmt := lpCoinTotalStakedAmt.Add(lpCoinTotalQueuedAmt).Sub(compoundingRewards)
	return totalFarmingAmt.Mul(unfarmingAmt).Quo(lfCoinTotalSupplyAmt)
}

// DeductFees deducts fee rates from the farming rewards.
func DeductFees(poolId uint64, rewards sdk.Coins, feeRate sdk.Dec) (sdk.Coins, error) {
	for i, reward := range rewards {
		multiplier := sdk.OneDec().Sub(feeRate)                                                     // 1 - feeRate
		rewards[i] = sdk.NewCoin(reward.Denom, reward.Amount.ToDec().Mul(multiplier).TruncateInt()) // reward * multiplier
	}
	return rewards, nil
}
