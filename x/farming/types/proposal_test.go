package types_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v2/x/farming/types"
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
				proposal.AddPlanRequests = []types.AddPlanRequest{}
				proposal.ModifyPlanRequests = []types.ModifyPlanRequest{}
				proposal.DeletePlanRequests = []types.DeletePlanRequest{}
			},
			"proposal request must not be empty: invalid request",
		},
		{
			"invalid add request proposal",
			func(proposal *types.PublicPlanProposal) {
				proposal.AddPlanRequests[0].Name = strings.Repeat("a", 256)
			},
			"plan name cannot be longer than max length of 140: invalid plan name",
		},
		{
			"invalid update request proposal",
			func(proposal *types.PublicPlanProposal) {
				proposal.ModifyPlanRequests[0].Name = strings.Repeat("a", 256)
			},
			"plan name cannot be longer than max length of 140: invalid plan name",
		},
		{
			"invalid delete request proposal",
			func(proposal *types.PublicPlanProposal) {
				proposal.DeletePlanRequests[0].PlanId = 0
			},
			"invalid plan id: 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proposal := types.NewPublicPlanProposal(
				"title",
				"description",
				[]types.AddPlanRequest{
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
				[]types.ModifyPlanRequest{
					{
						PlanId:             1,
						Name:               "new name",
						FarmingPoolAddress: sdk.AccAddress(crypto.AddressHash([]byte("address2"))).String(),
						TerminationAddress: sdk.AccAddress(crypto.AddressHash([]byte("address2"))).String(),
						StakingCoinWeights: sdk.NewDecCoins(sdk.NewInt64DecCoin("stake2", 1)),
						EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("reward2", 10000000)),
					},
				},
				[]types.DeletePlanRequest{
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

func TestAddPlanRequest_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(*types.AddPlanRequest)
		expectedErr string
	}{
		{
			"valid for fixed amount plan",
			func(req *types.AddPlanRequest) {},
			"",
		},
		{
			"valid for ratio plan",
			func(req *types.AddPlanRequest) {
				req.EpochAmount = nil
				req.EpochRatio = sdk.NewDecWithPrec(5, 2)
			},
			"",
		},
		{
			"invalid plan name",
			func(req *types.AddPlanRequest) {
				req.Name = "a|b|c"
			},
			"plan name cannot contain |: invalid plan name",
		},
		{
			"ambiguous plan type #1",
			func(req *types.AddPlanRequest) {
				req.EpochRatio = sdk.NewDecWithPrec(5, 2)
			},
			"exactly one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"ambiguous plan type #2",
			func(req *types.AddPlanRequest) {
				req.EpochAmount = nil
			},
			"exactly one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"empty name",
			func(req *types.AddPlanRequest) {
				req.Name = ""
			},
			"plan name must not be empty: invalid plan name",
		},
		{
			"too long name",
			func(req *types.AddPlanRequest) {
				req.Name = strings.Repeat("a", 256)
			},
			"plan name cannot be longer than max length of 140: invalid plan name",
		},
		{
			"invalid farming pool addr",
			func(req *types.AddPlanRequest) {
				req.FarmingPoolAddress = "invalid"
			},
			"invalid farming pool address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"invalid termination addr",
			func(req *types.AddPlanRequest) {
				req.TerminationAddress = "invalid"
			},
			"invalid termination address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"invalid staking coin weights - empty",
			func(req *types.AddPlanRequest) {
				req.StakingCoinWeights = nil
			},
			"staking coin weights must not be empty: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid",
			func(req *types.AddPlanRequest) {
				req.StakingCoinWeights = sdk.DecCoins{
					sdk.DecCoin{Denom: "stake1", Amount: sdk.ZeroDec()},
				}
			},
			"invalid staking coin weights: coin 0.000000000000000000stake1 amount is not positive: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid sum of weights",
			func(req *types.AddPlanRequest) {
				req.StakingCoinWeights = sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(5, 1)),
					sdk.NewDecCoinFromDec("stake2", sdk.NewDecWithPrec(6, 1)),
				)
			},
			"total weight must be 1: invalid staking coin weights",
		},
		{
			"invalid start/end time",
			func(req *types.AddPlanRequest) {
				req.StartTime = types.ParseTime("2021-10-01T00:00:00Z")
				req.EndTime = types.ParseTime("2021-09-01T00:00:00Z")
			},
			"end time 2021-09-01T00:00:00Z must be greater than start time 2021-10-01T00:00:00Z: invalid plan end time",
		},
		{
			"empty epoch amount",
			func(req *types.AddPlanRequest) {
				req.EpochAmount = sdk.NewCoins()
			},
			"exactly one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"invalid epoch amount",
			func(req *types.AddPlanRequest) {
				req.EpochAmount = sdk.Coins{sdk.NewInt64Coin("reward1", 0)}
			},
			"invalid epoch amount: coin 0reward1 amount is not positive: invalid request",
		},
		{
			"zero epoch ratio",
			func(req *types.AddPlanRequest) {
				req.EpochAmount = nil
				req.EpochRatio = sdk.ZeroDec()
			},
			"exactly one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"too big epoch ratio",
			func(req *types.AddPlanRequest) {
				req.EpochAmount = nil
				req.EpochRatio = sdk.NewDec(2)
			},
			"epoch ratio must be less than 1: 2.000000000000000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := types.NewAddPlanRequest(
				"name",
				sdk.AccAddress(crypto.AddressHash([]byte("address"))).String(),
				sdk.AccAddress(crypto.AddressHash([]byte("address"))).String(),
				sdk.NewDecCoins(sdk.NewInt64DecCoin("stake1", 1)),
				types.ParseTime("0001-01-01T00:00:00Z"),
				types.ParseTime("9999-12-31T00:00:00Z"),
				sdk.NewCoins(sdk.NewInt64Coin("reward1", 10000000)),
				sdk.Dec{},
			)
			tc.malleate(&req)
			err := req.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestModifyPlanRequest_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(*types.ModifyPlanRequest)
		expectedErr string
	}{
		{
			"valid for fixed amount plan",
			func(req *types.ModifyPlanRequest) {},
			"",
		},
		{
			"valid for ratio plan",
			func(req *types.ModifyPlanRequest) {
				req.EpochAmount = nil
				req.EpochRatio = sdk.NewDecWithPrec(5, 2)
			},
			"",
		},
		{
			"not updating distribution info",
			func(req *types.ModifyPlanRequest) {
				req.EpochAmount = nil
			},
			"",
		},
		{
			"invalid plan name",
			func(req *types.ModifyPlanRequest) {
				req.Name = "a|b|c"
			},
			"plan name cannot contain |: invalid plan name",
		},
		{
			"ambiguous plan type",
			func(req *types.ModifyPlanRequest) {
				req.EpochRatio = sdk.NewDecWithPrec(5, 2)
			},
			"at most one of epoch amount or epoch ratio must be provided: invalid request",
		},
		{
			"invalid plan id",
			func(req *types.ModifyPlanRequest) {
				req.PlanId = 0
			},
			"invalid plan id: 0: invalid request",
		},
		{
			"empty name",
			func(req *types.ModifyPlanRequest) {
				req.Name = ""
			},
			"",
		},
		{
			"too long name",
			func(req *types.ModifyPlanRequest) {
				req.Name = strings.Repeat("a", 256)
			},
			"plan name cannot be longer than max length of 140: invalid plan name",
		},
		{
			"not updating farming pool addr",
			func(req *types.ModifyPlanRequest) {
				req.FarmingPoolAddress = ""
			},
			"",
		},
		{
			"invalid farming pool addr",
			func(req *types.ModifyPlanRequest) {
				req.FarmingPoolAddress = "invalid"
			},
			"invalid farming pool address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"not updating termination addr",
			func(req *types.ModifyPlanRequest) {
				req.TerminationAddress = ""
			},
			"",
		},
		{
			"invalid termination addr",
			func(req *types.ModifyPlanRequest) {
				req.TerminationAddress = "invalid"
			},
			"invalid termination address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"not updating staking coin weights",
			func(req *types.ModifyPlanRequest) {
				req.StakingCoinWeights = nil
			},
			"",
		},
		{
			"empty staking coin weights",
			func(req *types.ModifyPlanRequest) {
				req.StakingCoinWeights = sdk.NewDecCoins()
			},
			"staking coin weights must not be empty: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid",
			func(req *types.ModifyPlanRequest) {
				req.StakingCoinWeights = sdk.DecCoins{
					sdk.DecCoin{Denom: "stake1", Amount: sdk.ZeroDec()},
				}
			},
			"invalid staking coin weights: coin 0.000000000000000000stake1 amount is not positive: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid sum of weights",
			func(req *types.ModifyPlanRequest) {
				req.StakingCoinWeights = sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(5, 1)),
					sdk.NewDecCoinFromDec("stake2", sdk.NewDecWithPrec(6, 1)),
				)
			},
			"total weight must be 1: invalid staking coin weights",
		},
		{
			"not updating start/end time",
			func(req *types.ModifyPlanRequest) {
				req.StartTime = nil
				req.EndTime = nil
			},
			"",
		},
		{
			"invalid start/end time",
			func(req *types.ModifyPlanRequest) {
				t := types.ParseTime("2021-10-01T00:00:00Z")
				req.StartTime = &t
				t2 := types.ParseTime("2021-09-01T00:00:00Z")
				req.EndTime = &t2
			},
			"end time 2021-09-01 00:00:00 +0000 UTC must be greater than start time 2021-10-01 00:00:00 +0000 UTC: invalid plan end time",
		},
		{
			"update only start time",
			func(req *types.ModifyPlanRequest) {
				t := types.ParseTime("2021-10-01T00:00:00Z")
				req.StartTime = &t
			},
			"",
		},
		{
			"update only end time",
			func(req *types.ModifyPlanRequest) {
				t := types.ParseTime("2021-10-01T00:00:00Z")
				req.EndTime = &t
			},
			"",
		},
		{
			"empty epoch amount",
			func(req *types.ModifyPlanRequest) {
				req.EpochAmount = sdk.NewCoins()
			},
			"",
		},
		{
			"invalid epoch amount",
			func(req *types.ModifyPlanRequest) {
				req.EpochAmount = sdk.Coins{sdk.NewInt64Coin("reward1", 0)}
			},
			"invalid epoch amount: coin 0reward1 amount is not positive: invalid request",
		},
		{
			"zero epoch ratio",
			func(req *types.ModifyPlanRequest) {
				req.EpochAmount = nil
				req.EpochRatio = sdk.ZeroDec()
			},
			"",
		},
		{
			"too big epoch ratio",
			func(req *types.ModifyPlanRequest) {
				req.EpochAmount = nil
				req.EpochRatio = sdk.NewDec(2)
			},
			"epoch ratio must be less than 1: 2.000000000000000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := types.NewModifyPlanRequest(
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
			tc.malleate(&req)
			err := req.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestDeletePlanRequest_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(*types.DeletePlanRequest)
		expectedErr string
	}{
		{
			"happy case",
			func(req *types.DeletePlanRequest) {},
			"",
		},
		{
			"invalid plan id",
			func(req *types.DeletePlanRequest) {
				req.PlanId = 0
			},
			"invalid plan id: 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := types.NewDeletePlanRequest(1)
			tc.malleate(&req)
			err := req.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
