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
	LastMarketIdKey              = []byte{0x01}
	LastOrderIdKey               = []byte{0x02}
	MarketKeyPrefix              = []byte{0x03}
	MarketStateKeyPrefix         = []byte{0x04}
	MarketByDenomsIndexKeyPrefix = []byte{0x05}
	OrderKeyPrefix               = []byte{0x06}
	OrderBookOrderKeyPrefix      = []byte{0x07}
)

func GetMarketKey(marketId uint64) []byte {
	return utils.Key(MarketKeyPrefix, sdk.Uint64ToBigEndian(marketId))
}

func GetMarketStateKey(marketId uint64) []byte {
	return utils.Key(MarketStateKeyPrefix, sdk.Uint64ToBigEndian(marketId))
}

func GetMarketByDenomsIndexKey(baseDenom, quoteDenom string) []byte {
	return utils.Key(
		MarketByDenomsIndexKeyPrefix,
		utils.LengthPrefixString(baseDenom),
		[]byte(quoteDenom))
}

func GetOrderKey(orderId uint64) []byte {
	return utils.Key(OrderKeyPrefix, sdk.Uint64ToBigEndian(orderId))
}

func GetOrderBookOrderKey(marketId uint64, isBuy bool, price sdk.Dec, orderId uint64) []byte {
	var orderIdBytes []byte
	if isBuy {
		orderIdBytes = sdk.Uint64ToBigEndian(-orderId)
	} else {
		orderIdBytes = sdk.Uint64ToBigEndian(orderId)
	}
	return utils.Key(
		OrderBookOrderKeyPrefix,
		sdk.Uint64ToBigEndian(marketId),
		isBuyToBytes(isBuy),
		sdk.SortableDecBytes(price),
		orderIdBytes)
}

func GetOrderBookIteratorPrefix(marketId uint64, isBuy bool) []byte {
	return utils.Key(
		OrderBookOrderKeyPrefix,
		sdk.Uint64ToBigEndian(marketId),
		isBuyToBytes(isBuy))
}

func ParseMarketByDenomsIndexKey(key []byte) (baseDenom, quoteDenom string) {
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
