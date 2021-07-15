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
	GlobalFarmingPlanIDKey = []byte("globalFarmingPlanId")

	PlanKeyPrefix                  = []byte{0x11}
	PlanByFarmerAddrIndexKeyPrefix = []byte{0x12}
	LastEpochTimeKeyPrefix         = []byte{0x13}

	StakingKeyPrefix = []byte{0x21}

	RewardKeyPrefix = []byte{0x31}
)

// GetPlanKey returns kv indexing key of the plan
func GetPlanKey(planID uint64) []byte {
	key := make([]byte, 9)
	key[0] = PlanKeyPrefix[0]
	copy(key[1:], sdk.Uint64ToBigEndian(planID))
	return key
}

// GetPlansByFarmerAddrIndexKey returns kv indexing key of the plan indexed by reserve account
func GetPlansByFarmerAddrIndexKey(farmerAcc sdk.AccAddress) []byte {
	return append(PlanByFarmerAddrIndexKeyPrefix, address.MustLengthPrefix(farmerAcc.Bytes())...)
}

// GetPlanByFarmerAddrIndexKey returns kv indexing key of the plan indexed by reserve account
func GetPlanByFarmerAddrIndexKey(farmerAcc sdk.AccAddress, planID uint64) []byte {
	return append(append(PlanByFarmerAddrIndexKeyPrefix, address.MustLengthPrefix(farmerAcc.Bytes())...), sdk.Uint64ToBigEndian(planID)...)
}

// GetStakingPrefix returns prefix of staking records in the plan
func GetStakingPrefix(planID uint64) []byte {
	key := make([]byte, 9)
	key[0] = StakingKeyPrefix[0]
	copy(key[1:9], sdk.Uint64ToBigEndian(planID))
	return key
}

// GetStakingIndexKey returns key for farmer's staking of corresponding the plan id
func GetStakingIndexKey(planID uint64, farmerAcc sdk.AccAddress) []byte {
	// TODO: review for addrLen,  <addrLen (1 Byte)><addrBytes>
	return append(append(StakingKeyPrefix, sdk.Uint64ToBigEndian(planID)...), address.MustLengthPrefix(farmerAcc.Bytes())...)
}

// GetRewardPrefix returns prefix of reward records in the plan
func GetRewardPrefix(planID uint64) []byte {
	key := make([]byte, 9)
	key[0] = RewardKeyPrefix[0]
	copy(key[1:9], sdk.Uint64ToBigEndian(planID))
	return key
}

// GetRewardIndexKey returns key for farmer's reward of corresponding the plan id
func GetRewardIndexKey(planID uint64, farmerAcc sdk.AccAddress) []byte {
	// TODO: review for addrLen,  <addrLen (1 Byte)><addrBytes>
	return append(append(RewardKeyPrefix, sdk.Uint64ToBigEndian(planID)...), address.MustLengthPrefix(farmerAcc.Bytes())...)
}
