package types_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/tendermint/farming/x/farming/types"
)

func TestPublicPlanProposal_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(*types.PublicPlanProposal)
		expectedErr string
	}{
		{
			"happy case",
			func(proposal *types.PublicPlanProposal) {},
			"",
		},
		{
			"empty proposals",
			func(proposal *types.PublicPlanProposal) {
				proposal.AddRequestProposals = []*types.AddRequestProposal{}
				proposal.UpdateRequestProposals = []*types.UpdateRequestProposal{}
				proposal.DeleteRequestProposals = []*types.DeleteRequestProposal{}
			},
			"proposal request must not be empty: invalid request",
		},
		{
			"invalid add request proposal",
			func(proposal *types.PublicPlanProposal) {
				proposal.AddRequestProposals[0].Name = strings.Repeat("a", 256)
			},
			"plan name cannot be longer than max length of 140: invalid plan name",
		},
		{
			"invalid update request proposal",
			func(proposal *types.PublicPlanProposal) {
				proposal.UpdateRequestProposals[0].Name = strings.Repeat("a", 256)
			},
			"plan name cannot be longer than max length of 140: invalid plan name",
		},
		{
			"invalid delete request proposal",
			func(proposal *types.PublicPlanProposal) {
				proposal.DeleteRequestProposals[0].PlanId = 0
			},
			"invalid plan id: 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proposal := types.NewPublicPlanProposal(
				"title",
				"description",
				[]*types.AddRequestProposal{
					{
						Name:               "name",
						FarmingPoolAddress: sdk.AccAddress(crypto.AddressHash([]byte("address1"))).String(),
						TerminationAddress: sdk.AccAddress(crypto.AddressHash([]byte("address1"))).String(),
						StakingCoinWeights: sdk.NewDecCoins(sdk.NewInt64DecCoin("stake1", 1)),
						StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
						EndTime:            types.ParseTime("9999-12-31T00:00:00Z"),
						EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("reward1", 10000000)),
					},
				},
				[]*types.UpdateRequestProposal{
					{
						PlanId:             1,
						Name:               "new name",
						FarmingPoolAddress: sdk.AccAddress(crypto.AddressHash([]byte("address2"))).String(),
						TerminationAddress: sdk.AccAddress(crypto.AddressHash([]byte("address2"))).String(),
						StakingCoinWeights: sdk.NewDecCoins(sdk.NewInt64DecCoin("stake2", 1)),
						EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("reward2", 10000000)),
					},
				},
				[]*types.DeleteRequestProposal{
					{
						PlanId: 1,
					},
				},
			)
			tc.malleate(proposal)
			err := proposal.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestAddRequestProposal_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(*types.AddRequestProposal)
		expectedErr string
	}{
		{
			"valid for fixed amount plan",
			func(proposal *types.AddRequestProposal) {},
			"",
		},
		{
			"valid for ratio plan",
			func(proposal *types.AddRequestProposal) {
				proposal.EpochAmount = nil
				proposal.EpochRatio = sdk.NewDecWithPrec(5, 2)
			},
			"",
		},
		{
			"invalid plan name",
			func(proposal *types.AddRequestProposal) {
				proposal.Name = "a|b|c"
			},
			"plan name cannot contain |: invalid plan name",
		},
		{
			"ambiguous plan type #1",
			func(proposal *types.AddRequestProposal) {
				proposal.EpochRatio = sdk.NewDecWithPrec(5, 2)
			},
			"exactly one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"ambiguous plan type #2",
			func(proposal *types.AddRequestProposal) {
				proposal.EpochAmount = nil
			},
			"exactly one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"empty name",
			func(proposal *types.AddRequestProposal) {
				proposal.Name = ""
			},
			"plan name must not be empty: invalid plan name",
		},
		{
			"too long name",
			func(proposal *types.AddRequestProposal) {
				proposal.Name = strings.Repeat("a", 256)
			},
			"plan name cannot be longer than max length of 140: invalid plan name",
		},
		{
			"invalid farming pool addr",
			func(proposal *types.AddRequestProposal) {
				proposal.FarmingPoolAddress = "invalid"
			},
			"invalid farming pool address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"invalid termination addr",
			func(proposal *types.AddRequestProposal) {
				proposal.TerminationAddress = "invalid"
			},
			"invalid termination address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"invalid staking coin weights - empty",
			func(proposal *types.AddRequestProposal) {
				proposal.StakingCoinWeights = nil
			},
			"staking coin weights must not be empty: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid",
			func(proposal *types.AddRequestProposal) {
				proposal.StakingCoinWeights = sdk.DecCoins{
					sdk.DecCoin{Denom: "stake1", Amount: sdk.ZeroDec()},
				}
			},
			"invalid staking coin weights: coin 0.000000000000000000stake1 amount is not positive: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid sum of weights",
			func(proposal *types.AddRequestProposal) {
				proposal.StakingCoinWeights = sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(5, 1)),
					sdk.NewDecCoinFromDec("stake2", sdk.NewDecWithPrec(6, 1)),
				)
			},
			"total weight must be 1: invalid staking coin weights",
		},
		{
			"invalid start/end time",
			func(proposal *types.AddRequestProposal) {
				proposal.StartTime = types.ParseTime("2021-10-01T00:00:00Z")
				proposal.EndTime = types.ParseTime("2021-09-01T00:00:00Z")
			},
			"end time 2021-09-01 00:00:00 +0000 UTC must be greater than start time 2021-10-01 00:00:00 +0000 UTC: invalid plan end time",
		},
		{
			"empty epoch amount",
			func(proposal *types.AddRequestProposal) {
				proposal.EpochAmount = sdk.NewCoins()
			},
			"exactly one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"invalid epoch amount",
			func(proposal *types.AddRequestProposal) {
				proposal.EpochAmount = sdk.Coins{sdk.NewInt64Coin("reward1", 0)}
			},
			"invalid epoch amount: coin 0reward1 amount is not positive: invalid request",
		},
		{
			"zero epoch ratio",
			func(proposal *types.AddRequestProposal) {
				proposal.EpochAmount = nil
				proposal.EpochRatio = sdk.ZeroDec()
			},
			"exactly one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"too big epoch ratio",
			func(proposal *types.AddRequestProposal) {
				proposal.EpochAmount = nil
				proposal.EpochRatio = sdk.NewDec(2)
			},
			"epoch ratio must be less than 1: 2.000000000000000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proposal := types.NewAddRequestProposal(
				"name",
				sdk.AccAddress(crypto.AddressHash([]byte("address"))).String(),
				sdk.AccAddress(crypto.AddressHash([]byte("address"))).String(),
				sdk.NewDecCoins(sdk.NewInt64DecCoin("stake1", 1)),
				types.ParseTime("0001-01-01T00:00:00Z"),
				types.ParseTime("9999-12-31T00:00:00Z"),
				sdk.NewCoins(sdk.NewInt64Coin("reward1", 10000000)),
				sdk.Dec{},
			)
			tc.malleate(proposal)
			err := proposal.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestUpdateRequestProposal_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(*types.UpdateRequestProposal)
		expectedErr string
	}{
		{
			"valid for fixed amount plan",
			func(proposal *types.UpdateRequestProposal) {},
			"",
		},
		{
			"valid for ratio plan",
			func(proposal *types.UpdateRequestProposal) {
				proposal.EpochAmount = nil
				proposal.EpochRatio = sdk.NewDecWithPrec(5, 2)
			},
			"",
		},
		{
			"not updating distribution info",
			func(proposal *types.UpdateRequestProposal) {
				proposal.EpochAmount = nil
			},
			"",
		},
		{
			"invalid plan name",
			func(proposal *types.UpdateRequestProposal) {
				proposal.Name = "a|b|c"
			},
			"plan name cannot contain |: invalid plan name",
		},
		{
			"ambiguous plan type",
			func(proposal *types.UpdateRequestProposal) {
				proposal.EpochRatio = sdk.NewDecWithPrec(5, 2)
			},
			"at most one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"invalid plan id",
			func(proposal *types.UpdateRequestProposal) {
				proposal.PlanId = 0
			},
			"invalid plan id: 0: invalid request",
		},
		{
			"empty name",
			func(proposal *types.UpdateRequestProposal) {
				proposal.Name = ""
			},
			"",
		},
		{
			"too long name",
			func(proposal *types.UpdateRequestProposal) {
				proposal.Name = strings.Repeat("a", 256)
			},
			"plan name cannot be longer than max length of 140: invalid plan name",
		},
		{
			"not updating farming pool addr",
			func(proposal *types.UpdateRequestProposal) {
				proposal.FarmingPoolAddress = ""
			},
			"",
		},
		{
			"invalid farming pool addr",
			func(proposal *types.UpdateRequestProposal) {
				proposal.FarmingPoolAddress = "invalid"
			},
			"invalid farming pool address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"not updating termination addr",
			func(proposal *types.UpdateRequestProposal) {
				proposal.TerminationAddress = ""
			},
			"",
		},
		{
			"invalid termination addr",
			func(proposal *types.UpdateRequestProposal) {
				proposal.TerminationAddress = "invalid"
			},
			"invalid termination address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"not updating staking coin weights",
			func(proposal *types.UpdateRequestProposal) {
				proposal.StakingCoinWeights = nil
			},
			"",
		},
		{
			"empty staking coin weights",
			func(proposal *types.UpdateRequestProposal) {
				proposal.StakingCoinWeights = sdk.NewDecCoins()
			},
			"staking coin weights must not be empty: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid",
			func(proposal *types.UpdateRequestProposal) {
				proposal.StakingCoinWeights = sdk.DecCoins{
					sdk.DecCoin{Denom: "stake1", Amount: sdk.ZeroDec()},
				}
			},
			"invalid staking coin weights: coin 0.000000000000000000stake1 amount is not positive: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid sum of weights",
			func(proposal *types.UpdateRequestProposal) {
				proposal.StakingCoinWeights = sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(5, 1)),
					sdk.NewDecCoinFromDec("stake2", sdk.NewDecWithPrec(6, 1)),
				)
			},
			"total weight must be 1: invalid staking coin weights",
		},
		{
			"not updating start/end time",
			func(proposal *types.UpdateRequestProposal) {
				proposal.StartTime = nil
				proposal.EndTime = nil
			},
			"",
		},
		{
			"invalid start/end time",
			func(proposal *types.UpdateRequestProposal) {
				t := types.ParseTime("2021-10-01T00:00:00Z")
				proposal.StartTime = &t
				t2 := types.ParseTime("2021-09-01T00:00:00Z")
				proposal.EndTime = &t2
			},
			"end time 2021-09-01 00:00:00 +0000 UTC must be greater than start time 2021-10-01 00:00:00 +0000 UTC: invalid plan end time",
		},
		{
			"update only start time",
			func(proposal *types.UpdateRequestProposal) {
				t := types.ParseTime("2021-10-01T00:00:00Z")
				proposal.StartTime = &t
			},
			"",
		},
		{
			"update only end time",
			func(proposal *types.UpdateRequestProposal) {
				t := types.ParseTime("2021-10-01T00:00:00Z")
				proposal.EndTime = &t
			},
			"",
		},
		{
			"empty epoch amount",
			func(proposal *types.UpdateRequestProposal) {
				proposal.EpochAmount = sdk.NewCoins()
			},
			"",
		},
		{
			"invalid epoch amount",
			func(proposal *types.UpdateRequestProposal) {
				proposal.EpochAmount = sdk.Coins{sdk.NewInt64Coin("reward1", 0)}
			},
			"invalid epoch amount: coin 0reward1 amount is not positive: invalid request",
		},
		{
			"zero epoch ratio",
			func(proposal *types.UpdateRequestProposal) {
				proposal.EpochAmount = nil
				proposal.EpochRatio = sdk.ZeroDec()
			},
			"",
		},
		{
			"too big epoch ratio",
			func(proposal *types.UpdateRequestProposal) {
				proposal.EpochAmount = nil
				proposal.EpochRatio = sdk.NewDec(2)
			},
			"epoch ratio must be less than 1: 2.000000000000000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proposal := types.NewUpdateRequestProposal(
				1,
				"name",
				sdk.AccAddress(crypto.AddressHash([]byte("address"))).String(),
				sdk.AccAddress(crypto.AddressHash([]byte("address"))).String(),
				sdk.NewDecCoins(sdk.NewInt64DecCoin("stake1", 1)),
				types.ParseTime("0001-01-01T00:00:00Z"),
				types.ParseTime("9999-12-31T00:00:00Z"),
				sdk.NewCoins(sdk.NewInt64Coin("reward1", 10000000)),
				sdk.Dec{},
			)
			tc.malleate(proposal)
			err := proposal.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestDeleteRequestProposal_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(*types.DeleteRequestProposal)
		expectedErr string
	}{
		{
			"happy case",
			func(proposal *types.DeleteRequestProposal) {},
			"",
		},
		{
			"invalid plan id",
			func(proposal *types.DeleteRequestProposal) {
				proposal.PlanId = 0
			},
			"invalid plan id: 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proposal := types.NewDeleteRequestProposal(1)
			tc.malleate(proposal)
			err := proposal.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
