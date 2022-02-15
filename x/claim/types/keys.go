package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName defines the module name
	ModuleName = "claim"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// Keys for store prefixes
var (
	LastAirdropIdKey = []byte{0xd0} // key for the latest airdrop id

	StartTimeKeyPrefix              = []byte{0xd5}
	EndTimeKeyPrefix                = []byte{0xd6}
	AirdropKeyPrefix                = []byte{0xd7}
	ClaimRecordKeyPrefix            = []byte{0xd8}
	ClaimRecordByRecipientKeyPrefix = []byte{0xd9}
)

// GetStartTimeKey returns the store key to retrieve the start time for the airdrop.
func GetStartTimeKey(auctionId uint64) []byte {
	return append(StartTimeKeyPrefix, sdk.Uint64ToBigEndian(auctionId)...)
}

// GetEndTimeKey returns the store key to retrieve the end time for the airdrop.
func GetEndTimeKey(auctionId uint64) []byte {
	return append(EndTimeKeyPrefix, sdk.Uint64ToBigEndian(auctionId)...)
}

// GetAirdropKey returns the store key to retrieve the airdrop object from the airdrop id.
func GetAirdropKey(airdropId uint64) []byte {
	return append(AirdropKeyPrefix, sdk.Uint64ToBigEndian(airdropId)...)
}

// GetClaimRecordKey returns the store key to retrieve the claim record by the airdrop id.
func GetClaimRecordKey(airdropId uint64) []byte {
	return append(ClaimRecordKeyPrefix, sdk.Uint64ToBigEndian(airdropId)...)
}

// GetClaimRecordByRecipientKey returns the tore key to retrieve the claim record by the airdrop id and the recipient address.
func GetClaimRecordByRecipientKey(airdropId uint64, recipient sdk.AccAddress) []byte {
	return append(append(ClaimRecordByRecipientKeyPrefix, sdk.Uint64ToBigEndian(airdropId)...), address.MustLengthPrefix(recipient)...)
}
