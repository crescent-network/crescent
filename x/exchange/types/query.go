package types

func NewMarketResponse(market Market, marketState MarketState) MarketResponse {
	return MarketResponse{
		Id:                  market.Id,
		BaseDenom:           market.BaseDenom,
		QuoteDenom:          market.QuoteDenom,
		EscrowAddress:       market.EscrowAddress,
		FeeCollector:        market.FeeCollector,
		MakerFeeRate:        market.MakerFeeRate,
		TakerFeeRate:        market.TakerFeeRate,
		OrderSourceFeeRatio: market.OrderSourceFeeRatio,
		MinOrderQuantity:    market.MinOrderQuantity,
		MinOrderQuote:       market.MinOrderQuote,
		MaxOrderQuantity:    market.MaxOrderQuantity,
		MaxOrderQuote:       market.MaxOrderQuote,
		LastPrice:           marketState.LastPrice,
		LastMatchingHeight:  marketState.LastMatchingHeight,
	}
}
