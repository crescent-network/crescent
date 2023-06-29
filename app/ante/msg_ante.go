package ante

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	claimtypes "github.com/crescent-network/crescent/v5/x/claim/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
	farmingtypes "github.com/crescent-network/crescent/v5/x/farming/types"
	liquiditytypes "github.com/crescent-network/crescent/v5/x/liquidity/types"
	lpfarmtypes "github.com/crescent-network/crescent/v5/x/lpfarm/types"
)

// initial deposit must be greater than or equal to 50% of the minimum deposit
var minInitialDepositFraction = sdk.NewDecWithPrec(50, 2)

type MsgFilterDecorator struct {
	govKeeper *govkeeper.Keeper
	cdc       codec.BinaryCodec
	enabled   bool
}

func NewMsgFilterDecorator(cdc codec.BinaryCodec, govKeeper *govkeeper.Keeper, enabled bool) MsgFilterDecorator {
	return MsgFilterDecorator{
		govKeeper: govKeeper,
		cdc:       cdc,
		enabled:   enabled,
	}
}

func (d MsgFilterDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if !d.enabled {
		return next(ctx, tx, simulate)
	}
	if err = d.ValidateMsgs(ctx, tx.GetMsgs()); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (d MsgFilterDecorator) ValidateMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	numMsg, numBatchMsg := 0, 0
	var minInitialDeposit sdk.Coins
	validateMsg := func(msg sdk.Msg, nested bool) error {
		numMsg++
		switch msg := msg.(type) {
		// prevent gov messages with insufficient initial deposit amount
		case *govtypes.MsgSubmitProposal:
			// mempool(check tx) level msg filter
			if ctx.IsCheckTx() {
				if minInitialDeposit.Empty() {
					depositParams := d.govKeeper.GetDepositParams(ctx)
					minInitialDeposit = CalcMinInitialDeposit(depositParams.MinDeposit, minInitialDepositFraction)
				}

				if !msg.InitialDeposit.IsAllGTE(minInitialDeposit) {
					return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "insufficient initial deposit amount - required: %v", minInitialDeposit)
				}
			}

		// tracking mixed batch msg with regular msg
		case *exchangetypes.MsgPlaceBatchLimitOrder,
			*exchangetypes.MsgPlaceMMBatchLimitOrder,
			*exchangetypes.MsgCancelOrder:
			numMsg--
			numBatchMsg++

		// block double nested MsgExec
		case *authz.MsgExec:
			if nested {
				return fmt.Errorf("double nested %s is not allowed", sdk.MsgTypeURL(msg))
			}
		default:
			// block deprecated module's msgs
			if legacyMsg, ok := msg.(legacytx.LegacyMsg); ok {
				switch legacyMsg.Route() {
				case liquiditytypes.RouterKey,
					farmingtypes.RouterKey,
					lpfarmtypes.RouterKey,
					claimtypes.RouterKey:
					return fmt.Errorf("%s is deprecated msg type", sdk.MsgTypeURL(msg))
				}
			}
		}

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
		// validate authz nested msgs
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

	// block mixed batch msg with regular msg
	if numBatchMsg > 0 && numMsg > 0 {
		return fmt.Errorf("cannot mix batch msg and regular msg in one tx")
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
