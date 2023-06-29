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
	if market.BaseDenom == market.QuoteDenom {
		return fmt.Errorf("base denom and quote denom must not be same: %s", market.BaseDenom)
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

func (market Market) DeductTakerFee(amt sdk.Int, halveFee bool) (deducted, fee sdk.Int) {
	takerFeeRate := market.TakerFeeRate
	if halveFee {
		takerFeeRate = takerFeeRate.QuoInt64(2)
	}
	deducted = utils.OneDec.Sub(takerFeeRate).MulInt(amt).TruncateInt()
	fee = amt.Sub(deducted)
	return
}

func (market Market) PayReceiveDenoms(isBuy bool) (payDenom, receiveDenom string) {
	if isBuy {
		return market.QuoteDenom, market.BaseDenom
	}
	return market.BaseDenom, market.QuoteDenom
}

func (market Market) MustGetEscrowAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(market.EscrowAddress)
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
		if _, valid := ValidateTickPrice(*marketState.LastPrice); !valid {
			return fmt.Errorf("invalid last price tick: %s", marketState.LastPrice)
		}
	}
	return nil
}
