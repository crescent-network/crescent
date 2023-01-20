package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

func TestBootstrapProposal_ValidateBasic(t *testing.T) {
	mm1 := sdk.AccAddress(crypto.AddressHash([]byte("mm1")))
	mm2 := sdk.AccAddress(crypto.AddressHash([]byte("mm2")))

	for _, tc := range []struct {
		name        string
		malleate    func(proposal *types.BootstrapProposal)
		expectedErr string
	}{
		{
			"happy case",
			func(proposal *types.BootstrapProposal) {},
			"",
		},
		{
			"empty proposals",
			func(proposal *types.BootstrapProposal) {
				proposal.Inclusions = []types.BootstrapHandle{}
				proposal.Exclusions = []types.BootstrapHandle{}
				proposal.Rejections = []types.BootstrapHandle{}
				proposal.Distributions = []types.IncentiveDistribution{}
			},
			"proposal request must not be empty: invalid request",
		},
		{
			"duplicated market maker on inclusion",
			func(proposal *types.BootstrapProposal) {
				proposal.Inclusions = []types.BootstrapHandle{
					{Address: mm1.String(), PairId: 1},
					{Address: mm1.String(), PairId: 1},
					{Address: mm1.String(), PairId: 3},
				}
			},
			"market maker can't be duplicated: invalid request",
		},
		{
			"duplicated market maker on exclusion",
			func(proposal *types.BootstrapProposal) {
				proposal.Exclusions = []types.BootstrapHandle{
					{Address: mm2.String(), PairId: 2},
					{Address: mm2.String(), PairId: 2},
				}
			},
			"market maker can't be duplicated: invalid request",
		},
		{
			"duplicated market maker on inclusion and exclusion",
			func(proposal *types.BootstrapProposal) {
				proposal.Inclusions = []types.BootstrapHandle{
					{Address: mm1.String(), PairId: 1},
				}
				proposal.Exclusions = []types.BootstrapHandle{
					{Address: mm1.String(), PairId: 1},
				}
			},
			"market maker can't be duplicated: invalid request",
		},
		{
			"duplicated market maker on inclusion and rejection",
			func(proposal *types.BootstrapProposal) {
				proposal.Inclusions = []types.BootstrapHandle{
					{Address: mm1.String(), PairId: 1},
				}
				proposal.Rejections = []types.BootstrapHandle{
					{Address: mm1.String(), PairId: 1},
				}
			},
			"market maker can't be duplicated: invalid request",
		},
		{
			"zero pair id",
			func(proposal *types.BootstrapProposal) {
				proposal.Inclusions = []types.BootstrapHandle{
					{Address: mm1.String(), PairId: 0},
				}
			},
			"invalid pair id",
		},
		{
			"invalid market maker address",
			func(proposal *types.BootstrapProposal) {
				proposal.Exclusions = []types.BootstrapHandle{
					{Address: "invalidaddr", PairId: 1},
				}
			},
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid incentive amount",
			func(proposal *types.BootstrapProposal) {
				proposal.Distributions = []types.IncentiveDistribution{
					{
						Address: mm1.String(),
						PairId:  1,
						Amount:  sdk.Coins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(0)}},
					},
				}
			},
			"coin 0stake amount is not positive",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proposal := types.NewBootstrapProposal(
				"title",
				"description",
				[]types.BootstrapHandle{
					{Address: mm1.String(), PairId: 1},
					{Address: mm1.String(), PairId: 2},
					{Address: mm1.String(), PairId: 3},
				},
				[]types.BootstrapHandle{
					{Address: mm2.String(), PairId: 4},
				},
				[]types.BootstrapHandle{
					{Address: mm2.String(), PairId: 1},
					{Address: mm2.String(), PairId: 2},
				},
				[]types.IncentiveDistribution{
					{
						Address: mm1.String(),
						PairId:  1,
						Amount:  sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000))),
					},
					{
						Address: mm1.String(),
						PairId:  2,
						Amount:  sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000))),
					},
				})
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
