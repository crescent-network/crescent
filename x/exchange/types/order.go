package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DeriveSpotLimitOrderId(marketId string, seq uint64) string {
	s := fmt.Sprintf("spot/order/%s/%d", marketId, seq)
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func NewSpotLimitOrder(
	ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, price sdk.Dec, qty sdk.Int, seq uint64) SpotLimitOrder {
	orderId := DeriveSpotLimitOrderId(marketId, seq)
	return SpotLimitOrder{
		Id:           orderId,
		Orderer:      ordererAddr.String(),
		MarketId:     marketId,
		IsBuy:        isBuy,
		Price:        price,
		Quantity:     qty,
		OpenQuantity: qty,
		Sequence:     seq,
	}
}
