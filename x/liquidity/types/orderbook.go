package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
)

func OrderBookBasePrice(ov amm.OrderView, tickPrec int) (sdk.Dec, bool) {
	highestBuyPrice, foundHighestBuyPrice := ov.HighestBuyPrice()
	lowestSellPrice, foundLowestSellPrice := ov.LowestSellPrice()

	switch {
	case foundHighestBuyPrice && foundLowestSellPrice:
		return amm.RoundPrice(highestBuyPrice.Add(lowestSellPrice).QuoInt64(2), tickPrec), true
	case foundHighestBuyPrice:
		return highestBuyPrice, true
	case foundLowestSellPrice:
		return lowestSellPrice, true
	default: // not found
		return sdk.Dec{}, false
	}
}

func MakeOrderBookResponse(ov amm.OrderView, tickPrec, numTicks int) OrderBookResponse {
	ammTickPrec := amm.TickPrecision(tickPrec)
	resp := OrderBookResponse{TickPrecision: uint32(tickPrec)}

	highestBuyPrice, foundHighestBuyPrice := ov.HighestBuyPrice()
	lowestSellPrice, foundLowestSellPrice := ov.LowestSellPrice()

	var tickGapBasePrice sdk.Dec
	if foundLowestSellPrice {
		tickGapBasePrice = lowestSellPrice
	} else {
		tickGapBasePrice = highestBuyPrice
	}
	tickGap := ammTickPrec.TickGap(
		ammTickPrec.TickFromIndex(
			ammTickPrec.TickToIndex(tickGapBasePrice) + (numTicks - 1)))

	if foundHighestBuyPrice {
		startPrice := FitPriceToTickGap(highestBuyPrice, tickGap, true)
		currentPrice := startPrice
		accAmt := sdk.ZeroInt()
		for i := 0; i < numTicks && !currentPrice.IsNegative(); i++ {
			amt := ov.BuyAmountOver(currentPrice, true).Sub(accAmt)
			if amt.IsPositive() {
				resp.Buys = append(resp.Buys, OrderBookTickResponse{
					Price:           currentPrice,
					UserOrderAmount: amt,
					PoolOrderAmount: sdk.ZeroInt(),
				})
				accAmt = accAmt.Add(amt)
			}
			currentPrice = currentPrice.Sub(tickGap)
		}
	}
	if foundLowestSellPrice {
		startPrice := FitPriceToTickGap(lowestSellPrice, tickGap, false)
		currentPrice := startPrice
		accAmt := sdk.ZeroInt()
		for i := 0; i < numTicks; i++ {
			amt := ov.SellAmountUnder(currentPrice, true).Sub(accAmt)
			if amt.IsPositive() {
				resp.Sells = append(resp.Sells, OrderBookTickResponse{
					Price:           currentPrice,
					UserOrderAmount: amt,
					PoolOrderAmount: sdk.ZeroInt(),
				})
				accAmt = accAmt.Add(amt)
			}
			currentPrice = currentPrice.Add(tickGap)
		}
		// Reverse sell ticks.
		for l, r := 0, len(resp.Sells)-1; l < r; l, r = l+1, r-1 {
			resp.Sells[l], resp.Sells[r] = resp.Sells[r], resp.Sells[l]
		}
	}

	return resp
}

// PrintOrderBookResponse prints out OrderBookResponse in human-readable form.
func PrintOrderBookResponse(ob OrderBookResponse, basePrice sdk.Dec) {
	fmt.Println("+------------------------------------------------------------------------+")
	for _, tick := range ob.Sells {
		fmt.Printf("| %18s | %28s |                    |\n", tick.UserOrderAmount, tick.Price.String())
	}
	fmt.Println("|------------------------------------------------------------------------|")
	fmt.Printf("|                      %28s                      |\n", basePrice.String())
	fmt.Println("|------------------------------------------------------------------------|")
	for _, tick := range ob.Buys {
		fmt.Printf("|                    | %28s | %-18s |\n", tick.Price.String(), tick.UserOrderAmount)
	}
	fmt.Println("+------------------------------------------------------------------------+")
}

// FitPriceToTickGap fits price into given tick gap.
func FitPriceToTickGap(price, gap sdk.Dec, down bool) sdk.Dec {
	b := price.BigInt()
	b.Quo(b, gap.BigInt()).Mul(b, gap.BigInt())
	tick := sdk.NewDecFromBigIntWithPrec(b, sdk.Precision)
	if !down && !tick.Equal(price) {
		tick = tick.Add(gap)
	}
	return tick
}
