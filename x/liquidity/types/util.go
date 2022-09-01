package types

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
)

type sendCoinsTxKey struct {
	from, to string
}

type sendCoinsTx struct {
	from, to sdk.AccAddress
	amt      sdk.Coins
}

// BulkSendCoinsOperation holds a list of SendCoins operations for bulk execution.
type BulkSendCoinsOperation struct {
	txSet map[sendCoinsTxKey]*sendCoinsTx
	txs   []*sendCoinsTx
}

// NewBulkSendCoinsOperation returns an empty BulkSendCoinsOperation.
func NewBulkSendCoinsOperation() *BulkSendCoinsOperation {
	return &BulkSendCoinsOperation{
		txSet: map[sendCoinsTxKey]*sendCoinsTx{},
	}
}

// QueueSendCoins queues a BankKeeper.SendCoins operation for later execution.
func (op *BulkSendCoinsOperation) QueueSendCoins(fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) {
	if amt.IsValid() && !amt.IsZero() {
		txKey := sendCoinsTxKey{fromAddr.String(), toAddr.String()}
		tx, ok := op.txSet[txKey]
		if !ok {
			tx = &sendCoinsTx{fromAddr, toAddr, sdk.Coins{}}
			op.txSet[txKey] = tx
			op.txs = append(op.txs, tx)
		}
		tx.amt = tx.amt.Add(amt...)
	}
}

// Run runs BankKeeper.InputOutputCoins once for queued operations.
func (op *BulkSendCoinsOperation) Run(ctx sdk.Context, bankKeeper BankKeeper) error {
	if len(op.txs) > 0 {
		var (
			inputs  []banktypes.Input
			outputs []banktypes.Output
		)
		for _, tx := range op.txs {
			inputs = append(inputs, banktypes.NewInput(tx.from, tx.amt))
			outputs = append(outputs, banktypes.NewOutput(tx.to, tx.amt))
		}
		return bankKeeper.InputOutputCoins(ctx, inputs, outputs)
	}
	return nil
}

// NewPoolResponse returns a new PoolResponse from given information.
func NewPoolResponse(pool Pool, rx, ry sdk.Coin, poolCoinSupply sdk.Int) PoolResponse {
	var price *sdk.Dec
	if !pool.Disabled {
		p := pool.AMMPool(rx.Amount, ry.Amount, sdk.Int{}).Price()
		price = &p
	}
	return PoolResponse{
		Type:           pool.Type,
		Id:             pool.Id,
		PairId:         pool.PairId,
		Creator:        pool.Creator,
		ReserveAddress: pool.ReserveAddress,
		PoolCoinDenom:  pool.PoolCoinDenom,
		PoolCoinSupply: poolCoinSupply,
		MinPrice:       pool.MinPrice,
		MaxPrice:       pool.MaxPrice,
		Price:          price,
		Balances: PoolBalances{
			BaseCoin:  ry,
			QuoteCoin: rx,
		},
		LastDepositRequestId:  pool.LastDepositRequestId,
		LastWithdrawRequestId: pool.LastWithdrawRequestId,
		Disabled:              pool.Disabled,
	}
}

// IsTooSmallOrderAmount returns whether the order amount is too small for
// matching, based on the order price.
func IsTooSmallOrderAmount(amt sdk.Int, price sdk.Dec) bool {
	return amt.LT(amm.MinCoinAmount) || price.MulInt(amt).LT(amm.MinCoinAmount.ToDec())
}

// PriceLimits returns the lowest and the highest price limits with given last price
// and price limit ratio.
func PriceLimits(lastPrice, priceLimitRatio sdk.Dec, tickPrec int) (lowestPrice, highestPrice sdk.Dec) {
	lowestPrice = amm.PriceToUpTick(lastPrice.Mul(sdk.OneDec().Sub(priceLimitRatio)), tickPrec)
	highestPrice = amm.PriceToDownTick(lastPrice.Mul(sdk.OneDec().Add(priceLimitRatio)), tickPrec)
	return
}

func NewMMOrderIndex(orderer sdk.AccAddress, pairId uint64, orderIds []uint64) MMOrderIndex {
	return MMOrderIndex{
		Orderer:  orderer.String(),
		PairId:   pairId,
		OrderIds: orderIds,
	}
}

func (index MMOrderIndex) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(index.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}

// MMOrderTick holds information about each tick's price and amount of an MMOrder.
type MMOrderTick struct {
	OfferCoinAmount sdk.Int
	Price           sdk.Dec
	Amount          sdk.Int
}

// MMOrderTicks returns fairly distributed tick information with given parameters.
func MMOrderTicks(dir OrderDirection, minPrice, maxPrice sdk.Dec, amt sdk.Int, maxNumTicks, tickPrec int) (ticks []MMOrderTick) {
	ammDir := amm.OrderDirection(dir)
	if minPrice.Equal(maxPrice) {
		return []MMOrderTick{{OfferCoinAmount: amm.OfferCoinAmount(ammDir, minPrice, amt), Price: minPrice, Amount: amt}}
	}
	gap := maxPrice.Sub(minPrice).QuoInt64(int64(maxNumTicks - 1))
	switch dir {
	case OrderDirectionBuy:
		var prevP sdk.Dec
		for i := 0; i < maxNumTicks-1; i++ {
			p := amm.PriceToDownTick(minPrice.Add(gap.MulInt64(int64(i))), tickPrec)
			if prevP.IsNil() || !p.Equal(prevP) {
				ticks = append(ticks, MMOrderTick{
					Price: p,
				})
				prevP = p
			}
		}
		tickAmt := amt.QuoRaw(int64(len(ticks) + 1))
		for i := range ticks {
			ticks[i].Amount = tickAmt
			ticks[i].OfferCoinAmount = amm.OfferCoinAmount(ammDir, ticks[i].Price, ticks[i].Amount)
			amt = amt.Sub(tickAmt)
		}
		ticks = append(ticks, MMOrderTick{
			OfferCoinAmount: amm.OfferCoinAmount(ammDir, maxPrice, amt),
			Price:           maxPrice,
			Amount:          amt,
		})
	case OrderDirectionSell:
		var prevP sdk.Dec
		for i := 0; i < maxNumTicks-1; i++ {
			p := amm.PriceToUpTick(maxPrice.Sub(gap.MulInt64(int64(i))), tickPrec)
			if prevP.IsNil() || !p.Equal(prevP) {
				ticks = append(ticks, MMOrderTick{
					Price: p,
				})
				prevP = p
			}
		}
		tickAmt := amt.QuoRaw(int64(len(ticks) + 1))
		for i := range ticks {
			ticks[i].Amount = tickAmt
			ticks[i].OfferCoinAmount = amm.OfferCoinAmount(ammDir, ticks[i].Price, ticks[i].Amount)
			amt = amt.Sub(tickAmt)
		}
		ticks = append(ticks, MMOrderTick{
			OfferCoinAmount: amm.OfferCoinAmount(ammDir, minPrice, amt),
			Price:           minPrice,
			Amount:          amt,
		})
	}
	return
}

// FormatUint64s returns comma-separated string representation of
// a slice of uint64.
func FormatUint64s(us []uint64) (s string) {
	ss := make([]string, 0, len(us))
	for _, u := range us {
		ss = append(ss, strconv.FormatUint(u, 10))
	}
	return strings.Join(ss, ",")
}
