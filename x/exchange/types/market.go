package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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
	fees Fees, orderQtyLimits, orderQuoteLimits AmountLimits) Market {
	return Market{
		Id:                  marketId,
		BaseDenom:           baseDenom,
		QuoteDenom:          quoteDenom,
		EscrowAddress:       DeriveMarketEscrowAddress(marketId).String(),
		FeeCollector:        DeriveFeeCollector(marketId).String(),
		Fees:                fees,
		OrderQuantityLimits: orderQtyLimits,
		OrderQuoteLimits:    orderQuoteLimits,
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
	if err := market.Fees.Validate(); err != nil {
		return err
	}
	if err := market.OrderQuantityLimits.Validate(); err != nil {
		return fmt.Errorf("invalid order quantity limits: %w", err)
	}
	if err := market.OrderQuoteLimits.Validate(); err != nil {
		return fmt.Errorf("invalid order quote limits: %w", err)
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
			feeRate = market.Fees.MakerFeeRate
		} else {
			feeRate = market.Fees.TakerFeeRate
		}
	} else { // order source order
		if isMaker {
			feeRate = market.Fees.TakerFeeRate.Neg().Mul(market.Fees.OrderSourceFeeRatio)
		} else {
			feeRate = utils.ZeroDec
		}
	}
	if halveFees {
		feeRate = feeRate.QuoInt64(2)
	}
	return feeRate
}

func (market Market) CheckOrderQuantityLimits(qty sdk.Int) error {
	if qty.LT(market.OrderQuantityLimits.Min) {
		return sdkerrors.Wrapf(
			ErrBadOrderAmount,
			"quantity is less than the minimum order quantity allowed: %s < %s",
			qty, market.OrderQuantityLimits.Min)
	}
	if qty.GT(market.OrderQuantityLimits.Max) {
		return sdkerrors.Wrapf(
			ErrBadOrderAmount,
			"quantity is greater than the maximum order quantity allowed: %s > %s",
			qty, market.OrderQuantityLimits.Max)
	}
	return nil
}

func (market Market) CheckOrderQuoteLimits(price sdk.Dec, qty sdk.Int) error {
	if quote := price.MulInt(qty).TruncateInt(); quote.LT(market.OrderQuoteLimits.Min) {
		return sdkerrors.Wrapf(
			ErrBadOrderAmount,
			"quote(=price*qty) is less than the minimum order quote allowed: %s < %s",
			quote, market.OrderQuoteLimits.Min)
	}
	if quote := price.MulInt(qty).Ceil().TruncateInt(); quote.GT(market.OrderQuoteLimits.Max) {
		return sdkerrors.Wrapf(
			ErrBadOrderAmount,
			"quote(=price*qty) is greater than the maximum order quote allowed: %s > %s",
			quote, market.OrderQuoteLimits.Max)
	}
	return nil
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
