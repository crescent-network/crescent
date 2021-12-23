package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the liquidstaking module
	ModuleName = "liquidstaking"

	// RouterKey is the message router key for the liquidstaking module
	RouterKey = ModuleName

	// StoreKey is the default store key for the liquidstaking module
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the liquidstaking module
	QuerierRoute = ModuleName
)

var (
	// Keys for store prefixes
	LiquidValidatorsKey = []byte{0x01} // prefix for each key to a liquid validator
)

// GetLiquidValidatorKey creates the key for the liquid validator with address
// VALUE: liquidstaking/LiquidValidator
func GetLiquidValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(LiquidValidatorsKey, address.MustLengthPrefix(operatorAddr)...)
}

//// GetTotalCollectedCoinsKey creates the key for the total collected coins for a liquidstaking.
//func GetTotalCollectedCoinsKey(liquidStakingName string) []byte {
//	return append(TotalCollectedCoinsKeyPrefix, []byte(liquidStakingName)...)
//}
//
//// ParseTotalCollectedCoinsKey parses the total collected coins key and returns the liquidstaking name.
//func ParseTotalCollectedCoinsKey(key []byte) (liquidStakingName string) {
//	if !bytes.HasPrefix(key, TotalCollectedCoinsKeyPrefix) {
//		panic("key does not have proper prefix")
//	}
//	return string(key[1:])
//}
