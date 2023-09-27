package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v5/types"
)

func DeriveMarketEscrowAddress(marketId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("MarketEscrowAddress/%d", marketId)))
}

func DeriveFeeCollector(marketId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("FeeCollector/%d", marketId)))
}

func NewMarket(
	marketId uint64, baseDenom, quoteDenom string,
	makerFeeRate, takerFeeRate, orderSourceFeeRatio sdk.Dec,
	minOrderQty, minOrderQuote, maxOrderQty, maxOrderQuote sdk.Int) Market {
	return Market{
		Id:                  marketId,
		BaseDenom:           baseDenom,
		QuoteDenom:          quoteDenom,
		EscrowAddress:       DeriveMarketEscrowAddress(marketId).String(),
		FeeCollector:        DeriveFeeCollector(marketId).String(),
		MakerFeeRate:        makerFeeRate,
		TakerFeeRate:        takerFeeRate,
		OrderSourceFeeRatio: orderSourceFeeRatio,
		MinOrderQuantity:    minOrderQty,
		MinOrderQuote:       minOrderQuote,
		MaxOrderQuantity:    maxOrderQty,
		MaxOrderQuote:       maxOrderQuote,
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
	if _, err := sdk.AccAddressFromBech32(market.FeeCollector); err != nil {
		return fmt.Errorf("invalid fee collector: %w", err)
	}
	if err := ValidateFees(
		market.MakerFeeRate, market.TakerFeeRate, market.OrderSourceFeeRatio); err != nil {
		return err
	}
	if market.MinOrderQuantity.IsNegative() {
		return fmt.Errorf("min order quantity must not be negative: %s", market.MinOrderQuantity)
	}
	if market.MinOrderQuote.IsNegative() {
		return fmt.Errorf("min order quote must not be negative: %s", market.MinOrderQuote)
	}
	if market.MaxOrderQuantity.IsNegative() {
		return fmt.Errorf("max order quantity must not be negative: %s", market.MaxOrderQuantity)
	}
	if market.MaxOrderQuote.IsNegative() {
		return fmt.Errorf("max order quote must not be negative: %s", market.MaxOrderQuote)
	}
	return nil
}

func (market Market) MustGetEscrowAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(market.EscrowAddress)
}

func (market Market) MustGetFeeCollectorAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(market.FeeCollector)
}

func (market Market) FeeRate(isOrderSourceOrder, isMaker, halveFees bool) (feeRate sdk.Dec) {
	if !isOrderSourceOrder { // user order
		if isMaker {
			feeRate = market.MakerFeeRate
		} else {
			feeRate = market.TakerFeeRate
		}
	} else { // order source order
		if isMaker {
			feeRate = market.TakerFeeRate.Neg().Mul(market.OrderSourceFeeRatio)
		} else {
			feeRate = utils.ZeroDec
		}
	}
	if halveFees {
		feeRate = feeRate.QuoInt64(2)
	}
	return feeRate
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

func OrderPriceLimit(basePrice, maxOrderPriceRatio sdk.Dec) (minPrice, maxPrice sdk.Dec) {
	// Manually round up the min tick.
	minTick, valid := ValidateTickPrice(basePrice.Mul(utils.OneDec.Sub(maxOrderPriceRatio)))
	if !valid {
		minTick++
	}
	minPrice = PriceAtTick(minTick)
	// TickAtPrice automatically rounds down the tick.
	maxPrice = PriceAtTick(
		TickAtPrice(basePrice.Mul(utils.OneDec.Add(maxOrderPriceRatio))))
	return
}
