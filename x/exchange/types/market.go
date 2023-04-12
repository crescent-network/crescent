package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func DeriveMarketId(baseDenom, quoteDenom string) string {
	s := fmt.Sprintf("spot/market/%s/%s", baseDenom, quoteDenom)
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func DeriveEscrowAddress(marketId string) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("SpotMarketEscrowAddress/%s", marketId)))
}

func NewSpotMarket(baseDenom, quoteDenom string) SpotMarket {
	marketId := DeriveMarketId(baseDenom, quoteDenom)
	return SpotMarket{
		Id:            marketId,
		BaseDenom:     baseDenom,
		QuoteDenom:    quoteDenom,
		EscrowAddress: DeriveEscrowAddress(marketId).String(),
	}
}

func NewSpotMarketState(lastPrice *sdk.Dec) SpotMarketState {
	return SpotMarketState{
		LastPrice: lastPrice,
	}
}

func (market SpotMarket) DepositCoin(isBuy bool, amt sdk.Int) sdk.Coin {
	if isBuy {
		return sdk.NewCoin(market.QuoteDenom, amt)
	}
	return sdk.NewCoin(market.BaseDenom, amt)
}
