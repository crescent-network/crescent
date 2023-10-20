package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

type Ledger struct {
	baseDenom, quoteDenom string
	addrDeltas            map[string]sdk.Coins // string(addr) => delta
	addrs                 []sdk.AccAddress     // keeps tracks the order of addresses

	baseDelta, quoteDelta       sdk.Int
	baseFeeDelta, quoteFeeDelta sdk.Int
}

func NewLedger(baseDenom, quoteDenom string) *Ledger {
	return &Ledger{
		baseDenom:     baseDenom,
		quoteDenom:    quoteDenom,
		addrDeltas:    map[string]sdk.Coins{},
		addrs:         nil,
		baseDelta:     utils.ZeroInt,
		quoteDelta:    utils.ZeroInt,
		baseFeeDelta:  utils.ZeroInt,
		quoteFeeDelta: utils.ZeroInt,
	}
}

func (ledger *Ledger) Pay(addr sdk.AccAddress, amt ...sdk.Coin) {
	addrStr := addr.String()
	// We saw a new address, append it to addrs.
	if _, ok := ledger.addrDeltas[addrStr]; !ok {
		ledger.addrs = append(ledger.addrs, addr)
	}
	// To allow negative result, we use SafeSub and ignore the second return value.
	ledger.addrDeltas[addrStr], _ = ledger.addrDeltas[addrStr].SafeSub(amt)
}

func (ledger *Ledger) Receive(addr sdk.AccAddress, amt ...sdk.Coin) {
	addrStr := addr.String()
	// We saw a new address, append it to addrs.
	if _, ok := ledger.addrDeltas[addrStr]; !ok {
		ledger.addrs = append(ledger.addrs, addr)
	}
	ledger.addrDeltas[addrStr] = ledger.addrDeltas[addrStr].Add(amt...)
}

func (ledger *Ledger) FeedMatchResult(isBuy bool, res MatchResult) {
	if isBuy {
		ledger.quoteDelta = ledger.quoteDelta.Add(res.Paid)
		ledger.baseDelta = ledger.baseDelta.Sub(res.Received)
		if res.FeePaid.IsPositive() {
			ledger.baseFeeDelta = ledger.baseFeeDelta.Add(res.FeePaid)
		} else if res.FeeReceived.IsPositive() {
			ledger.quoteFeeDelta = ledger.quoteFeeDelta.Sub(res.FeeReceived)
		}
	} else {
		ledger.baseDelta = ledger.baseDelta.Add(res.Paid)
		ledger.quoteDelta = ledger.quoteDelta.Sub(res.Received)
		if res.FeePaid.IsPositive() {
			ledger.quoteFeeDelta = ledger.quoteFeeDelta.Add(res.FeePaid)
		} else if res.FeeReceived.IsPositive() {
			ledger.baseFeeDelta = ledger.baseFeeDelta.Sub(res.FeeReceived)
		}
	}
}

// Transact runs the actual coin transactions between escrow and other addresses.
// Contract: do not call this method twice.
func (ledger *Ledger) Transact(
	ctx sdk.Context, bankKeeper BankKeeper, escrowAddr, feeCollectorAddr sdk.AccAddress) (quoteDust sdk.Coin, err error) {
	escrow := escrowAddr.String()
	var (
		payInputs, receiveInputs   []banktypes.Input
		payOutputs, receiveOutputs []banktypes.Output
	)
	for _, addr := range ledger.addrs {
		addrStr := addr.String()
		pays, receives := sdk.Coins{}, sdk.Coins{}
		for _, coin := range ledger.addrDeltas[addrStr] {
			if coin.IsPositive() {
				receives = receives.Add(coin)
			} else { // it means coin.IsNegative(), since there's no zero coin
				coin.Amount = coin.Amount.Neg() // negate the amount
				pays = pays.Add(coin)
			}
		}
		if !pays.IsZero() {
			payInputs = append(payInputs, banktypes.Input{Address: addrStr, Coins: pays})
			payOutputs = append(payOutputs, banktypes.Output{Address: escrow, Coins: pays})
		}
		if !receives.IsZero() {
			receiveInputs = append(receiveInputs, banktypes.Input{Address: escrow, Coins: receives})
			receiveOutputs = append(receiveOutputs, banktypes.Output{Address: addrStr, Coins: receives})
		}
	}
	if err = bankKeeper.InputOutputCoins(ctx, payInputs, payOutputs); err != nil {
		return
	}
	if err = bankKeeper.InputOutputCoins(ctx, receiveInputs, receiveOutputs); err != nil {
		return
	}
	if !ledger.baseDelta.Equal(ledger.baseFeeDelta) {
		err = fmt.Errorf("baseDelta must be same as baseFeeDelta: %s != %s",
			ledger.baseDelta, ledger.baseFeeDelta)
		return
	}
	fees := sdk.NewCoins(
		sdk.NewCoin(ledger.baseDenom, ledger.baseFeeDelta),
		sdk.NewCoin(ledger.quoteDenom, ledger.quoteFeeDelta))
	if fees.IsAllPositive() {
		if err = bankKeeper.SendCoins(ctx, escrowAddr, feeCollectorAddr, fees); err != nil {
			return
		}
	}
	quoteDust = sdk.NewCoin(ledger.quoteDenom, ledger.quoteDelta.Sub(ledger.quoteFeeDelta))
	return quoteDust, nil
}
