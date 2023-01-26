package wasmbinding_test

import (
	"encoding/json"

	"github.com/crescent-network/crescent/v4/wasmbinding/bindings"
)

func (s *WasmBindingTestSuite) TestPairs() {
	s.createPair(s.addr(0), "denom1", "denom2", true)
	s.createPair(s.addr(0), "denom3", "denom4", true)
	s.createPair(s.addr(0), "denom5", "denom6", true)

	s.storeCode(s.addr(0), "../testdata/crescent_liquidity.wasm")

	contractAddr := s.instantiateEmptyContract(s.addr(0), s.addr(0))
	s.Require().NotEmpty(contractAddr)

	req := bindings.CrescentQuery{
		Pairs: &bindings.Pairs{},
	}

	queryBz, err := json.Marshal(req)
	s.Require().NoError(err)

	resBz, err := s.app.WasmKeeper.QuerySmart(s.ctx, contractAddr, queryBz)
	s.Require().NoError(err)

	var resp bindings.PairsResponse
	err = json.Unmarshal(resBz, &resp)
	s.Require().NoError(err)
	s.Require().Len(resp.Pairs, 3)
}

func (s *WasmBindingTestSuite) TestPair() {
	s.createPair(s.addr(0), "denom1", "denom2", true)

	s.storeCode(s.addr(0), "../testdata/crescent_liquidity.wasm")

	contractAddr := s.instantiateEmptyContract(s.addr(0), s.addr(0))
	s.Require().NotEmpty(contractAddr)

	req := bindings.CrescentQuery{
		Pair: &bindings.Pair{Id: 1},
	}

	queryBz, err := json.Marshal(req)
	s.Require().NoError(err)

	resBz, err := s.app.WasmKeeper.QuerySmart(s.ctx, contractAddr, queryBz)
	s.Require().NoError(err)

	var resp bindings.PairResponse
	err = json.Unmarshal(resBz, &resp)
	s.Require().NoError(err)
}
