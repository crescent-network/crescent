package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		ModuleAccountBalance: sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, sdk.ZeroInt()),
		ClaimRecords:         []ClaimRecord{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// TODO: not implemented yet
	return nil
}
