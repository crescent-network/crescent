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
	LastOrderIdKey              = []byte{0x01}
	SpotMarketKeyPrefix         = []byte{0x02}
	SpotMarketStateKeyPrefix    = []byte{0x03}
	SpotOrderKeyPrefix          = []byte{0x04}
	SpotOrderBookOrderKeyPrefix = []byte{0x05}
)

func GetSpotMarketKey(marketId string) []byte {
	return utils.Key(SpotMarketKeyPrefix, utils.LengthPrefixString(marketId))
}

func GetSpotMarketStateKey(marketId string) []byte {
	return utils.Key(SpotMarketStateKeyPrefix, utils.LengthPrefixString(marketId))
}

func GetSpotOrderKey(orderId uint64) []byte {
	return utils.Key(SpotOrderKeyPrefix, sdk.Uint64ToBigEndian(orderId))
}

func GetSpotOrderBookOrderKey(marketId string, isBuy bool, price sdk.Dec, orderId uint64) []byte {
	var orderIdBytes []byte
	if isBuy {
		orderIdBytes = sdk.Uint64ToBigEndian(-orderId)
	} else {
		orderIdBytes = sdk.Uint64ToBigEndian(orderId)
	}
	return utils.Key(
		SpotOrderBookOrderKeyPrefix,
		utils.LengthPrefixString(marketId),
		isBuyToBytes(isBuy),
		sdk.SortableDecBytes(price),
		orderIdBytes)
}

func GetSpotOrderBookIteratorPrefix(marketId string, isBuy bool) []byte {
	return utils.Key(
		SpotOrderBookOrderKeyPrefix,
		utils.LengthPrefixString(marketId),
		isBuyToBytes(isBuy))
}

func GetSpotOrderBookIteratorEndBytes(marketId string, isBuy bool, price sdk.Dec) []byte {
	prefix := utils.Key(
		SpotOrderBookOrderKeyPrefix,
		utils.LengthPrefixString(marketId),
		isBuyToBytes(isBuy),
		sdk.SortableDecBytes(price))
	if isBuy {
		return prefix
	}
	return sdk.PrefixEndBytes(prefix)
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
