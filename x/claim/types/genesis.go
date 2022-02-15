package types

import (
	"errors"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Airdrops:     []Airdrop{},
		ClaimRecords: []ClaimRecord{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, a := range gs.Airdrops {
		if err := a.Validate(); err != nil {
			return err
		}
	}

	for _, cr := range gs.ClaimRecords {
		if err := cr.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates an airdrop object.
func (a Airdrop) Validate() error {
	if _, err := sdk.AccAddressFromBech32(a.SourceAddress); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(a.TerminationAddress); err != nil {
		return err
	}
	if !a.SourceCoins.IsAllPositive() {
		return errors.New("source coins must be all positive")
	}
	if !a.EndTime.After(a.StartTime) {
		return errors.New("end time must be greater than start time")
	}
	return nil
}

// Validate validates claim record object.
func (r ClaimRecord) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.Recipient); err != nil {
		return err
	}
	if !r.InitialClaimableCoins.IsAllPositive() {
		return fmt.Errorf("initial claimable amount must be positive: %s", r.InitialClaimableCoins.String())
	}
	if !r.ClaimableCoins.IsAllPositive() {
		return fmt.Errorf("claimable amount must be positive: %s", r.InitialClaimableCoins.String())
	}
	for _, action := range r.Actions {
		if action.ActionType != ActionTypeDeposit &&
			action.ActionType != ActionTypeSwap &&
			action.ActionType != ActionTypeFarming {
			return fmt.Errorf("unknown action type %T", action.ActionType)
		}
	}
	return nil
}
