package bindings

import (
	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
)

// CrescentQuery contains custom queries that can be called from a contract.
type CrescentQuery struct {
	Pairs *Pairs `json:"pairs,omitempty"`
	Pair  *Pair  `json:"pair,omitempty"`
	Pools *Pools `json:"pools,omitempty"`
	Pool  *Pool  `json:"pool,omitempty"`
}

type Pairs struct{}

type Pair struct {
	PairId uint64 `json:"pair_id"`
}

type Pools struct{}

type Pool struct {
	PoolId uint64 `json:"pool_id"`
}

type PairsResponse struct {
	Pairs []PairResponse `json:"pairs"`
}

type PairResponse struct {
	Id             uint64 `json:"id"`
	BaseCoinDenom  string `json:"base_coin_denom"`
	QuoteCoinDenom string `json:"quote_coin_denom"`
	EscrowAddress  string `json:"escrow_address"`
}

type PoolsResponse struct {
	Pools []liquiditytypes.Pool `json:"pools"`
}

type PoolResponse struct {
	Pool liquiditytypes.Pool `json:"pool"`
}
