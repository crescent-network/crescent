package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgMintShare)(nil)
	_ sdk.Msg = (*MsgBurnShare)(nil)
	_ sdk.Msg = (*MsgPlaceBid)(nil)
	_ sdk.Msg = (*MsgCancelBid)(nil)
)

// Message types for the module
const (
	TypeMsgMintShare = "mint_share"
	TypeMsgBurnShare = "burn_share"
	TypeMsgPlaceBid  = "place_bid"
	TypeMsgCancelBid = "cancel_bid"
)

// NewMsgMintShare creates a new MsgMintShare
func NewMsgMintShare(senderAddr sdk.AccAddress, liquidFarmId uint64, desiredAmt sdk.Coins) *MsgMintShare {
	return &MsgMintShare{
		Sender:        senderAddr.String(),
		LiquidFarmId:  liquidFarmId,
		DesiredAmount: desiredAmt,
	}
}

func (msg MsgMintShare) Route() string { return RouterKey }
func (msg MsgMintShare) Type() string  { return TypeMsgMintShare }

func (msg MsgMintShare) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgMintShare) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgMintShare) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.LiquidFarmId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "liquid farm id must not be 0")
	}
	if err := msg.DesiredAmount.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid desired amount: %v", err)
	}
	return nil
}

// NewMsgBurnShare creates a new MsgBurnShare
func NewMsgBurnShare(senderAddr sdk.AccAddress, liquidFarmId uint64, share sdk.Coin) *MsgBurnShare {
	return &MsgBurnShare{
		Sender:       senderAddr.String(),
		LiquidFarmId: liquidFarmId,
		Share:        share,
	}
}

func (msg MsgBurnShare) Route() string { return RouterKey }
func (msg MsgBurnShare) Type() string  { return TypeMsgBurnShare }

func (msg MsgBurnShare) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgBurnShare) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgBurnShare) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.LiquidFarmId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "liquid farm id must not be 0")
	}
	if err := msg.Share.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid share: %v", err)
	}
	if !msg.Share.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "share amount must be positive: %s", msg.Share)
	}
	return nil
}

// NewMsgPlaceBid creates a new MsgPlaceBid
func NewMsgPlaceBid(senderAddr sdk.AccAddress, liquidFarmId, auctionId uint64, share sdk.Coin) *MsgPlaceBid {
	return &MsgPlaceBid{
		Sender:           senderAddr.String(),
		LiquidFarmId:     liquidFarmId,
		RewardsAuctionId: auctionId,
		Share:            share,
	}
}

func (msg MsgPlaceBid) Route() string { return RouterKey }
func (msg MsgPlaceBid) Type() string  { return TypeMsgPlaceBid }

func (msg MsgPlaceBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceBid) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceBid) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.LiquidFarmId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "liquid farm id must not be 0")
	}
	if msg.RewardsAuctionId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "rewards auction id must not be 0")
	}
	if err := msg.Share.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid share: %v", err)
	}
	if !msg.Share.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "share amount must be positive: %s", msg.Share)
	}
	return nil
}

// NewMsgCancelBid creates a new MsgCancelBid
func NewMsgCancelBid(senderAddr sdk.AccAddress, liquidFarmId, rewardsAuctionId uint64) *MsgCancelBid {
	return &MsgCancelBid{
		Sender:           senderAddr.String(),
		LiquidFarmId:     liquidFarmId,
		RewardsAuctionId: rewardsAuctionId,
	}
}

func (msg MsgCancelBid) Route() string { return RouterKey }
func (msg MsgCancelBid) Type() string  { return TypeMsgCancelBid }

func (msg MsgCancelBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCancelBid) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCancelBid) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.LiquidFarmId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "liquid farm id must not be 0")
	}
	if msg.RewardsAuctionId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "rewards auction id must not be 0")
	}
	return nil
}
