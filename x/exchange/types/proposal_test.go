package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func ExampleMarketParameterChange_String() {
	p := types.NewMarketParameterChangeProposal(
		"Title", "Description", []types.MarketParameterChange{
			types.NewMarketParameterChange(1, utils.ParseDec("0.001"), utils.ParseDec("0.002")),
			types.NewMarketParameterChange(2, utils.ParseDec("-0.0015"), utils.ParseDec("0.003")),
		})
	fmt.Println(p.String())

	// Output:
	// Market Parameter Change Proposal:
	//   Title:       Title
	//   Description: Description
	//   Changes:
	//     Market Parameter Change:
	//       Market Id:      1
	//       Maker Fee Rate: 0.001000000000000000
	//       Taker Fee Rate: 0.002000000000000000
	//     Market Parameter Change:
	//       Market Id:      2
	//       Maker Fee Rate: -0.001500000000000000
	//       Taker Fee Rate: 0.003000000000000000
}

func TestMarketParameterChangeProposal_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(p *types.MarketParameterChangeProposal)
		expectedErr string
	}{
		{
			"happy case",
			func(p *types.MarketParameterChangeProposal) {},
			"",
		},
		{
			"empty title",
			func(p *types.MarketParameterChangeProposal) {
				p.Title = ""
			},
			"proposal title cannot be blank: invalid proposal content",
		},
		{
			"empty description",
			func(p *types.MarketParameterChangeProposal) {
				p.Description = ""
			},
			"proposal description cannot be blank: invalid proposal content",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewMarketParameterChangeProposal(
				"Title", "Description", []types.MarketParameterChange{
					types.NewMarketParameterChange(
						1, utils.ParseDec("0.001"), utils.ParseDec("0.003")),
				})
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

func TestMarketParameterChange_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		change      types.MarketParameterChange
		expectedErr string
	}{
		{
			"happy case",
			types.NewMarketParameterChange(
				1, utils.ParseDec("0.001"), utils.ParseDec("0.003")),
			"",
		},
		{
			"invalid market id",
			types.NewMarketParameterChange(
				0, utils.ParseDec("0.001"), utils.ParseDec("0.003")),
			"market id must not be 0: invalid request",
		},
		{
			"too high maker fee rate",
			types.NewMarketParameterChange(
				1, utils.ParseDec("1.01"), utils.ParseDec("0.003")),
			"maker fee rate must be in range [-0.003000000000000000, 1]: 1.010000000000000000: invalid request",
		},
		{
			"too low maker fee rate",
			types.NewMarketParameterChange(
				1, utils.ParseDec("-1.01"), utils.ParseDec("0.003")),
			"maker fee rate must be in range [-0.003000000000000000, 1]: -1.010000000000000000: invalid request",
		},
		{
			"negative taker fee rate",
			types.NewMarketParameterChange(
				1, utils.ParseDec("0.001"), utils.ParseDec("-0.001")),
			"taker fee rate must be in range [0, 1]: -0.001000000000000000: invalid request",
		},
		{
			"too high taker fee rate",
			types.NewMarketParameterChange(
				1, utils.ParseDec("0.001"), utils.ParseDec("1.01")),
			"taker fee rate must be in range [0, 1]: 1.010000000000000000: invalid request",
		},
		{
			"-(maker fee rate) > taker fee rate",
			types.NewMarketParameterChange(
				1, utils.ParseDec("-0.003"), utils.ParseDec("0.001")),
			"maker fee rate must be in range [-0.001000000000000000, 1]: -0.003000000000000000: invalid request",
		},
		{
			"-(maker fee rate) <= taker fee rate",
			types.NewMarketParameterChange(
				1, utils.ParseDec("-0.001"), utils.ParseDec("0.001")),
			"",
		},
		{
			"maker fee rate > taker fee rate",
			types.NewMarketParameterChange(
				1, utils.ParseDec("0.002"), utils.ParseDec("0.001")),
			"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.change.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
