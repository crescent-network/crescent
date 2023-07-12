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
	LastMarketIdKey               = []byte{0x60}
	LastOrderIdKey                = []byte{0x61}
	MarketKeyPrefix               = []byte{0x62}
	MarketStateKeyPrefix          = []byte{0x63}
	MarketByDenomsIndexKeyPrefix  = []byte{0x64}
	OrderKeyPrefix                = []byte{0x65}
	OrderBookOrderIndexKeyPrefix  = []byte{0x66}
	OrdersByOrdererIndexKeyPrefix = []byte{0x67}
	NumMMOrdersKeyPrefix          = []byte{0x68}
	TransientBalanceKeyPrefix     = []byte{0x69}
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

func GetOrderBookOrderIndexKey(marketId uint64, isBuy bool, price sdk.Dec, orderId uint64) []byte {
	return utils.Key(
		OrderBookOrderIndexKeyPrefix,
		sdk.Uint64ToBigEndian(marketId),
		isBuyToBytes(isBuy),
		PriceToBytes(price),
		sdk.Uint64ToBigEndian(orderId))
}

func GetOrderBookSideIteratorPrefix(marketId uint64, isBuy bool) []byte {
	return utils.Key(
		OrderBookOrderIndexKeyPrefix,
		sdk.Uint64ToBigEndian(marketId),
		isBuyToBytes(isBuy))
}

func GetOrdersByMarketIteratorPrefix(marketId uint64) []byte {
	return utils.Key(
		OrderBookOrderIndexKeyPrefix,
		sdk.Uint64ToBigEndian(marketId))
}

func GetOrdersByOrdererIndexKey(ordererAddr sdk.AccAddress, marketId, orderId uint64) []byte {
	return utils.Key(
		OrdersByOrdererIndexKeyPrefix,
		address.MustLengthPrefix(ordererAddr),
		sdk.Uint64ToBigEndian(marketId),
		sdk.Uint64ToBigEndian(orderId))
}

func GetOrdersByOrdererIteratorPrefix(ordererAddr sdk.AccAddress) []byte {
	return utils.Key(OrdersByOrdererIndexKeyPrefix, address.MustLengthPrefix(ordererAddr))
}

func GetOrdersByOrdererAndMarketIteratorPrefix(ordererAddr sdk.AccAddress, marketId uint64) []byte {
	return utils.Key(
		OrdersByOrdererIndexKeyPrefix,
		address.MustLengthPrefix(ordererAddr),
		sdk.Uint64ToBigEndian(marketId))
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

func ParseOrderIdFromOrderBookOrderIndexKey(key []byte) (orderId uint64) {
	orderId = sdk.BigEndianToUint64(key[1+1+32+8:])
	return
}

func ParseOrderIdFromOrdersByOrdererIndexKey(key []byte) (orderId uint64) {
	addrLen := key[1]
	orderId = sdk.BigEndianToUint64(key[2+addrLen+8:])
	return
}

func ParseNumMMOrdersKey(key []byte) (ordererAddr sdk.AccAddress, marketId uint64) {
	addrLen := key[1]
	ordererAddr = key[2 : 2+addrLen]
	marketId = sdk.BigEndianToUint64(key[2+addrLen:])
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
