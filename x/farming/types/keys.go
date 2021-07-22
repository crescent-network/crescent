package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the farming module
	ModuleName = "farming"

	// RouterKey is the message router key for the farming module
	RouterKey = ModuleName

	// StoreKey is the default store key for the farming module
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the farming module
	QuerierRoute = ModuleName

	// RewardPoolAccKeyPrefix is prefix for generating deterministic reward pool module account of the each plan
	RewardPoolAccKeyPrefix = "RewardPoolAcc"

	// StakingReserveAccKeyPrefix is prefix for generating deterministic staking reserve module account of the each plan
	StakingReserveAccKeyPrefix = "StakingReserveAcc"
)

var (
	// param key for global farming plan IDs
	GlobalPlanIdKey           = []byte("globalPlanId")
	GlobalLastEpochTimePrefix = []byte("globalLastEpochTime")
	GlobalStakingIdKey        = []byte("globalStakingId")

	PlanKeyPrefix                     = []byte{0x11}
	PlansByFarmerIndexKeyPrefix       = []byte{0x12}
	LastDistributedTimeKeyPrefix      = []byte{0x13}
	TotalDistributedRewardCoinsPrefix = []byte{0x14}

	StakingKeyPrefix                         = []byte{0x21}
	StakingByFarmerIndexKeyPrefix            = []byte{0x22}
	StakingsByStakingCoinDenomIndexKeyPrefix = []byte{0x23}

	RewardKeyPrefix               = []byte{0x31}
	RewardsByFarmerIndexKeyPrefix = []byte{0x32}

	StakingReserveAcc = sdk.AccAddress(address.Module(ModuleName, []byte("StakingReserveAcc")))
)

// GetPlanKey returns kv indexing key of the plan
func GetPlanKey(planID uint64) []byte {
	return append(PlanKeyPrefix, sdk.Uint64ToBigEndian(planID)...)
}

// GetPlansByFarmerIndexKey returns kv indexing key of the plan indexed by reserve account
func GetPlansByFarmerIndexKey(farmerAcc sdk.AccAddress) []byte {
	return append(PlansByFarmerIndexKeyPrefix, address.MustLengthPrefix(farmerAcc.Bytes())...)
}

// GetPlanByFarmerAddrIndexKey returns kv indexing key of the plan indexed by reserve account
func GetPlanByFarmerAddrIndexKey(farmerAcc sdk.AccAddress, planID uint64) []byte {
	return append(append(PlansByFarmerIndexKeyPrefix, address.MustLengthPrefix(farmerAcc.Bytes())...), sdk.Uint64ToBigEndian(planID)...)
}

// GetStakingKey returns a key for staking of corresponding the id
func GetStakingKey(id uint64) []byte {
	return append(StakingKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

// GetStakingByFarmerIndexKey returns key for the farmer's staking of corresponding
func GetStakingByFarmerIndexKey(farmerAcc sdk.AccAddress) []byte {
	return append(StakingByFarmerIndexKeyPrefix, address.MustLengthPrefix(farmerAcc.Bytes())...)
}

// GetStakingsByStakingCoinDenomIndexKey returns prefix for the iterable staking list by the staking coin denomination
func GetStakingsByStakingCoinDenomIndexKey(denom string) []byte {
	return append(StakingsByStakingCoinDenomIndexKeyPrefix, LengthPrefixString(denom)...)
}

// GetStakingByStakingCoinDenomIndexKey returns key for the staking index by the staking coin denomination
func GetStakingByStakingCoinDenomIndexKey(denom string, id uint64) []byte {
	return append(append(StakingsByStakingCoinDenomIndexKeyPrefix, LengthPrefixString(denom)...), sdk.Uint64ToBigEndian(id)...)
}

// GetRewardKey returns key for staking coin denomination's reward of corresponding the farmer
func GetRewardKey(stakingCoinDenom string, farmerAcc sdk.AccAddress) []byte {
	return append(append(RewardKeyPrefix, LengthPrefixString(stakingCoinDenom)...), address.MustLengthPrefix(farmerAcc.Bytes())...)
}

// GetRewardByFarmerAndStakingCoinDenomIndexKey returns key for farmer's reward of corresponding the staking coin denomination
func GetRewardByFarmerAndStakingCoinDenomIndexKey(farmerAcc sdk.AccAddress, stakingCoinDenom string) []byte {
	return append(append(RewardsByFarmerIndexKeyPrefix, address.MustLengthPrefix(farmerAcc.Bytes())...), LengthPrefixString(stakingCoinDenom)...)
}

// GetRewardsByStakingCoinDenomKey returns prefix for staking coin denomination's reward list
func GetRewardsByStakingCoinDenomKey(stakingCoinDenom string) []byte {
	return append(RewardKeyPrefix, LengthPrefixString(stakingCoinDenom)...)
}

// GetRewardsByFarmerIndexKey returns prefix for farmer's reward list
func GetRewardsByFarmerIndexKey(farmerAcc sdk.AccAddress) []byte {
	return append(RewardsByFarmerIndexKeyPrefix, address.MustLengthPrefix(farmerAcc.Bytes())...)
}

// ParseRewardKey parses a RewardKey.
func ParseRewardKey(key []byte) (stakingCoinDenom string, farmerAcc sdk.AccAddress) {
	denomLen := key[1]
	stakingCoinDenom = string(key[2 : 2+denomLen])
	farmerAcc = key[2+denomLen+1:]
	return
}

// ParseRewardsByFarmerIndexKey parses a key of RewardsByFarmerIndex from bytes.
func ParseRewardsByFarmerIndexKey(key []byte) (farmerAcc sdk.AccAddress, stakingCoinDenom string) {
	addrLen := key[1]
	farmerAcc = key[2 : 2+addrLen]
	stakingCoinDenom = string(key[2+addrLen+1:])
	return
}

// ParseStakingsByStakingCoinDenomIndexKey parses a key of StakingsByStakingCoinDenomIndex from bytes.
func ParseStakingsByStakingCoinDenomIndexKey(bz []byte) (stakingCoinDenom string, stakingID uint64) {
	denomLen := bz[1]
	stakingCoinDenom = string(bz[2 : 2+denomLen])
	stakingID = sdk.BigEndianToUint64(bz[2+denomLen:])
	return
}

// LengthPrefixString is LengthPrefix for string.
func LengthPrefixString(s string) []byte {
	bz := []byte(s)
	bzLen := len(bz)
	if bzLen == 0 {
		return bz
	}
	return append([]byte{byte(bzLen)}, bz...)
}
