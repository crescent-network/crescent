package types

// LastBlockTimeKey is the key to use for the keeper store.
var LastBlockTimeKey = []byte{0x90}

const (
	// module name
	ModuleName = "mint"

	// StoreKey is the default store key for mint
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the mint store.
	QuerierRoute = StoreKey

	// Query endpoints supported by the mint querier
	QueryParameters = "parameters"
)
