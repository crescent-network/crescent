package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

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

var (
	LastMarketIdKey              = []byte{0x60}
	LastOrderIdKey               = []byte{0x61}
	MarketKeyPrefix              = []byte{0x62}
	MarketStateKeyPrefix         = []byte{0x63}
	MarketByDenomsIndexKeyPrefix = []byte{0x64}
	OrderKeyPrefix               = []byte{0x65}
	OrderBookOrderKeyPrefix      = []byte{0x66}
	NumMMOrdersKeyPrefix         = []byte{0x67}
	TransientBalanceKeyPrefix    = []byte{0x68}
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
	return utils.Key(
		OrderBookOrderKeyPrefix,
		sdk.Uint64ToBigEndian(marketId),
		isBuyToBytes(isBuy),
		sdk.SortableDecBytes(price),
		sdk.Uint64ToBigEndian(orderId))
}

func GetOrderIdsByMarketKey(marketId uint64) []byte {
	return utils.Key(OrderBookOrderKeyPrefix, sdk.Uint64ToBigEndian(marketId))
}

func GetOrderBookIteratorPrefix(marketId uint64, isBuy bool) []byte {
	return utils.Key(
		OrderBookOrderKeyPrefix,
		sdk.Uint64ToBigEndian(marketId),
		isBuyToBytes(isBuy))
}

func GetNumMMOrdersKey(ordererAddr sdk.AccAddress, marketId uint64) []byte {
	return utils.Key(
		NumMMOrdersKeyPrefix,
		address.MustLengthPrefix(ordererAddr),
		sdk.Uint64ToBigEndian(marketId))
}

func GetTransientBalanceKey(addr sdk.AccAddress, denom string) []byte {
	return utils.Key(
		TransientBalanceKeyPrefix,
		address.MustLengthPrefix(addr),
		[]byte(denom))
}

func ParseMarketByDenomsIndexKey(key []byte) (baseDenom, quoteDenom string) {
	baseDenomLen := key[1]
	baseDenom = string(key[2 : 2+baseDenomLen])
	quoteDenom = string(key[2+baseDenomLen:])
	return
}

func ParseTransientBalanceKey(key []byte) (addr sdk.AccAddress, denom string) {
	addrLen := key[1]
	addr = key[2 : 2+addrLen]
	denom = string(key[2+addrLen:])
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
