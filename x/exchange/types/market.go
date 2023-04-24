package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func DeriveEscrowAddress(marketId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("SpotMarketEscrowAddress/%d", marketId)))
}

func NewSpotMarket(marketId uint64, baseDenom, quoteDenom string) SpotMarket {
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
