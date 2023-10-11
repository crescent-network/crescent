package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Escrow is a structure used to facilitate coin transfers between
// escrow account and other accounts all at once.
type Escrow struct {
	escrowAddr sdk.AccAddress
	deltas     map[string]sdk.Coins // string(addr) => delta
	addrs      []sdk.AccAddress     // for ordered access on deltas
}

func NewEscrow(escrowAddr sdk.AccAddress) *Escrow {
	return &Escrow{
		escrowAddr: escrowAddr,
		deltas:     map[string]sdk.Coins{},
	}
}

func (e *Escrow) Escrow(addr sdk.AccAddress, amt ...sdk.Coin) {
	saddr := addr.String()
	before, ok := e.deltas[saddr]
	if !ok {
		e.addrs = append(e.addrs, addr)
	}
	e.deltas[saddr], _ = before.SafeSub(amt)
}

func (e *Escrow) Release(addr sdk.AccAddress, amt ...sdk.Coin) {
	saddr := addr.String()
	before, ok := e.deltas[saddr]
	if !ok {
		e.addrs = append(e.addrs, addr)
	}
	e.deltas[saddr] = before.Add(amt...)
}

// Pays returns how much coins an account would pay by summing negative
// balance diffs up.
func (e *Escrow) Pays(addr sdk.AccAddress) sdk.Coins {
	var pays sdk.Coins
	for _, coin := range e.deltas[addr.String()] {
		if coin.IsNegative() {
			coin.Amount = coin.Amount.Neg()
			pays = pays.Add(coin)
		}
	}
	return pays
}

// Receives returns how much coins an account would receive by summing positive
// balance diffs up.
func (e *Escrow) Receives(addr sdk.AccAddress) sdk.Coins {
	var receives sdk.Coins
	for _, coin := range e.deltas[addr.String()] {
		if coin.IsPositive() {
			receives = receives.Add(coin)
		}
	}
	return receives
}

// Transact runs the actual coin transactions between escrow account and
// other accounts.
func (e *Escrow) Transact(ctx sdk.Context, bankKeeper BankKeeper) error {
	escrow := e.escrowAddr.String()
	var (
		payInputs, receiveInputs   []banktypes.Input
		payOutputs, receiveOutputs []banktypes.Output
	)
	for _, addr := range e.addrs {
		saddr := addr.String()
		pays := e.Pays(addr)
		receives := e.Receives(addr)
		if !pays.IsZero() {
			payInputs = append(payInputs, banktypes.Input{Address: saddr, Coins: pays})
			payOutputs = append(payOutputs, banktypes.Output{Address: escrow, Coins: pays})
		}
		if !receives.IsZero() {
			receiveInputs = append(receiveInputs, banktypes.Input{Address: escrow, Coins: receives})
			receiveOutputs = append(receiveOutputs, banktypes.Output{Address: saddr, Coins: receives})
		}
	}
	if err := bankKeeper.InputOutputCoins(ctx, payInputs, payOutputs); err != nil {
		return err
	}
	return bankKeeper.InputOutputCoins(ctx, receiveInputs, receiveOutputs)
}
