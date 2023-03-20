package types

import (
	"fmt"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                       DefaultParams(),
		LastPairId:                   0,
		LastPoolId:                   0,
		Pairs:                        []PairRecord{},
		Pools:                        []PoolRecord{},
		DepositRequests:              []DepositRequest{},
		WithdrawRequests:             []WithdrawRequest{},
		Orders:                       []OrderRecord{},
		NumMarketMakingOrdersRecords: []NumMMOrdersRecord{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	pairRecordMap := map[uint64]PairRecord{}
	for i, record := range genState.Pairs {
		if err := record.Pair.Validate(); err != nil {
			return fmt.Errorf("invalid pair at index %d: %w", i, err)
		}
		if record.Pair.Id > genState.LastPairId {
			return fmt.Errorf("pair at index %d has an id greater than last pair id: %d", i, record.Pair.Id)
		}
		if _, ok := pairRecordMap[record.Pair.Id]; ok {
			return fmt.Errorf("pair at index %d has a duplicate id: %d", i, record.Pair.Id)
		}
		pairRecordMap[record.Pair.Id] = record
	}
	poolRecordMap := map[uint64]PoolRecord{}
	for i, record := range genState.Pools {
		if err := record.Pool.Validate(); err != nil {
			return fmt.Errorf("invalid pool at index %d: %w", i, err)
		}
		if record.Pool.Id > genState.LastPoolId {
			return fmt.Errorf("pool at index %d has an id greater than last pool id: %d", i, record.Pool.Id)
		}
		if _, ok := pairRecordMap[record.Pool.PairId]; !ok {
			return fmt.Errorf("pool at index %d has unknown pair id: %d", i, record.Pool.PairId)
		}
		if _, ok := poolRecordMap[record.Pool.Id]; ok {
			return fmt.Errorf("pool at index %d has a duplicate pool id: %d", i, record.Pool.Id)
		}
		poolRecordMap[record.Pool.Id] = record
	}
	depositReqSet := map[uint64]map[uint64]struct{}{}
	for i, req := range genState.DepositRequests {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("invalid deposit request at index %d: %w", i, err)
		}
		poolRecord, ok := poolRecordMap[req.PoolId]
		if !ok {
			return fmt.Errorf("deposit request at index %d has unknown pool id: %d", i, req.PoolId)
		}
		if req.MintedPoolCoin.Denom != poolRecord.Pool.PoolCoinDenom {
			return fmt.Errorf("deposit request at index %d has wrong minted pool coin: %s", i, req.MintedPoolCoin)
		}
		pairRecord := pairRecordMap[poolRecord.Pool.PairId]
		if req.DepositCoins.AmountOf(pairRecord.Pair.BaseCoinDenom).IsZero() ||
			req.DepositCoins.AmountOf(pairRecord.Pair.QuoteCoinDenom).IsZero() {
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
		poolRecord, ok := poolRecordMap[req.PoolId]
		if !ok {
			return fmt.Errorf("withdraw request at index %d has unknown pool id: %d", i, req.PoolId)
		}
		if req.PoolCoin.Denom != poolRecord.Pool.PoolCoinDenom {
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
	for i, record := range genState.Orders {
		if err := record.Order.Validate(); err != nil {
			return fmt.Errorf("invalid order at index %d: %w", i, err)
		}
		pairRecord, ok := pairRecordMap[record.Order.PairId]
		if !ok {
			return fmt.Errorf("order at index %d has unknown pair id: %d", i, record.Order.PairId)
		}
		if record.Order.BatchId > pairRecord.State.CurrentBatchId {
			return fmt.Errorf("order at index %d has a batch id greater than its pair's current batch id: %d", i, record.Order.BatchId)
		}
		if set, ok := orderSet[record.Order.PairId]; ok {
			if _, ok := set[record.Order.Id]; ok {
				return fmt.Errorf("order at index %d has a duplicate id: %d", i, record.Order.Id)
			}
		} else {
			orderSet[record.Order.PairId] = map[uint64]struct{}{}
		}
		orderSet[record.Order.PairId][record.Order.Id] = struct{}{}
	}
	for _, record := range genState.NumMarketMakingOrdersRecords {
		if record.NumMarketMakingOrders == 0 {
			return fmt.Errorf("number of MM order must be positive")
		}
	}
	return nil
}
