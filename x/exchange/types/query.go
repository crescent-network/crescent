package types

func NewMarketResponse(market Market, marketState MarketState) MarketResponse {
	return MarketResponse{
		Id:                  market.Id,
		BaseDenom:           market.BaseDenom,
		QuoteDenom:          market.QuoteDenom,
		EscrowAddress:       market.EscrowAddress,
		MakerFeeRate:        market.MakerFeeRate,
		TakerFeeRate:        market.TakerFeeRate,
		OrderSourceFeeRatio: market.OrderSourceFeeRatio,
		LastPrice:           marketState.LastPrice,
		LastMatchingHeight:  marketState.LastMatchingHeight,
	}
}
