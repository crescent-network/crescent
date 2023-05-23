package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgMintShare)(nil)
	_ sdk.Msg = (*MsgBurnShare)(nil)
	_ sdk.Msg = (*MsgPlaceBid)(nil)
	_ sdk.Msg = (*MsgRefundBid)(nil)
	_ sdk.Msg = (*MsgFinishAuctions)(nil)
)

// Message types for the module
const (
	TypeMsgMintShare      = "mint_share"
	TypeMsgBurnShare      = "burn_share"
	TypeMsgPlaceBid       = "place_bid"
	TypeMsgRefundBid      = "refund_bid"
	TypeMsgFinishAuctions = "finish_auctions"
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

// NewMsgRefundBid creates a new MsgRefundBid
func NewMsgRefundBid(senderAddr sdk.AccAddress, liquidFarmId, rewardsAuctionId uint64) *MsgRefundBid {
	return &MsgRefundBid{
		Sender:           senderAddr.String(),
		LiquidFarmId:     liquidFarmId,
		RewardsAuctionId: rewardsAuctionId,
	}
}

func (msg MsgRefundBid) Route() string { return RouterKey }
func (msg MsgRefundBid) Type() string  { return TypeMsgRefundBid }

func (msg MsgRefundBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRefundBid) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgRefundBid) ValidateBasic() error {
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

// NewMsgFinishAuctions creates a new MsgFinishAuctions.
func NewMsgFinishAuctions(senderAddr sdk.AccAddress) *MsgFinishAuctions {
	return &MsgFinishAuctions{
		Sender: senderAddr.String(),
	}
}

func (msg MsgFinishAuctions) Route() string { return RouterKey }
func (msg MsgFinishAuctions) Type() string  { return TypeMsgFinishAuctions }

func (msg MsgFinishAuctions) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgFinishAuctions) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgFinishAuctions) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}
