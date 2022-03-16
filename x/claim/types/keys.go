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
	AirdropKeyPrefix     = []byte{0xd5}
	ClaimRecordKeyPrefix = []byte{0xd6}
)

// GetAirdropKey returns the store key to retrieve the airdrop object from the airdrop id.
func GetAirdropKey(airdropId uint64) []byte {
	return append(AirdropKeyPrefix, sdk.Uint64ToBigEndian(airdropId)...)
}

// GetClaimRecordsByAirdropKeyPrefix returns the store key to retrieve the claim record by the airdrop id.
func GetClaimRecordsByAirdropKeyPrefix(airdropId uint64) []byte {
	return append(ClaimRecordKeyPrefix, sdk.Uint64ToBigEndian(airdropId)...)
}

// GetClaimRecordKey returns the tore key to retrieve the claim record by the airdrop id and the recipient address.
func GetClaimRecordKey(airdropId uint64, recipient sdk.AccAddress) []byte {
	return append(append(ClaimRecordKeyPrefix, sdk.Uint64ToBigEndian(airdropId)...), address.MustLengthPrefix(recipient)...)
}
