package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DeriveMarketId(baseDenom, quoteDenom string) string {
	s := fmt.Sprintf("spot/market/%s/%s", baseDenom, quoteDenom)
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func NewSpotMarket(baseDenom, quoteDenom string) SpotMarket {
	marketId := DeriveMarketId(baseDenom, quoteDenom)
	return SpotMarket{
		Id:         marketId,
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
		LastPrice:  nil,
	}
}

func (market SpotMarket) OfferCoin(isBuy bool, price sdk.Dec, qty sdk.Int) sdk.Coin {
	offerAmt := OfferAmount(isBuy, price, qty)
	if isBuy {
		return sdk.NewCoin(market.QuoteDenom, offerAmt)
	}
	return sdk.NewCoin(market.BaseDenom, offerAmt)
}
