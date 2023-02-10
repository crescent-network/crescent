package antehandlers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type ProposalExtended interface {
	govtypes.Content

	GetProposerAddress() string
}

type ValidateProposerDecorator struct{}

func NewValidateProposerDecorator() ValidateProposerDecorator {
	return ValidateProposerDecorator{}
}

func (decorator ValidateProposerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		switch msg := msg.(type) {
		case *govtypes.MsgSubmitProposal:
			content := msg.GetContent()
			switch c := content.(type) {
			case ProposalExtended:
				if msg.Proposer != c.GetProposerAddress() {
					return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid proposer address")
				}
			default:
				return next(ctx, tx, simulate)
			}
		default:
			return next(ctx, tx, simulate)
		}
	}
	return next(ctx, tx, simulate)
}
