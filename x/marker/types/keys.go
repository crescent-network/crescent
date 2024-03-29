package types

const (
	// ModuleName defines the module name
	ModuleName = "marker"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	LastBlockTimeKey = []byte{0xff}
)
