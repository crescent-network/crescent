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
)

var (
	GlobalPlanIdKey        = []byte("globalPlanId")
	GlobalLastEpochTimeKey = []byte("globalLastEpochTime")

	PlanKeyPrefix               = []byte{0x11}
	PlansByFarmerIndexKeyPrefix = []byte{0x12}

	StakingKeyPrefix       = []byte{0x21}
	StakingIndexKeyPrefix  = []byte{0x22}
	QueuedStakingKeyPrefix = []byte{0x23}
	TotalStakingKeyPrefix  = []byte{0x24}

	HistoricalRewardsKeyPrefix = []byte{0x31}
	CurrentRewardsKeyPrefix    = []byte{0x32}
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
func GetStakingKey(stakingCoinDenom string, farmerAcc sdk.AccAddress) []byte {
	return append(append(StakingKeyPrefix, LengthPrefixString(stakingCoinDenom)...), farmerAcc...)
}

func GetQueuedStakingKey(stakingCoinDenom string, farmerAcc sdk.AccAddress) []byte {
	return append(append(QueuedStakingKeyPrefix, LengthPrefixString(stakingCoinDenom)...), farmerAcc...)
}

func GetTotalStakingKey(stakingCoinDenom string) []byte {
	return append(TotalStakingKeyPrefix, []byte(stakingCoinDenom)...)
}

func GetHistoricalRewardsKey(stakingCoinDenom string, epoch uint64) []byte {
	return append(append(HistoricalRewardsKeyPrefix, LengthPrefixString(stakingCoinDenom)...), sdk.Uint64ToBigEndian(epoch)...)
}

func GetCurrentRewardsKey(stakingCoinDenom string) []byte {
	return append(CurrentRewardsKeyPrefix, []byte(stakingCoinDenom)...)
}

func ParseQueuedStakingKey(key []byte) (stakingCoinDenom string, farmerAcc sdk.AccAddress) {
	denomLen := key[1]
	stakingCoinDenom = string(key[2 : 2+denomLen])
	farmerAcc = key[2+denomLen:]
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
