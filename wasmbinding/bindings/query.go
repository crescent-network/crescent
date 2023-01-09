package bindings

import (
	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
)

// CrescentQuery contains custom queries for Crescent Network.
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
	Pairs []liquiditytypes.Pair `json:"pairs"`
}

type PairResponse struct {
	Pair liquiditytypes.Pair `json:"pair"`
}

type PoolsResponse struct {
	Pools []liquiditytypes.Pool `json:"pools"`
}

type PoolResponse struct {
	Pool liquiditytypes.Pool `json:"pool"`
}
