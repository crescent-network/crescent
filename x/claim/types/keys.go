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

var (
	ClaimRecordKeyPrefix = []byte{0xc0}
)

// GetClaimRecordKey returns the store key to retrieve ClaimRecord.
func GetClaimRecordKey(recipient sdk.AccAddress) []byte {
	return append(ClaimRecordKeyPrefix, address.MustLengthPrefix(recipient)...)
}
