package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	OffloadedOrderKeyPrefix = []byte{0xba}
)

func GetOffloadedOrderKey(blockHeight int64, pairId, orderId uint64) (key []byte) {
	key = append(OffloadedOrderKeyPrefix, sdk.Uint64ToBigEndian(uint64(blockHeight))...)
	key = append(key, sdk.Uint64ToBigEndian(pairId)...)
	key = append(key, sdk.Uint64ToBigEndian(orderId)...)
	return
}
