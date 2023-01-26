package bindings

// CrescentQuery contains custom queries that can be called from a contract.
type CrescentQuery struct {
	Pairs *Pairs `json:"pairs,omitempty"`
	Pair  *Pair  `json:"pair,omitempty"`
}

type Pairs struct{}

type Pair struct {
	Id uint64 `json:"id"`
}

type PairsResponse struct {
	Pairs []PairResponse `json:"pairs"`
}

type PairResponse struct {
	Id             uint64 `json:"id"`
	BaseCoinDenom  string `json:"base_coin_denom"`
	QuoteCoinDenom string `json:"quote_coin_denom"`
	EscrowAddress  string `json:"escrow_address"`
	// TODO: test
}
