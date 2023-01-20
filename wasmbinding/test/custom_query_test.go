package wasmbinding_test

import (
	"github.com/crescent-network/crescent/v4/wasmbinding/bindings"
)

func (s *WasmBindingTestSuite) TestPairs() {
	s.createPair(s.addr(0), "denom1", "denom2", true)
	s.createPair(s.addr(0), "denom3", "denom4", true)
	s.createPair(s.addr(0), "denom5", "denom6", true)

	s.storeCode(s.addr(0), "../testdata/crescent_examples.wasm")

	contractAddr := s.instantiateEmptyContract(s.addr(0), s.addr(0))
	s.Require().NotEmpty(contractAddr)

	query := bindings.CrescentQuery{
		Pairs: &bindings.Pairs{},
	}
	resp := bindings.PairsResponse{}

	s.querySmart(contractAddr, query, resp)
}
