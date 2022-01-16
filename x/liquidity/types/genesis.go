package types

import (
	"fmt"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	for i, pair := range genState.Pairs {
		if err := pair.Validate(); err != nil {
			return fmt.Errorf("invalid pair at index %d: %w", i, err)
		}
	}
	for i, pool := range genState.Pools {
		if err := pool.Validate(); err != nil {
			return fmt.Errorf("invalid pool at index %d: %w", i, err)
		}
	}
	for i, req := range genState.DepositRequests {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("invalid deposit request at index %d: %w", i, err)
		}
	}
	for i, req := range genState.WithdrawRequests {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("invalid withdraw request at index %d: %w", i, err)
		}
	}
	for i, req := range genState.SwapRequests {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("invalid swap request at index %d: %w", i, err)
		}
	}
	for i, req := range genState.CancelSwapRequests {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("invalid cancel swap request at index %d: %w", i, err)
		}
	}
	return nil
}
