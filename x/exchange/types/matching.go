package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

type MemOrderType int

const (
	UserMemOrder MemOrderType = iota + 1
	OrderSourceMemOrder
)

type MemOrder struct {
	typ              MemOrderType
	orderId          uint64 // 0 for orders from order sources
	msgHeight        int64  // 0 for orders from order sources
	ordererAddr      sdk.AccAddress
	isBuy            bool
	price            sdk.Dec
	qty              sdk.Dec
	openQty          sdk.Dec
	remainingDeposit sdk.Dec
	executedQty      sdk.Dec
	paid             sdk.DecCoin
	received         sdk.DecCoins
	isMatched        bool
	source           OrderSource // nil for orders from users
}

func NewUserMemOrder(market Market, order Order) *MemOrder {
	payDenom, _ := market.PayReceiveDenoms(order.IsBuy)
	return &MemOrder{
		typ:              UserMemOrder,
		orderId:          order.Id,
		msgHeight:        order.MsgHeight,
		ordererAddr:      order.MustGetOrdererAddress(),
		isBuy:            order.IsBuy,
		price:            order.Price,
		qty:              order.Quantity,
		openQty:          order.OpenQuantity,
		remainingDeposit: order.RemainingDeposit,
		executedQty:      utils.ZeroDec,
		paid:             sdk.NewDecCoinFromDec(payDenom, utils.ZeroDec),
		received:         nil,
		isMatched:        false,
		source:           nil,
	}
}

func NewOrderSourceMemOrder(
	market Market, source OrderSource, ordererAddr sdk.AccAddress,
	isBuy bool, price, qty, deposit sdk.Dec) *MemOrder {
	payDenom, _ := market.PayReceiveDenoms(isBuy)
	return &MemOrder{
		typ:              OrderSourceMemOrder,
		orderId:          0,
		msgHeight:        0,
		ordererAddr:      ordererAddr,
		isBuy:            isBuy,
		price:            price,
		qty:              qty,
		openQty:          qty,
		remainingDeposit: deposit,
		executedQty:      utils.ZeroDec,
		paid:             sdk.NewDecCoinFromDec(payDenom, utils.ZeroDec),
		received:         nil,
		isMatched:        false,
		source:           source,
	}
}

func (order *MemOrder) executableQty(price sdk.Dec) sdk.Dec {
	if order.isBuy {
		return sdk.MinDec(
			order.openQty.Sub(order.executedQty),
			order.remainingDeposit.QuoTruncate(price))
	}
	return sdk.MinDec(order.openQty.Sub(order.executedQty), order.remainingDeposit)
}

func (order *MemOrder) HasPriorityOver(other *MemOrder) bool {
	if !order.price.Equal(other.price) { // sanity check
		panic(fmt.Sprintf("orders with different price: %s != %s", order.price, other.price))
	}
	if !order.qty.Equal(other.qty) {
		return order.qty.GT(other.qty)
	}
	switch {
	case order.typ == UserMemOrder && other.typ == UserMemOrder:
		return order.orderId < other.orderId
	case order.typ == UserMemOrder && other.typ == OrderSourceMemOrder:
		return true
	case order.typ == OrderSourceMemOrder && other.typ == UserMemOrder:
		return false
	default:
		return order.source.Name() < other.source.Name() // lexicographical ordering
	}
}

type MemOrderBookPriceLevel struct {
	isBuy  bool
	price  sdk.Dec
	orders []*MemOrder
}

func NewMemOrderBookPriceLevel(order *MemOrder) *MemOrderBookPriceLevel {
	return &MemOrderBookPriceLevel{order.isBuy, order.price, []*MemOrder{order}}
}

type MemOrderBookSide struct {
	isBuy  bool
	levels []*MemOrderBookPriceLevel
}

func NewMemOrderBookSide(isBuy bool) *MemOrderBookSide {
	return &MemOrderBookSide{isBuy: isBuy}
}

func (obs *MemOrderBookSide) AddOrder(order *MemOrder) {
	if order.isBuy != obs.isBuy { // sanity check
		panic("inconsistent order isBuy")
	}
	i := sort.Search(len(obs.levels), func(i int) bool {
		if obs.isBuy {
			return obs.levels[i].price.LTE(order.price)
		}
		return obs.levels[i].price.GTE(order.price)
	})
	if i < len(obs.levels) && obs.levels[i].price.Equal(order.price) {
		obs.levels[i].orders = append(obs.levels[i].orders, order)
	} else {
		// Insert a new level.
		newLevels := make([]*MemOrderBookPriceLevel, len(obs.levels)+1)
		copy(newLevels[:i], obs.levels[:i])
		newLevels[i] = NewMemOrderBookPriceLevel(order)
		copy(newLevels[i+1:], obs.levels[i:])
		obs.levels = newLevels
	}
}

type MatchingContext struct {
	baseDenom    string
	quoteDenom   string
	makerFeeRate sdk.Dec
	takerFeeRate sdk.Dec
}

