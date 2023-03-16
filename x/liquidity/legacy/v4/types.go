package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

var (
	MMOrderIndexKeyPrefix = []byte{0xb6}
)

// GetMMOrderIndexKey returns the store key to retrieve MMOrderIndex object by
// orderer and pair id.
func GetMMOrderIndexKey(orderer sdk.AccAddress, pairId uint64) []byte {
	return append(append(MMOrderIndexKeyPrefix, address.MustLengthPrefix(orderer)...), sdk.Uint64ToBigEndian(pairId)...)
}
