package types

func NewMarketResponse(market Market, marketState MarketState) MarketResponse {
	return MarketResponse{
		Id:                  market.Id,
		BaseDenom:           market.BaseDenom,
		QuoteDenom:          market.QuoteDenom,
		EscrowAddress:       market.EscrowAddress,
		FeeCollector:        market.FeeCollector,
		Fees:                market.Fees,
		OrderQuantityLimits: market.OrderQuantityLimits,
		OrderQuoteLimits:    market.OrderQuoteLimits,
		LastPrice:           marketState.LastPrice,
		LastMatchingHeight:  marketState.LastMatchingHeight,
	}
}