func NewMatchingContext(market Market, halveFees bool) *MatchingContext {
	makerFeeRate := market.MakerFeeRate
	takerFeeRate := market.TakerFeeRate
	if halveFees {
		makerFeeRate = makerFeeRate.QuoInt64(2)
		takerFeeRate = takerFeeRate.QuoInt64(2)
	}
	return &MatchingContext{
		baseDenom:    market.BaseDenom,
		quoteDenom:   market.QuoteDenom,
		makerFeeRate: makerFeeRate,
		takerFeeRate: takerFeeRate,
	}
}

func (ctx *MatchingContext) FillOrder(order MemOrder, qty, price sdk.Dec, isMaker bool) {
	// TODO: refactor code
	if qty.GT(order.executableQty(price)) { // sanity check
		panic("open quantity is less than quantity")
	}
	negativeMakerFeeRate := ctx.makerFeeRate.IsNegative()
	order.executedQty = order.executedQty.Add(qty)
	if order.isBuy {
		paid := QuoteAmount(true, price, qty)
		order.paid.Amount = order.paid.Amount.Add(paid)
		order.remainingDeposit = order.remainingDeposit.Sub(paid)
		if order.typ == OrderSourceMemOrder || (isMaker && negativeMakerFeeRate) {
			order.received = order.received.Add(sdk.NewDecCoinFromDec(ctx.baseDenom, qty))
		} else {
			if isMaker {
				order.received = order.received.Add(
					sdk.NewDecCoinFromDec(
						ctx.baseDenom,
						utils.OneDec.Sub(ctx.makerFeeRate).MulTruncate(qty)))
			} else {
				order.received = order.received.Add(
					sdk.NewDecCoinFromDec(
						ctx.baseDenom,
						utils.OneDec.Sub(ctx.takerFeeRate).MulTruncate(qty)))
			}
		}
		if isMaker && negativeMakerFeeRate {
			order.received = order.received.Add(
				sdk.NewDecCoinFromDec(
					ctx.quoteDenom,
					ctx.makerFeeRate.Neg().MulTruncate(paid)))
		}
	} else {
		order.paid.Amount = order.paid.Amount.Add(qty)
		order.remainingDeposit = order.remainingDeposit.Sub(qty)
		quote := QuoteAmount(false, price, qty)
		if order.typ == OrderSourceMemOrder || (isMaker && negativeMakerFeeRate) {
			order.received = order.received.Add(sdk.NewDecCoinFromDec(ctx.quoteDenom, quote))
		} else {
			if isMaker {
				order.received = order.received.Add(
					sdk.NewDecCoinFromDec(
						ctx.quoteDenom,
						utils.OneDec.Sub(ctx.makerFeeRate).MulTruncate(quote)))
			} else {
				order.received = order.received.Add(
					sdk.NewDecCoinFromDec(
						ctx.quoteDenom,
						utils.OneDec.Sub(ctx.takerFeeRate).MulTruncate(quote)))
			}
		}
		if isMaker && negativeMakerFeeRate {
			order.received = order.received.Add(
				sdk.NewDecCoinFromDec(
					ctx.baseDenom,
					ctx.makerFeeRate.Neg().MulTruncate(qty)))
		}
	}
	order.isMatched = true
}

func (ctx *MatchingContext) FillOrders(orders []*MemOrder, qty, price sdk.Dec, isMaker bool) {
	totalExecutableQty := TotalExecutableQuantity(orders, price)
	if totalExecutableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	}
	if qty.LT(totalExecutableQty) { // partial matches
		sort.Slice(orders, func(i, j int) bool {
			return orders[i].HasPriorityOver(orders[j])
		})
	}
	totalExecQty := utils.ZeroDec
	// First, distribute quantity evenly.
	for _, order := range orders {
		remainingQty := qty.Sub(totalExecQty)
		if remainingQty.IsZero() {
			break
		}
		executableQty := order.ExecutableQuantity(price)
		if executableQty.IsZero() {
			continue
		}
		ratio := order.Quantity.QuoTruncate(totalExecutableQty)
		execQty := sdk.MinDec(
			remainingQty,
			sdk.MinDec(executableQty, ratio.MulTruncate(order.Quantity)))
		if execQty.IsPositive() {
			market.FillTempOrder(order, execQty, price, isMaker, halveFees)
			totalExecQty = totalExecQty.Add(execQty)
		}
	}
	// Then, distribute remaining quantity based on priority.
	// TODO: sort?
	for _, order := range orders {
		remainingQty := qty.Sub(totalExecQty)
		if remainingQty.IsZero() {
			break
		}
		execQty := sdk.MinDec(remainingQty, order.ExecutableQuantity(price))
		if execQty.IsPositive() {
			market.FillTempOrder(order, execQty, price, isMaker, halveFees)
			totalExecQty = totalExecQty.Add(execQty)
		}
	}
}

func (ctx *MatchingContext) FillOrderBookPriceLevel(level *MemOrderBookPriceLevel, qty, price sdk.Dec, isMaker bool) {

}

func (ctx *MatchingContext) MatchOrderBookPriceLevels(
	levelA *MemOrderBookPriceLevel, isLevelAMaker bool,
	levelB *MemOrderBookPriceLevel, isLevelBMaker bool, price sdk.Dec) {

}
