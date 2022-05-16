package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/crescent-network/crescent/x/liquidity/amm"
)

// BulkSendCoinsOperation holds a list of SendCoins operations for bulk execution.
type BulkSendCoinsOperation struct {
	Inputs  []banktypes.Input
	Outputs []banktypes.Output
}

// NewBulkSendCoinsOperation returns an empty BulkSendCoinsOperation.
func NewBulkSendCoinsOperation() *BulkSendCoinsOperation {
	return &BulkSendCoinsOperation{
		Inputs:  []banktypes.Input{},
		Outputs: []banktypes.Output{},
	}
}

// QueueSendCoins queues a BankKeeper.SendCoins operation for later execution.
func (op *BulkSendCoinsOperation) QueueSendCoins(fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) {
	if amt.IsValid() && !amt.IsZero() {
		op.Inputs = append(op.Inputs, banktypes.NewInput(fromAddr, amt))
		op.Outputs = append(op.Outputs, banktypes.NewOutput(toAddr, amt))
	}
}

// Run runs BankKeeper.InputOutputCoins once for queued operations.
func (op *BulkSendCoinsOperation) Run(ctx sdk.Context, bankKeeper BankKeeper) error {
	if len(op.Inputs) > 0 && len(op.Outputs) > 0 {
		return bankKeeper.InputOutputCoins(ctx, op.Inputs, op.Outputs)
	}
	return nil
}

// IsTooSmallOrderAmount returns whether the order amount is too small for
// matching, based on the order price.
func IsTooSmallOrderAmount(amt sdk.Int, price sdk.Dec) bool {
	return amt.LT(amm.MinCoinAmount) || price.MulInt(amt).LT(amm.MinCoinAmount.ToDec())
}

// OrderBookBasePrice returns the base(middle) price of the order book.
func OrderBookBasePrice(ov amm.OrderView, tickPrec int) (sdk.Dec, bool) {
	highestBuyPrice, foundHighestBuyPrice := ov.HighestBuyPrice()
	lowestSellPrice, foundLowestSellPrice := ov.LowestSellPrice()
	switch {
	case !foundHighestBuyPrice && !foundLowestSellPrice:
		return sdk.Dec{}, false
	case foundHighestBuyPrice && foundLowestSellPrice:
		matchPrice, found := amm.FindMatchPrice(ov, tickPrec)
		if !found {
			return amm.RoundPrice(highestBuyPrice.Add(lowestSellPrice).QuoInt64(2), tickPrec), true
		}
		return matchPrice, true
	case foundHighestBuyPrice:
		return highestBuyPrice, true
	default: // foundLowestSellPrice
		return lowestSellPrice, true
	}
}

// MakeOrderBookResponse returns OrderBookResponse from given inputs.
func MakeOrderBookResponse(ov amm.OrderView, basePrice sdk.Dec, tickPrec, numTicks int) OrderBookResponse {
	ammTickPrec := amm.TickPrecision(tickPrec)
	tickGap := ammTickPrec.TickGap(
		ammTickPrec.TickFromIndex(
			ammTickPrec.TickToIndex(ammTickPrec.PriceToDownTick(basePrice)) + (numTicks - 1)))
	matchableAmt := sdk.MinInt(ov.BuyAmountOver(basePrice), ov.SellAmountUnder(basePrice))

	makeTicks := func(dir OrderDirection) []OrderBookTickResponse {
		startPrice := FitPriceToGap(basePrice, tickGap)
		switch dir {
		case OrderDirectionBuy:
			if !ov.BuyAmountOver(startPrice).Sub(matchableAmt).IsPositive() {
				startPrice = startPrice.Sub(tickGap)
			}
		case OrderDirectionSell:
			if !ov.SellAmountUnder(startPrice).Sub(matchableAmt).IsPositive() {
				startPrice = startPrice.Add(tickGap)
			}
		}

		var ticks []OrderBookTickResponse
		accAmt := matchableAmt
		tick := startPrice
		for i := 0; i < numTicks && !tick.IsNegative(); i++ {
			var amt sdk.Int
			switch dir {
			case OrderDirectionBuy:
				amt = ov.BuyAmountOver(tick)
			case OrderDirectionSell:
				amt = ov.SellAmountUnder(tick)
			}
			amt = amt.Sub(accAmt)
			if amt.IsPositive() {
				ticks = append(ticks, OrderBookTickResponse{
					Price:           tick,
					UserOrderAmount: amt,
					PoolOrderAmount: sdk.ZeroInt(),
				})
				accAmt = accAmt.Add(amt)
			}
			switch dir {
			case OrderDirectionBuy:
				tick = tick.Sub(tickGap)
			case OrderDirectionSell:
				tick = tick.Add(tickGap)
			}
		}
		return ticks
	}

	buys, sells := makeTicks(OrderDirectionBuy), makeTicks(OrderDirectionSell)
	// Reverse sell ticks.
	for left, right := 0, len(sells)-1; left < right; left, right = left+1, right-1 {
		sells[left], sells[right] = sells[right], sells[left]
	}

	return OrderBookResponse{
		TickPrecision: uint32(tickPrec),
		Buys:          buys,
		Sells:         sells,
	}
}

// FitPriceToGap fits price into given unit gap.
func FitPriceToGap(price, gap sdk.Dec) sdk.Dec {
	b := price.BigInt()
	b.Quo(b, gap.BigInt()).Mul(b, gap.BigInt())
	return sdk.NewDecFromBigIntWithPrec(b, sdk.Precision)
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
