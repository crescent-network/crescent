package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "exchange"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	TStoreKey = "transient_exchange"

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// TODO: reallocate key prefix bytes
var (
	LastSpotMarketIdKey              = []byte{0x01}
	LastSpotOrderIdKey               = []byte{0x02}
	SpotMarketKeyPrefix              = []byte{0x03}
	SpotMarketStateKeyPrefix         = []byte{0x04}
	SpotMarketByDenomsIndexKeyPrefix = []byte{0x05}
	SpotOrderKeyPrefix               = []byte{0x06}
	SpotOrderBookOrderKeyPrefix      = []byte{0x07}
)

func GetSpotMarketKey(marketId uint64) []byte {
	return utils.Key(SpotMarketKeyPrefix, sdk.Uint64ToBigEndian(marketId))
}

func GetSpotMarketStateKey(marketId uint64) []byte {
	return utils.Key(SpotMarketStateKeyPrefix, sdk.Uint64ToBigEndian(marketId))
}

func GetSpotMarketByDenomsIndexKey(baseDenom, quoteDenom string) []byte {
	return utils.Key(
		SpotMarketByDenomsIndexKeyPrefix,
		utils.LengthPrefixString(baseDenom),
		[]byte(quoteDenom))
}

func GetSpotOrderKey(orderId uint64) []byte {
	return utils.Key(SpotOrderKeyPrefix, sdk.Uint64ToBigEndian(orderId))
}

func GetSpotOrderBookOrderKey(marketId uint64, isBuy bool, price sdk.Dec, orderId uint64) []byte {
	var orderIdBytes []byte
	if isBuy {
		orderIdBytes = sdk.Uint64ToBigEndian(-orderId)
	} else {
		orderIdBytes = sdk.Uint64ToBigEndian(orderId)
	}
	return utils.Key(
		SpotOrderBookOrderKeyPrefix,
		sdk.Uint64ToBigEndian(marketId),
		isBuyToBytes(isBuy),
		sdk.SortableDecBytes(price),
		orderIdBytes)
}

func GetSpotOrderBookIteratorPrefix(marketId uint64, isBuy bool) []byte {
	return utils.Key(
		SpotOrderBookOrderKeyPrefix,
		sdk.Uint64ToBigEndian(marketId),
		isBuyToBytes(isBuy))
}

func ParseSpotMarketByDenomsIndexKey(key []byte) (baseDenom, quoteDenom string) {
	baseDenomLen := key[1]
	baseDenom = string(key[2 : 2+baseDenomLen])
	quoteDenom = string(key[2+baseDenomLen:])
	return
}

var (
	buyBytes  = []byte{0}
	sellBytes = []byte{1}
)

func isBuyToBytes(isBuy bool) []byte {
	if isBuy {
		return buyBytes
	}
	return sellBytes
}
