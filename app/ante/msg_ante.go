package ante

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	claimtypes "github.com/crescent-network/crescent/v5/x/claim/types"
	farmingtypes "github.com/crescent-network/crescent/v5/x/farming/types"
)

// initial deposit must be greater than or equal to 50% of the minimum deposit
var minInitialDepositFraction = sdk.NewDecWithPrec(50, 2)

type MsgFilterDecorator struct {
	govKeeper *govkeeper.Keeper
	cdc       codec.BinaryCodec
}

func NewMsgFilterDecorator(cdc codec.BinaryCodec, govKeeper *govkeeper.Keeper) MsgFilterDecorator {
	return MsgFilterDecorator{
		govKeeper: govKeeper,
		cdc:       cdc,
	}
}

func (d MsgFilterDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	if err = d.ValidateMsgs(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (d MsgFilterDecorator) ValidateMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	var minInitialDeposit sdk.Coins
	validateMsg := func(msg sdk.Msg, nested bool) error {
		// mempool(check tx) level msg filter
		if ctx.IsCheckTx() {
			switch msg := msg.(type) {
			// prevent messages with insufficient initial deposit amount
			case *govtypes.MsgSubmitProposal:
				if minInitialDeposit.Empty() {
					depositParams := d.govKeeper.GetDepositParams(ctx)
					minInitialDeposit = CalcMinInitialDeposit(depositParams.MinDeposit, minInitialDepositFraction)
				}

				if !msg.InitialDeposit.IsAllGTE(minInitialDeposit) {
					return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "insufficient initial deposit amount - required: %v", minInitialDeposit)
				}
			}
		}

		// deliver tx level msg filter
		switch msg := msg.(type) {
		// deprecated msgs
		case *claimtypes.MsgClaim,
			*farmingtypes.MsgCreateFixedAmountPlan,
			*farmingtypes.MsgCreateRatioPlan,
			*farmingtypes.MsgStake,
			*farmingtypes.MsgUnstake,
			*farmingtypes.MsgHarvest,
			*farmingtypes.MsgRemovePlan,
			*farmingtypes.MsgAdvanceEpoch:
			return fmt.Errorf("%s is deprecated msg type", sdk.MsgTypeURL(msg))
		// block double nested MsgExec
		case *authz.MsgExec:
			if nested {
				return fmt.Errorf("double nested %s is not allowed", sdk.MsgTypeURL(msg))
			}
		}

		// TODO: on next PR
		// - add other deprecated msg types
		// - prevent authz nested midblock, batch msgs
		// - prevent multi msgs midblock, batch msgs with normal msg

		return nil
	}

	validateAuthz := func(execMsg *authz.MsgExec) error {
		for _, v := range execMsg.Msgs {
			var innerMsg sdk.Msg
			if err := d.cdc.UnpackAny(v, &innerMsg); err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
			}

			if err := validateMsg(innerMsg, true); err != nil {
				return err
			}
		}
		return nil
	}

	for _, m := range msgs {
		if authzMsg, ok := m.(*authz.MsgExec); ok {
			if err := validateAuthz(authzMsg); err != nil {
				return err
			}
			continue
		}

		// validate normal msgs
		if err := validateMsg(m, false); err != nil {
			return err
		}
	}
	return nil
}

func CalcMinInitialDeposit(minDeposit sdk.Coins, minInitialDepositFraction sdk.Dec) (minInitialDeposit sdk.Coins) {
	for _, coin := range minDeposit {
		minInitialCoins := minInitialDepositFraction.MulInt(coin.Amount).RoundInt()
		minInitialDeposit = minInitialDeposit.Add(sdk.NewCoin(coin.Denom, minInitialCoins))
	}
	return
}
