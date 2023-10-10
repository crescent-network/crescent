package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func ExampleMarketParameterChange_String() {
	p := types.NewMarketParameterChangeProposal(
		"Title", "Description", []types.MarketParameterChange{
			types.NewMarketParameterChange(
				1,
				types.NewFees(
					utils.ParseDec("0.001"), utils.ParseDec("0.002"), utils.ParseDec("0.5")),
				types.NewAmountLimits(sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30)),
				types.NewAmountLimits(sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30))),
			types.NewMarketParameterChange(
				2,
				types.NewFees(
					utils.ParseDec("0.0015"), utils.ParseDec("0.003"), utils.ParseDec("0.5")),
				types.NewAmountLimits(sdk.NewInt(10), sdk.NewIntWithDecimal(1, 20)),
				types.NewAmountLimits(sdk.NewInt(100), sdk.NewIntWithDecimal(1, 22))),
		})
	fmt.Println(p.String())

	// Output:
	// Market Parameter Change Proposal:
	//   Title:       Title
	//   Description: Description
	//   Changes:
	//     Market Parameter Change:
	//       Market Id: 1
	//       Fees:
	//         Maker Fee Rate:         0.001000000000000000
	//         Taker Fee Rate:         0.002000000000000000
	//         Order Source Fee Ratio: 0.500000000000000000
	//       Order Quantity Limits:
	//         Min: 10000
	//         Max: 1000000000000000000000000000000
	//       Order Quote Limits:
	//         Min: 10000
	//         Max: 1000000000000000000000000000000
	//     Market Parameter Change:
	//       Market Id: 2
	//       Fees:
	//         Maker Fee Rate:         0.001500000000000000
	//         Taker Fee Rate:         0.003000000000000000
	//         Order Source Fee Ratio: 0.500000000000000000
	//       Order Quantity Limits:
	//         Min: 10
	//         Max: 100000000000000000000
	//       Order Quote Limits:
	//         Min: 100
	//         Max: 10000000000000000000000
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
						1,
						types.NewFees(
							utils.ParseDec("0.001"), utils.ParseDec("0.003"), utils.ParseDec("0.5")),
						types.NewAmountLimits(sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30)),
						types.NewAmountLimits(sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30))),
				})
			require.Equal(t, types.ProposalTypeMarketParameterChange, p.ProposalType())
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
		malleate    func(change *types.MarketParameterChange)
		expectedErr string
	}{
		{
			"happy case",
			func(change *types.MarketParameterChange) {},
			"",
		},
		{
			"invalid market id",
			func(change *types.MarketParameterChange) {
				change.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"invalid fees",
			func(change *types.MarketParameterChange) {
				change.Fees = types.NewFees(
					utils.ParseDec("1.01"), utils.ParseDec("0.003"), utils.ParseDec("0.5"))
			},
			"maker fee rate must be in range [0, 1]: 1.010000000000000000: invalid request",
		},
		{
			"invalid order quantity limits",
			func(change *types.MarketParameterChange) {
				change.OrderQuantityLimits = types.NewAmountLimits(sdk.NewInt(10000), sdk.NewInt(-10000))
			},
			"invalid order quantity limits: the maximum value must not be negative: -10000: invalid request",
		},
		{
			"invalid order quote limits",
			func(change *types.MarketParameterChange) {
				change.OrderQuoteLimits = types.NewAmountLimits(sdk.NewInt(-10000), sdk.NewInt(10000))
			},
			"invalid order quote limits: the minimum value must not be negative: -10000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			change := types.NewMarketParameterChange(
				1, types.NewFees(
					utils.ParseDec("0.0015"), utils.ParseDec("0.003"), utils.ParseDec("0.7")),
				types.NewAmountLimits(sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30)),
				types.NewAmountLimits(sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30)))
			tc.malleate(&change)
			err := change.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
