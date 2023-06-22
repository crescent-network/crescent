package types

import (
	"bytes"
)

const (
	// ModuleName is the name of the budget module
	ModuleName = "budget"

	// RouterKey is the message router key for the budget module
	RouterKey = ModuleName

	// StoreKey is the default store key for the budget module
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the budget module
	QuerierRoute = ModuleName
)

var (
	// Keys for store prefixes
	TotalCollectedCoinsKeyPrefix = []byte{0x11}
)

// GetTotalCollectedCoinsKey creates the key for the total collected coins for a budget.
func GetTotalCollectedCoinsKey(budgetName string) []byte {
	return append(TotalCollectedCoinsKeyPrefix, []byte(budgetName)...)
}

// ParseTotalCollectedCoinsKey parses the total collected coins key and returns the budget name.
func ParseTotalCollectedCoinsKey(key []byte) (budgetName string) {
	if !bytes.HasPrefix(key, TotalCollectedCoinsKeyPrefix) {
		panic("key does not have proper prefix")
	}
	return string(key[1:])
}
