package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v5/x/marketmaker/types"
)

func TestMarketMakerProposal_ValidateBasic(t *testing.T) {
	mm1 := sdk.AccAddress(crypto.AddressHash([]byte("mm1")))
	mm2 := sdk.AccAddress(crypto.AddressHash([]byte("mm2")))

	for _, tc := range []struct {
		name        string
		malleate    func(proposal *types.MarketMakerProposal)
		expectedErr string
	}{
		{
			"happy case",
			func(proposal *types.MarketMakerProposal) {},
			"",
		},
		{
			"empty proposals",
			func(proposal *types.MarketMakerProposal) {
				proposal.Inclusions = []types.MarketMakerHandle{}
				proposal.Exclusions = []types.MarketMakerHandle{}
				proposal.Rejections = []types.MarketMakerHandle{}
				proposal.Distributions = []types.IncentiveDistribution{}
			},
			"proposal request must not be empty: invalid request",
		},
		{
			"duplicated market maker on inclusion",
			func(proposal *types.MarketMakerProposal) {
				proposal.Inclusions = []types.MarketMakerHandle{
					{Address: mm1.String(), PairId: 1},
					{Address: mm1.String(), PairId: 1},
					{Address: mm1.String(), PairId: 3},
				}
			},
			"market maker can't be duplicated: invalid request",
		},
		{
			"duplicated market maker on exclusion",
			func(proposal *types.MarketMakerProposal) {
				proposal.Exclusions = []types.MarketMakerHandle{
					{Address: mm2.String(), PairId: 2},
					{Address: mm2.String(), PairId: 2},
				}
			},
			"market maker can't be duplicated: invalid request",
		},
		{
			"duplicated market maker on inclusion and exclusion",
			func(proposal *types.MarketMakerProposal) {
				proposal.Inclusions = []types.MarketMakerHandle{
					{Address: mm1.String(), PairId: 1},
				}
				proposal.Exclusions = []types.MarketMakerHandle{
					{Address: mm1.String(), PairId: 1},
				}
			},
			"market maker can't be duplicated: invalid request",
		},
		{
			"duplicated market maker on inclusion and rejection",
			func(proposal *types.MarketMakerProposal) {
				proposal.Inclusions = []types.MarketMakerHandle{
					{Address: mm1.String(), PairId: 1},
				}
				proposal.Rejections = []types.MarketMakerHandle{
					{Address: mm1.String(), PairId: 1},
				}
			},
			"market maker can't be duplicated: invalid request",
		},
		{
			"zero pair id",
			func(proposal *types.MarketMakerProposal) {
				proposal.Inclusions = []types.MarketMakerHandle{
					{Address: mm1.String(), PairId: 0},
				}
			},
			"invalid pair id",
		},
		{
			"invalid market maker address",
			func(proposal *types.MarketMakerProposal) {
				proposal.Exclusions = []types.MarketMakerHandle{
					{Address: "invalidaddr", PairId: 1},
				}
			},
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid incentive amount",
			func(proposal *types.MarketMakerProposal) {
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
			proposal := types.NewMarketMakerProposal(
				"title",
				"description",
				[]types.MarketMakerHandle{
					{Address: mm1.String(), PairId: 1},
					{Address: mm1.String(), PairId: 2},
					{Address: mm1.String(), PairId: 3},
				},
				[]types.MarketMakerHandle{
					{Address: mm2.String(), PairId: 4},
				},
				[]types.MarketMakerHandle{
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
