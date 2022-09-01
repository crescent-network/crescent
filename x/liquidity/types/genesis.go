package types

import (
	"fmt"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                   DefaultParams(),
		LastPairId:               0,
		LastPoolId:               0,
		Pairs:                    []Pair{},
		Pools:                    []Pool{},
		DepositRequests:          []DepositRequest{},
		WithdrawRequests:         []WithdrawRequest{},
		Orders:                   []Order{},
		MarketMakingOrderIndexes: []MMOrderIndex{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	pairMap := map[uint64]Pair{}
	for i, pair := range genState.Pairs {
		if err := pair.Validate(); err != nil {
			return fmt.Errorf("invalid pair at index %d: %w", i, err)
		}
		if pair.Id > genState.LastPairId {
			return fmt.Errorf("pair at index %d has an id greater than last pair id: %d", i, pair.Id)
		}
		if _, ok := pairMap[pair.Id]; ok {
			return fmt.Errorf("pair at index %d has a duplicate id: %d", i, pair.Id)
		}
		pairMap[pair.Id] = pair
	}
	poolMap := map[uint64]Pool{}
	for i, pool := range genState.Pools {
		if err := pool.Validate(); err != nil {
			return fmt.Errorf("invalid pool at index %d: %w", i, err)
		}
		if pool.Id > genState.LastPoolId {
			return fmt.Errorf("pool at index %d has an id greater than last pool id: %d", i, pool.Id)
		}
		if _, ok := pairMap[pool.PairId]; !ok {
			return fmt.Errorf("pool at index %d has unknown pair id: %d", i, pool.PairId)
		}
		if _, ok := poolMap[pool.Id]; ok {
			return fmt.Errorf("pool at index %d has a duplicate pool id: %d", i, pool.Id)
		}
		poolMap[pool.Id] = pool
	}
	depositReqSet := map[uint64]map[uint64]struct{}{}
	for i, req := range genState.DepositRequests {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("invalid deposit request at index %d: %w", i, err)
		}
		pool, ok := poolMap[req.PoolId]
		if !ok {
			return fmt.Errorf("deposit request at index %d has unknown pool id: %d", i, req.PoolId)
		}
		if req.MintedPoolCoin.Denom != pool.PoolCoinDenom {
			return fmt.Errorf("deposit request at index %d has wrong minted pool coin: %s", i, req.MintedPoolCoin)
		}
		pair := pairMap[pool.PairId]
		if req.DepositCoins.AmountOf(pair.BaseCoinDenom).IsZero() ||
			req.DepositCoins.AmountOf(pair.QuoteCoinDenom).IsZero() {
			return fmt.Errorf("deposit request at index %d has wrong deposit coins: %s", i, req.DepositCoins)
		}
		if set, ok := depositReqSet[req.PoolId]; ok {
			if _, ok := set[req.Id]; ok {
				return fmt.Errorf("deposit request at index %d has a duplicate id: %d", i, req.Id)
			}
		} else {
			depositReqSet[req.PoolId] = map[uint64]struct{}{}
		}
		depositReqSet[req.PoolId][req.Id] = struct{}{}
	}
	withdrawReqSet := map[uint64]map[uint64]struct{}{}
	for i, req := range genState.WithdrawRequests {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("invalid withdraw request at index %d: %w", i, err)
		}
		pool, ok := poolMap[req.PoolId]
		if !ok {
			return fmt.Errorf("withdraw request at index %d has unknown pool id: %d", i, req.PoolId)
		}
		if req.PoolCoin.Denom != pool.PoolCoinDenom {
			return fmt.Errorf("withdraw request at index %d has wrong pool coin: %s", i, req.PoolCoin)
		}
		if set, ok := withdrawReqSet[req.PoolId]; ok {
			if _, ok := set[req.Id]; ok {
				return fmt.Errorf("withdraw request at index %d has a duplicate id: %d", i, req.Id)
			}
		} else {
			withdrawReqSet[req.PoolId] = map[uint64]struct{}{}
		}
		withdrawReqSet[req.PoolId][req.Id] = struct{}{}
	}
	orderSet := map[uint64]map[uint64]struct{}{}
	for i, order := range genState.Orders {
		if err := order.Validate(); err != nil {
			return fmt.Errorf("invalid order at index %d: %w", i, err)
		}
		pair, ok := pairMap[order.PairId]
		if !ok {
			return fmt.Errorf("order at index %d has unknown pair id: %d", i, order.PairId)
		}
		if order.BatchId > pair.CurrentBatchId {
			return fmt.Errorf("order at index %d has a batch id greater than its pair's current batch id: %d", i, order.BatchId)
		}
		var offerCoinDenom, demandCoinDenom string
		switch order.Direction {
		case OrderDirectionBuy:
			offerCoinDenom, demandCoinDenom = pair.QuoteCoinDenom, pair.BaseCoinDenom
		case OrderDirectionSell:
			offerCoinDenom, demandCoinDenom = pair.BaseCoinDenom, pair.QuoteCoinDenom
		}
		if order.OfferCoin.Denom != offerCoinDenom {
			return fmt.Errorf("order at index %d has wrong offer coin denom: %s != %s", i, order.OfferCoin.Denom, offerCoinDenom)
		}
		if order.ReceivedCoin.Denom != demandCoinDenom {
			return fmt.Errorf("order at index %d has wrong demand coin denom: %s != %s", i, order.OfferCoin.Denom, demandCoinDenom)
		}
		if set, ok := orderSet[order.PairId]; ok {
			if _, ok := set[order.Id]; ok {
				return fmt.Errorf("order at index %d has a duplicate id: %d", i, order.Id)
			}
		} else {
			orderSet[order.PairId] = map[uint64]struct{}{}
		}
		orderSet[order.PairId][order.Id] = struct{}{}
	}
	return nil
}
