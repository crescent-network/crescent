package types

func NewSpotMarketResponse(market SpotMarket, marketState SpotMarketState) SpotMarketResponse {
	return SpotMarketResponse{
		Id:            market.Id,
		BaseDenom:     market.BaseDenom,
		QuoteDenom:    market.QuoteDenom,
		EscrowAddress: market.EscrowAddress,
		LastPrice:     marketState.LastPrice,
	}
}
