package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func TestPublicPositionCreateProposal_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(p *types.PublicPositionCreateProposal)
		expectedErr string
	}{
		{
			"valid",
			func(p *types.PublicPositionCreateProposal) {},
			"",
		},
		{
			"invalid pool id",
			func(p *types.PublicPositionCreateProposal) {
				p.PoolId = 0
			},
			"pool id must not be 0: invalid request",
		},
		{
			"invalid lower price",
			func(p *types.PublicPositionCreateProposal) {
				p.LowerPrice = utils.ParseDec("0")
			},
			"lower price must be positive: 0.000000000000000000: invalid request",
		},
		{
			"invalid upper price",
			func(p *types.PublicPositionCreateProposal) {
				p.UpperPrice = utils.ParseDec("0")
			},
			"upper price must be positive: 0.000000000000000000: invalid request",
		},
		{
			"invalid fee rate",
			func(p *types.PublicPositionCreateProposal) {
				p.FeeRate = utils.ParseDec("1.01")
			},
			"fee rate must be in range [0, 1]: 1.010000000000000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewPublicPositionCreateProposal(
				"Title", "Description", 1,
				utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("0.003"))
			require.Equal(t, types.ProposalTypePublicPositionCreate, p.ProposalType())
			tc.malleate(p)
			err := p.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestPublicPositionParameterChangeProposal_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(p *types.PublicPositionParameterChangeProposal)
		expectedErr string
	}{
		{
			"valid",
			func(p *types.PublicPositionParameterChangeProposal) {},
			"",
		},
		{
			"no changes",
			func(p *types.PublicPositionParameterChangeProposal) {
				p.Changes = nil
			},
			"changes must not be empty: invalid request",
		},
		{
			"invalid pool id",
			func(p *types.PublicPositionParameterChangeProposal) {
				p.Changes = []types.PublicPositionParameterChange{
					types.NewPublicPositionParameterChange(0, utils.ParseDec("0.002")),
				}
			},
			"public position id must not be 0: invalid request",
		},
		{
			"invalid fee rate",
			func(p *types.PublicPositionParameterChangeProposal) {
				p.Changes = []types.PublicPositionParameterChange{
					types.NewPublicPositionParameterChange(2, utils.ParseDec("1.01")),
				}
			},
			"fee rate must be in range [0, 1]: 1.010000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewPublicPositionParameterChangeProposal(
				"Title", "Description", []types.PublicPositionParameterChange{
					types.NewPublicPositionParameterChange(2, utils.ParseDec("0.002"))})
			require.Equal(t, types.ProposalTypePublicPositionParameterChange, p.ProposalType())
			tc.malleate(p)
			err := p.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func ExamplePublicPositionCreateProposal_String() {
	p := types.NewPublicPositionCreateProposal(
		"Title", "Description", 2,
		utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("0.003"))
	fmt.Println(p.String())

	// Output:
	// Public Position Create Proposal:
	//   Title:       Title
	//   Description: Description
	//   Pool Id:     2
	//   Lower Price: 0.900000000000000000
	//   Upper Price: 1.100000000000000000
	//   Fee Rate:    0.003000000000000000
}

func ExamplePublicPositionParameterChangeProposal_String() {
	p := types.NewPublicPositionParameterChangeProposal(
		"Title", "Description", []types.PublicPositionParameterChange{
			types.NewPublicPositionParameterChange(2, utils.ParseDec("0.002"))})
	fmt.Println(p.String())

	// Output:
	// Public Position Parameter Change Proposal:
	//   Title:       Title
	//   Description: Description
	//   Changes:
	//     Public Position Parameter Change:
	//       Public Position Id: 2
	//       Fee Rate:           0.002000000000000000
}
