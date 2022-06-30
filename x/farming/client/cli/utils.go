package cli

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

// PrivateFixedPlanRequest defines CLI request for a private fixed plan.
type PrivateFixedPlanRequest struct {
	Name               string       `json:"name"`
	StakingCoinWeights sdk.DecCoins `json:"staking_coin_weights"`
	StartTime          time.Time    `json:"start_time"`
	EndTime            time.Time    `json:"end_time"`
	EpochAmount        sdk.Coins    `json:"epoch_amount"`
}

// PrivateRatioPlanRequest defines CLI request for a private ratio plan.
type PrivateRatioPlanRequest struct {
	Name               string       `json:"name"`
	StakingCoinWeights sdk.DecCoins `json:"staking_coin_weights"`
	StartTime          time.Time    `json:"start_time"`
	EndTime            time.Time    `json:"end_time"`
	EpochRatio         sdk.Dec      `json:"epoch_ratio"`
}

// ParsePrivateFixedPlan reads and parses a PrivateFixedPlanRequest from a file.
func ParsePrivateFixedPlan(file string) (PrivateFixedPlanRequest, error) {
	plan := PrivateFixedPlanRequest{}

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return plan, err
	}

	if err = json.Unmarshal(contents, &plan); err != nil {
		return plan, err
	}

	return plan, nil
}

// ParsePrivateRatioPlan reads and parses a PrivateRatioPlanRequest from a file.
func ParsePrivateRatioPlan(file string) (PrivateRatioPlanRequest, error) {
	plan := PrivateRatioPlanRequest{}

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return plan, err
	}

	if err = json.Unmarshal(contents, &plan); err != nil {
		return plan, err
	}

	return plan, nil
}

// ParsePublicPlanProposal reads and parses a PublicPlanProposal from a file.
func ParsePublicPlanProposal(cdc codec.JSONCodec, proposalFile string) (types.PublicPlanProposal, error) {
	proposal := types.PublicPlanProposal{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}

// String returns a human readable string representation of the request.
func (req PrivateFixedPlanRequest) String() string {
	result, err := json.Marshal(&req)
	if err != nil {
		panic(err)
	}
	return string(result)
}

// String returns a human readable string representation of the request.
func (req PrivateRatioPlanRequest) String() string {
	result, err := json.Marshal(&req)
	if err != nil {
		panic(err)
	}
	return string(result)
}
