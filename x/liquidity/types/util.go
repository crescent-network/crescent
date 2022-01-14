package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type BulkSendCoinsOperation struct {
	Inputs  []banktypes.Input
	Outputs []banktypes.Output
}

func NewBulkSendCoinsOperation() *BulkSendCoinsOperation {
	return &BulkSendCoinsOperation{
		Inputs:  []banktypes.Input{},
		Outputs: []banktypes.Output{},
	}
}

func (op *BulkSendCoinsOperation) SendCoins(fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) {
	if amt.IsValid() && !amt.IsZero() {
		op.Inputs = append(op.Inputs, banktypes.NewInput(fromAddr, amt))
		op.Outputs = append(op.Outputs, banktypes.NewOutput(toAddr, amt))
	}
}

func (op *BulkSendCoinsOperation) Run(ctx sdk.Context, bankKeeper BankKeeper) error {
	if len(op.Inputs) > 0 && len(op.Outputs) > 0 {
		return bankKeeper.InputOutputCoins(ctx, op.Inputs, op.Outputs)
	}
	return nil
}
