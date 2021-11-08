package types

import (
	"bytes"

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
)

// keys for farming store prefixes
var (
	GlobalPlanIdKey     = []byte("globalPlanId")
	LastEpochTimeKey    = []byte("lastEpochTime")
	CurrentEpochDaysKey = []byte("currentEpochDays")

	PlanKeyPrefix = []byte{0x11}

	StakingKeyPrefix            = []byte{0x21}
	StakingIndexKeyPrefix       = []byte{0x22}
	QueuedStakingKeyPrefix      = []byte{0x23}
	QueuedStakingIndexKeyPrefix = []byte{0x24}
	TotalStakingKeyPrefix       = []byte{0x25}

	HistoricalRewardsKeyPrefix  = []byte{0x31}
	CurrentEpochKeyPrefix       = []byte{0x32}
	OutstandingRewardsKeyPrefix = []byte{0x33}
)

// GetPlanKey returns kv indexing key of the plan
func GetPlanKey(planID uint64) []byte {
	return append(PlanKeyPrefix, sdk.Uint64ToBigEndian(planID)...)
}

// GetStakingKey returns a key for staking of corresponding the id
func GetStakingKey(stakingCoinDenom string, farmerAcc sdk.AccAddress) []byte {
	return append(append(StakingKeyPrefix, LengthPrefixString(stakingCoinDenom)...), farmerAcc...)
}

// GetStakingIndexKey returns an indexing key for a staking.
func GetStakingIndexKey(farmerAcc sdk.AccAddress, stakingCoinDenom string) []byte {
	return append(append(StakingIndexKeyPrefix, address.MustLengthPrefix(farmerAcc)...), []byte(stakingCoinDenom)...)
}

// GetStakingsByFarmerPrefix returns a key prefix used to iterate
// stakings by a farmer.
func GetStakingsByFarmerPrefix(farmerAcc sdk.AccAddress) []byte {
	return append(StakingIndexKeyPrefix, address.MustLengthPrefix(farmerAcc)...)
}

// GetQueuedStakingKey returns a key for a queued staking.
func GetQueuedStakingKey(stakingCoinDenom string, farmerAcc sdk.AccAddress) []byte {
	return append(append(QueuedStakingKeyPrefix, LengthPrefixString(stakingCoinDenom)...), farmerAcc...)
}

// GetQueuedStakingIndexKey returns an indexing key for a queuded staking.
func GetQueuedStakingIndexKey(farmerAcc sdk.AccAddress, stakingCoinDenom string) []byte {
	return append(append(QueuedStakingIndexKeyPrefix, address.MustLengthPrefix(farmerAcc)...), []byte(stakingCoinDenom)...)
}

// GetQueuedStakingByFarmerPrefix returns a key prefix used to iterate
// queued stakings by a farmer.
func GetQueuedStakingByFarmerPrefix(farmerAcc sdk.AccAddress) []byte {
	return append(QueuedStakingIndexKeyPrefix, address.MustLengthPrefix(farmerAcc)...)
}

// GetTotalStakingsKey returns a key for a total stakings info.
func GetTotalStakingsKey(stakingCoinDenom string) []byte {
	return append(TotalStakingKeyPrefix, []byte(stakingCoinDenom)...)
}

// GetHistoricalRewardsKey returns a key for a historical rewards record.
func GetHistoricalRewardsKey(stakingCoinDenom string, epoch uint64) []byte {
	return append(append(HistoricalRewardsKeyPrefix, LengthPrefixString(stakingCoinDenom)...), sdk.Uint64ToBigEndian(epoch)...)
}

// GetHistoricalRewardsPrefix returns a key prefix used to iterate
// historical rewards by a staking coin denom.
func GetHistoricalRewardsPrefix(stakingCoinDenom string) []byte {
	return append(HistoricalRewardsKeyPrefix, LengthPrefixString(stakingCoinDenom)...)
}

// GetCurrentEpochKey returns a key for a current epoch info.
func GetCurrentEpochKey(stakingCoinDenom string) []byte {
	return append(CurrentEpochKeyPrefix, []byte(stakingCoinDenom)...)
}

