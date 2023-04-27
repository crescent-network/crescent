package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func CalculateAmountsAndFee(
	market Market, isTakerBuy bool, execQty, execQuote sdk.Int, isTemporaryMaker bool) (makerPays, makerReceives, makerFee, takerPays, takerReceives sdk.Coin) {
	negativeMakerFee := market.MakerFeeRate.IsNegative()
	if isTakerBuy {
		makerPays = sdk.NewCoin(market.BaseDenom, execQty)
		makerReceives = sdk.NewCoin(market.QuoteDenom, execQuote)
		if !negativeMakerFee {
			if isTemporaryMaker { // Don't charge maker fee to temporary(module) market makers
				makerFee = sdk.NewCoin(market.QuoteDenom, utils.ZeroInt)
			} else {
				makerFee = sdk.NewCoin(market.QuoteDenom, market.MakerFeeRate.MulInt(execQuote).Ceil().TruncateInt())
			}
		} else {
			// Don't use sdk.NewCoin here because the amount is negative
			makerFee = sdk.Coin{Denom: market.BaseDenom, Amount: market.MakerFeeRate.MulInt(execQty).TruncateInt()}
		}
		takerPays = sdk.NewCoin(market.QuoteDenom, execQuote)
		takerReceives = sdk.NewCoin(
			market.BaseDenom,
			utils.OneDec.Sub(market.TakerFeeRate).MulInt(execQty).TruncateInt())
	} else {
		makerPays = sdk.NewCoin(market.QuoteDenom, execQuote)
		makerReceives = sdk.NewCoin(market.BaseDenom, execQty)
		if !negativeMakerFee {
			if isTemporaryMaker {
				makerFee = sdk.NewCoin(market.BaseDenom, utils.ZeroInt)
			} else {
				makerFee = sdk.NewCoin(market.BaseDenom, market.MakerFeeRate.MulInt(execQty).Ceil().TruncateInt())
			}
		} else {
			makerFee = sdk.Coin{Denom: market.QuoteDenom, Amount: market.MakerFeeRate.MulInt(execQuote).TruncateInt()}
		}
		takerPays = sdk.NewCoin(market.BaseDenom, execQty)
		takerReceives = sdk.NewCoin(
			market.QuoteDenom,
			utils.OneDec.Sub(market.TakerFeeRate).MulInt(execQuote).TruncateInt())
	}
	return
}
