package types_test

import (
	"fmt"

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
