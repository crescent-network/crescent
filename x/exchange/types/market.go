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

func NewMarket(
	marketId uint64, baseDenom, quoteDenom string, makerFeeRate, takerFeeRate sdk.Dec) Market {
	return Market{
		Id:            marketId,
		BaseDenom:     baseDenom,
		QuoteDenom:    quoteDenom,
		EscrowAddress: DeriveMarketEscrowAddress(marketId).String(),
		MakerFeeRate:  makerFeeRate,
		TakerFeeRate:  takerFeeRate,
	}
}

func (market Market) DepositCoin(isBuy bool, amt sdk.Int) sdk.Coin {
	if isBuy {
		return sdk.NewCoin(market.QuoteDenom, amt)
	}
	return sdk.NewCoin(market.BaseDenom, amt)
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
	if _, err := sdk.AccAddressFromBech32(market.EscrowAddress); err != nil {
		return fmt.Errorf("invalid escrow address: %w", err)
	}
	if market.TakerFeeRate.IsNegative() {
		return fmt.Errorf("taker fee rate must not be negative: %s", market.TakerFeeRate)
	}
	if market.TakerFeeRate.GT(utils.OneDec) {
		return fmt.Errorf("taker fee rate must not exceed 1.0: %s", market.TakerFeeRate)
	}
	if market.MakerFeeRate.IsNegative() {
		negMakerFeeRate := market.MakerFeeRate.Neg()
		if negMakerFeeRate.GT(utils.OneDec) {
			return fmt.Errorf("minus maker fee rate must not exceed 1.0: %s", market.MakerFeeRate)
		}
		if market.TakerFeeRate.LT(negMakerFeeRate) {
			return fmt.Errorf("minus maker fee rate must not exceed %s", market.TakerFeeRate)
		}
	} else if market.MakerFeeRate.GT(utils.OneDec) {
		return fmt.Errorf("maker fee rate must not exceed 1.0:% s", market.MakerFeeRate)
	}
	return nil
}

func (market Market) DeductTakerFee(amt sdk.Int, halveFee bool) sdk.Int {
	takerFeeRate := market.TakerFeeRate
	if halveFee {
		takerFeeRate = takerFeeRate.QuoInt64(2)
	}
	return utils.OneDec.Sub(takerFeeRate).MulInt(amt).TruncateInt()
}

func (market Market) PayDenom(isBuy bool) string {
	if isBuy {
		return market.QuoteDenom
	}
	return market.BaseDenom
}

func (market Market) ReceiveDenom(isBuy bool) string {
	if isBuy {
		return market.BaseDenom
	}
	return market.QuoteDenom
}

func NewMarketState(lastPrice *sdk.Dec) MarketState {
	return MarketState{
		LastPrice: lastPrice,
	}
}

func (marketState MarketState) Validate() error {
	if marketState.LastPrice != nil {
		if !marketState.LastPrice.IsPositive() {
			return fmt.Errorf("last price must be positive: %s", marketState.LastPrice)
		}
	}
	return nil
}
