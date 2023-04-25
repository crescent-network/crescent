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

func (market SpotMarket) DepositCoin(isBuy bool, amt sdk.Int) sdk.Coin {
	if isBuy {
		return sdk.NewCoin(market.QuoteDenom, amt)
	}
	return sdk.NewCoin(market.BaseDenom, amt)
}

func (market SpotMarket) Validate() error {
	if market.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if err := sdk.ValidateDenom(market.BaseDenom); err != nil {
		return fmt.Errorf("invalid base denom: %w", err)
	}
	if err := sdk.ValidateDenom(market.QuoteDenom); err != nil {
		return fmt.Errorf("invalid quote denom: %w", err)
	}
	if _, err := sdk.AccAddressFromBech32(market.EscrowAddress); err != nil {
		return fmt.Errorf("invalid escrow address: %w", err)
	}
	return nil
}

func NewSpotMarketState(lastPrice *sdk.Dec) SpotMarketState {
	return SpotMarketState{
		LastPrice: lastPrice,
	}
}

func (marketState SpotMarketState) Validate() error {
	if marketState.LastPrice != nil {
		if !marketState.LastPrice.IsPositive() {
			return fmt.Errorf("last price must be positive: %s", marketState.LastPrice)
		}
	}
	return nil
}
