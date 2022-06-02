package amm

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/types"
)

func parseOrders(s string) []Order {
	var orders []Order
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		chunks := strings.Split(line, "|")
		if len(chunks) != 3 {
			panic(fmt.Errorf("wrong number of chunks in %q: %d", line, len(chunks)))
		}
		price := sdk.MustNewDecFromStr(strings.TrimSpace(chunks[1]))
		parseAmounts := func(s string) []sdk.Int {
			var amounts []sdk.Int
			for _, amtChunk := range strings.Split(s, " ") {
				amtChunk = strings.TrimSpace(amtChunk)
				if amtChunk == "" {
					continue
				}
				amt, ok := sdk.NewIntFromString(amtChunk)
				if !ok {
					panic(fmt.Errorf("invalid amount: %s", amtChunk))
				}
				amounts = append(amounts, amt)
			}
			return amounts
		}
		for _, amt := range parseAmounts(chunks[0]) {
			orders = append(orders, newOrder(Sell, price, amt))
		}
		for _, amt := range parseAmounts(chunks[2]) {
			orders = append(orders, newOrder(Buy, price, amt))
		}
	}
	return orders
}

func TestInstantMatch(t *testing.T) {
	orders := parseOrders(`
        | 1.2 | 5 7
5       | 0.9 |
6 3     | 0.8 | 
4       | 0.7 |
`)
	ob := NewOrderBook(orders...)
	ctx := NewMatchContext()
	matched := ob.InstantMatch(ctx, utils.ParseDec("1.0"))
	require.True(t, matched)
	for _, order := range orders {
		fmt.Printf("%s %s at %s\n", order.GetDirection(), order.GetAmount(), order.GetPrice())
		if _, ok := ctx[order]; !ok {
			continue
		}
		for _, record := range ctx[order].MatchRecords {
			fmt.Printf("  match %s at %s\n", record.Amount, record.Price)
		}
	}
}