// GetOutstandingRewardsKey returns a key for an outstanding rewards record.
func GetOutstandingRewardsKey(stakingCoinDenom string) []byte {
	return append(OutstandingRewardsKeyPrefix, []byte(stakingCoinDenom)...)
}

// ParseStakingKey parses a staking key.
func ParseStakingKey(key []byte) (stakingCoinDenom string, farmerAcc sdk.AccAddress) {
	if !bytes.HasPrefix(key, StakingKeyPrefix) {
		panic("key does not have proper prefix")
	}
	denomLen := key[1]
	stakingCoinDenom = string(key[2 : 2+denomLen])
	farmerAcc = key[2+denomLen:]
	return
}

// ParseStakingIndexKey parses a staking index key.
func ParseStakingIndexKey(key []byte) (farmerAcc sdk.AccAddress, stakingCoinDenom string) {
	if !bytes.HasPrefix(key, StakingIndexKeyPrefix) {
		panic("key does not have proper prefix")
	}
	addrLen := key[1]
	farmerAcc = key[2 : 2+addrLen]
	stakingCoinDenom = string(key[2+addrLen:])
	return
}

// ParseQueuedStakingKey parses a queued staking key.
func ParseQueuedStakingKey(key []byte) (stakingCoinDenom string, farmerAcc sdk.AccAddress) {
	if !bytes.HasPrefix(key, QueuedStakingKeyPrefix) {
		panic("key does not have proper prefix")
	}
	denomLen := key[1]
	stakingCoinDenom = string(key[2 : 2+denomLen])
	farmerAcc = key[2+denomLen:]
	return
}

// ParseQueuedStakingIndexKey parses a queued staking index key.
func ParseQueuedStakingIndexKey(key []byte) (farmerAcc sdk.AccAddress, stakingCoinDenom string) {
	if !bytes.HasPrefix(key, QueuedStakingIndexKeyPrefix) {
		panic("key does not have proper prefix")
	}
	addrLen := key[1]
	farmerAcc = key[2 : 2+addrLen]
	stakingCoinDenom = string(key[2+addrLen:])
	return
}

// ParseTotalStakingsKey parses a total stakings key.
func ParseTotalStakingsKey(key []byte) (stakingCoinDenom string) {
	if !bytes.HasPrefix(key, TotalStakingKeyPrefix) {
		panic("key does not have proper prefix")
	}
	stakingCoinDenom = string(key[1:])
	return
}

// ParseHistoricalRewardsKey parses a historical rewards key.
func ParseHistoricalRewardsKey(key []byte) (stakingCoinDenom string, epoch uint64) {
	if !bytes.HasPrefix(key, HistoricalRewardsKeyPrefix) {
		panic("key does not have proper prefix")
	}
	denomLen := key[1]
	stakingCoinDenom = string(key[2 : 2+denomLen])
	epoch = sdk.BigEndianToUint64(key[2+denomLen:])
	return
}

// ParseCurrentEpochKey parses a current epoch key.
func ParseCurrentEpochKey(key []byte) (stakingCoinDenom string) {
	if !bytes.HasPrefix(key, CurrentEpochKeyPrefix) {
		panic("key does not have proper prefix")
	}
	stakingCoinDenom = string(key[1:])
	return
}

// ParseOutstandingRewardsKey parses an outstanding rewards key.
func ParseOutstandingRewardsKey(key []byte) (stakingCoinDenom string) {
	if !bytes.HasPrefix(key, OutstandingRewardsKeyPrefix) {
		panic("key does not have proper prefix")
	}
	stakingCoinDenom = string(key[1:])
	return
}

// LengthPrefixString returns length-prefixed bytes representation
// of a string.
func LengthPrefixString(s string) []byte {
	bz := []byte(s)
	bzLen := len(bz)
	if bzLen == 0 {
		return bz
	}
	return append([]byte{byte(bzLen)}, bz...)
}
