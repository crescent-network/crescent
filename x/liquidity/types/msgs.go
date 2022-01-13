package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgCreatePool)(nil)
	_ sdk.Msg = (*MsgDepositBatch)(nil)
	_ sdk.Msg = (*MsgWithdrawBatch)(nil)
	_ sdk.Msg = (*MsgSwapBatch)(nil)
	_ sdk.Msg = (*MsgCancelSwapBatch)(nil)
)

// Message types for the liquidity module
const (
	TypeMsgCreatePool      = "create_pool"
	TypeMsgDepositBatch    = "deposit_batch"
	TypeMsgWithdrawBatch   = "withdraw_batch"
	TypeMsgSwapBatch       = "swap_batch"
	TypeMsgCancelSwapBatch = "cancel_swap_batch"
)

// NewMsgCreatePool creates a new MsgCreatePool.
func NewMsgCreatePool(
	creator sdk.AccAddress,
	xCoin sdk.Coin,
	yCoin sdk.Coin,
) *MsgCreatePool {
	return &MsgCreatePool{
		Creator: creator.String(),
		XCoin:   xCoin,
		YCoin:   yCoin,
	}
}

func (msg MsgCreatePool) Route() string { return RouterKey }

func (msg MsgCreatePool) Type() string { return TypeMsgCreatePool }

func (msg MsgCreatePool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address: %v", err)
	}
	if err := msg.XCoin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid deposit coin: %v", err)
	}
	if err := msg.YCoin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid deposit coin: %v", err)
	}
	if !msg.XCoin.IsPositive() || !msg.YCoin.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "deposit coins must be positive")
	}
	if msg.XCoin.Denom == msg.YCoin.Denom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "x coin denom and y coin denom must be different")
	}
	return nil
}

func (msg MsgCreatePool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreatePool) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreatePool) GetCreator() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgDepositBatch creates a new MsgDepositBatch.
func NewMsgDepositBatch(
	depositor sdk.AccAddress,
	poolId uint64,
	xCoin sdk.Coin,
	yCoin sdk.Coin,
) *MsgDepositBatch {
	return &MsgDepositBatch{
		Depositor: depositor.String(),
		PoolId:    poolId,
		XCoin:     xCoin,
		YCoin:     yCoin,
	}
}

func (msg MsgDepositBatch) Route() string { return RouterKey }

func (msg MsgDepositBatch) Type() string { return TypeMsgDepositBatch }

func (msg MsgDepositBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid depositor address: %v", err)
	}
	if err := msg.XCoin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid deposit coin: %v", err)
	}
	if err := msg.YCoin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid deposit coin: %v", err)
	}
	if !msg.XCoin.IsPositive() || !msg.YCoin.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "deposit coins must be positive")
	}
	return nil
}

func (msg MsgDepositBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgDepositBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgDepositBatch) GetDepositor() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgWithdrawBatch creates a new MsgWithdrawBatch.
func NewMsgWithdrawBatch(
	withdrawer sdk.AccAddress,
	poolId uint64,
	poolCoin sdk.Coin,
) *MsgWithdrawBatch {
	return &MsgWithdrawBatch{
		Withdrawer: withdrawer.String(),
		PoolId:     poolId,
		PoolCoin:   poolCoin,
	}
}

func (msg MsgWithdrawBatch) Route() string { return RouterKey }

func (msg MsgWithdrawBatch) Type() string { return TypeMsgWithdrawBatch }

func (msg MsgWithdrawBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Withdrawer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid withdrawer address: %v", err)
	}
	if err := msg.PoolCoin.Validate(); err != nil {
		return err
	}
	if !msg.PoolCoin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pool coin must be positive")
	}
	return nil
}

func (msg MsgWithdrawBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgWithdrawBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Withdrawer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgWithdrawBatch) GetWithdrawer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Withdrawer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgSwapBatch creates a new MsgSwapBatch.
func NewMsgSwapBatch(
	orderer sdk.AccAddress,
	xCoinDenom string,
	yCoinDenom string,
	offerCoin sdk.Coin,
	demandCoinDenom string,
	price sdk.Dec,
	orderLifespan time.Duration,
) *MsgSwapBatch {
	return &MsgSwapBatch{
		Orderer:         orderer.String(),
		XCoinDenom:      xCoinDenom,
		YCoinDenom:      yCoinDenom,
		OfferCoin:       offerCoin,
		DemandCoinDenom: demandCoinDenom,
		Price:           price,
		OrderLifespan:   orderLifespan,
	}
}

func (msg MsgSwapBatch) Route() string { return RouterKey }

func (msg MsgSwapBatch) Type() string { return TypeMsgSwapBatch }

func (msg MsgSwapBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Orderer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid orderer address: %v", err)
	}
	if err := sdk.ValidateDenom(msg.XCoinDenom); err != nil {
		return sdkerrors.Wrap(err, "invalid x coin denom")
	}
	if err := sdk.ValidateDenom(msg.YCoinDenom); err != nil {
		return sdkerrors.Wrap(err, "invalid y coin denom")
	}
	if msg.XCoinDenom == msg.YCoinDenom {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "x and y coin denoms must be different")
	}
	if err := msg.OfferCoin.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid offer coin")
	}
	if err := sdk.ValidateDenom(msg.DemandCoinDenom); err != nil {
		return sdkerrors.Wrap(err, "invalid demand coin denom")
	}
	if msg.GetDirection() == SwapDirectionUnspecified {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer and demand coin denom pair doesn't match with x and y coin denom pair")
	}
	if !msg.OfferCoin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin must be positive")
	}
	if !msg.OfferCoin.Amount.GTE(MinOfferCoinAmount) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin is less than minimum offer coin amount")
	}
	if !msg.Price.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "price must be positive")
	}
	return nil
}

func (msg MsgSwapBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSwapBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgSwapBatch) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgSwapBatch) GetDirection() SwapDirection {
	switch {
	case msg.OfferCoin.Denom == msg.XCoinDenom && msg.DemandCoinDenom == msg.YCoinDenom:
		return SwapDirectionBuy
	case msg.OfferCoin.Denom == msg.YCoinDenom && msg.DemandCoinDenom == msg.XCoinDenom:
		return SwapDirectionSell
	default:
		return SwapDirectionUnspecified
	}
}

// NewMsgCancelSwapBatch creates a new MsgCancelSwapBatch.
func NewMsgCancelSwapBatch(
	orderer sdk.AccAddress,
	pairId uint64,
	swapRequestId uint64,
) *MsgCancelSwapBatch {
	return &MsgCancelSwapBatch{
		SwapRequestId: swapRequestId,
		PairId:        pairId,
		Orderer:       orderer.String(),
	}
}

func (msg MsgCancelSwapBatch) Route() string { return RouterKey }

func (msg MsgCancelSwapBatch) Type() string { return TypeMsgCancelSwapBatch }

func (msg MsgCancelSwapBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Orderer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid orderer address: %v", err)
	}
	return nil
}

func (msg MsgCancelSwapBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCancelSwapBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCancelSwapBatch) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}
