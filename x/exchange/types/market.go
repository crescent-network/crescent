package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func DeriveMarketEscrowAddress(marketId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("MarketEscrowAddress/%d", marketId)))
}

func NewMarket(
	marketId uint64, baseDenom, quoteDenom string, makerFeeRate, takerFeeRate, orderSourceFeeRatio sdk.Dec) Market {
	return Market{
		Id:                  marketId,
		BaseDenom:           baseDenom,
		QuoteDenom:          quoteDenom,
		EscrowAddress:       DeriveMarketEscrowAddress(marketId).String(),
		MakerFeeRate:        makerFeeRate,
		TakerFeeRate:        takerFeeRate,
		OrderSourceFeeRatio: orderSourceFeeRatio,
	}
}

func (market Market) Validate() error {
	if market.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if err := sdk.ValidateDenom(market.BaseDenom); err != nil {
		return fmt.Errorf("invalid base denom: %w", err)
	}
	if err := sdk.ValidateDenom(market.QuoteDenom); err != nil {
		return fmt.Errorf("invalid quote denom: %w", err)
	}
	if market.BaseDenom == market.QuoteDenom {
		return fmt.Errorf("base denom and quote denom must not be same: %s", market.BaseDenom)
	}
	if _, err := sdk.AccAddressFromBech32(market.EscrowAddress); err != nil {
		return fmt.Errorf("invalid escrow address: %w", err)
	}
	if err := ValidateFees(
		market.MakerFeeRate, market.TakerFeeRate, market.OrderSourceFeeRatio); err != nil {
		return err
	}
	return nil
}

func (market Market) MustGetEscrowAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(market.EscrowAddress)
}

func NewMarketState(lastPrice *sdk.Dec) MarketState {
	return MarketState{
		LastPrice:          lastPrice,
		LastMatchingHeight: -1, // Not matched
	}
}

func (marketState MarketState) Validate() error {
	if marketState.LastPrice != nil {
		if !marketState.LastPrice.IsPositive() {
			return fmt.Errorf("last price must be positive: %s", marketState.LastPrice)
		}
		if _, valid := ValidateTickPrice(*marketState.LastPrice); !valid {
			return fmt.Errorf("invalid last price tick: %s", marketState.LastPrice)
		}
	}
	if marketState.LastMatchingHeight < -1 {
		return fmt.Errorf("invalid last matching height: %d", marketState.LastMatchingHeight)
	}
	if marketState.LastPrice != nil && marketState.LastMatchingHeight == -1 ||
		marketState.LastPrice == nil && marketState.LastMatchingHeight >= 0 {
		return fmt.Errorf(
			"inconsistent last matching info: %s, %d",
			marketState.LastPrice, marketState.LastMatchingHeight)
	}
	return nil
}
