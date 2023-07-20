package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type contextKey string

const EscrowContextKey = contextKey("escrow")

func WithEscrow(ctx context.Context, escrow *Escrow) context.Context {
	return context.WithValue(ctx, EscrowContextKey, escrow)
}

func GetEscrow(ctx context.Context) *Escrow {
	val := ctx.Value(EscrowContextKey)
	if val == nil {
		panic("escrow not found in context")
	}
	return val.(*Escrow)
}

type Escrow struct {
	escrowAddr sdk.AccAddress
	deltas     map[string]sdk.DecCoins
	addrs      []sdk.AccAddress
}

func NewEscrow(escrowAddr sdk.AccAddress) *Escrow {
	return &Escrow{
		escrowAddr: escrowAddr,
		deltas:     map[string]sdk.DecCoins{},
	}
}

func (e *Escrow) Lock(addr sdk.AccAddress, amt ...sdk.DecCoin) {
	saddr := addr.String()
	before, ok := e.deltas[saddr]
	if !ok {
		e.addrs = append(e.addrs, addr)
	}
	e.deltas[saddr], _ = before.SafeSub(amt)
}

func (e *Escrow) Unlock(addr sdk.AccAddress, amt ...sdk.DecCoin) {
	saddr := addr.String()
	before, ok := e.deltas[saddr]
	if !ok {
		e.addrs = append(e.addrs, addr)
	}
	e.deltas[saddr] = before.Add(amt...)
}

func (e *Escrow) Pays(addr sdk.AccAddress) sdk.Coins {
	var pays sdk.Coins
	for _, decCoin := range e.deltas[addr.String()] {
		if decCoin.IsNegative() {
			coin := sdk.NewCoin(decCoin.Denom, decCoin.Amount.Neg().Ceil().TruncateInt())
			pays = pays.Add(coin)
		}
	}
	return pays
}

func (e *Escrow) Receives(addr sdk.AccAddress) sdk.Coins {
	var receives sdk.Coins
	for _, decCoin := range e.deltas[addr.String()] {
		if decCoin.IsPositive() {
			coin, _ := decCoin.TruncateDecimal()
			receives = receives.Add(coin)
		}
	}
	return receives
}

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
