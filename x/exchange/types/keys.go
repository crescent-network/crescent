package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "exchange"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	OrderSequenceKey            = []byte{0x01}
	SpotMarketKeyPrefix         = []byte{0x02}
	SpotLimitOrderKeyPrefix     = []byte{0x03}
	SpotOrderBookOrderKeyPrefix = []byte{0x04}
)

func GetSpotMarketKey(marketId string) []byte {
	return append(SpotMarketKeyPrefix, marketId...)
}

func GetSpotLimitOrderKey(marketId, orderId string) []byte {
	return append(append(SpotLimitOrderKeyPrefix, marketId...), orderId...)
}

func GetSpotOrderBookOrderKey(marketId string, isBuy bool, price sdk.Dec, seq uint64) (key []byte) {
	key = append(SpotOrderBookOrderKeyPrefix, marketId...)
	key = append(key, boolToByte(isBuy))
	key = append(key, sdk.SortableDecBytes(price)...)
	key = append(key, sdk.Uint64ToBigEndian(seq)...)
	return
}

func GetSpotOrderBookIteratorPrefix(marketId string, isBuy bool) []byte {
	return append(append(SpotOrderBookOrderKeyPrefix, marketId...), boolToByte(isBuy))
}

func GetSpotOrderBookIteratorEndBytes(marketId string, isBuy bool, price sdk.Dec) []byte {
	prefix := append(SpotOrderBookOrderKeyPrefix, marketId...)
	prefix = append(prefix, boolToByte(isBuy))
	prefix = append(prefix, sdk.SortableDecBytes(price)...)
	if isBuy {
		return prefix
	}
	return sdk.PrefixEndBytes(prefix)
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
