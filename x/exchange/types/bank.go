package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type BalanceDeltas struct {
	deltas map[string]sdk.DecCoins // string(addr) => delta
	addrs  []sdk.AccAddress
}

func NewBalanceDeltas() *BalanceDeltas {
	return &BalanceDeltas{
		deltas: map[string]sdk.DecCoins{},
	}
}

func (bd *BalanceDeltas) Add(addr sdk.AccAddress, delta ...sdk.DecCoin) {
	saddr := addr.String()
	before, ok := bd.deltas[saddr]
	if !ok {
		bd.addrs = append(bd.addrs, addr)
	}
	bd.deltas[saddr] = before.Add(delta...)
}

func (bd *BalanceDeltas) Sub(addr sdk.AccAddress, delta ...sdk.DecCoin) {
	saddr := addr.String()
	before, ok := bd.deltas[saddr]
	if !ok {
		bd.addrs = append(bd.addrs, addr)
	}
	bd.deltas[saddr], _ = before.SafeSub(delta)
}

func (bd *BalanceDeltas) Settle(ctx sdk.Context, bankKeeper BankKeeper) error {
	var (
		inputs  []banktypes.Input
		outputs []banktypes.Output
	)
	for _, addr := range bd.addrs {
		saddr := addr.String()
		var input, output sdk.Coins
		for _, decCoin := range bd.deltas[saddr] {
			var coin sdk.Coin
			if decCoin.IsNegative() {
				coin = sdk.NewCoin(decCoin.Denom, decCoin.Amount.Neg().Ceil().TruncateInt())
				input = input.Add(coin)
			} else {
				coin, _ = decCoin.TruncateDecimal()
				output = output.Add(coin)
			}
		}
		if !input.IsZero() {
			inputs = append(inputs, banktypes.Input{Address: saddr, Coins: input})
		}
		if !output.IsZero() {
			outputs = append(outputs, banktypes.Output{Address: saddr, Coins: input})
		}
	}
	return bankKeeper.InputOutputCoins(ctx, inputs, outputs)
}
