package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var (
	_ sdk.Msg = (*MsgCreatePool)(nil)
	_ sdk.Msg = (*MsgAddLiquidity)(nil)
	_ sdk.Msg = (*MsgRemoveLiquidity)(nil)
)

// Message types for the module
const (
	TypeMsgCreatePool      = "create_pool"
	TypeMsgAddLiquidity    = "add_liquidity"
	TypeMsgRemoveLiquidity = "remove_liquidity"
)

func NewMsgCreatePool(
	senderAddr sdk.AccAddress, marketId uint64, tickSpacing uint32, price sdk.Dec) *MsgCreatePool {
	return &MsgCreatePool{
		Sender:      senderAddr.String(),
		MarketId:    marketId,
		TickSpacing: tickSpacing,
		Price:       price,
	}
}

func (msg MsgCreatePool) Route() string { return RouterKey }
func (msg MsgCreatePool) Type() string  { return TypeMsgCreatePool }

func (msg MsgCreatePool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreatePool) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreatePool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.MarketId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market id must not be 0")
	}
	if msg.TickSpacing == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "tick spacing must be positive")
	}
	if !msg.Price.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "price must be positive")
	}
	if msg.Price.LT(exchangetypes.MinPrice) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price is lower than the min price %s", exchangetypes.MinPrice)
	}
	if msg.Price.LT(exchangetypes.MaxPrice) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price is higher than the max price %s", exchangetypes.MaxPrice)
	}
	return nil
}

func NewMsgAddLiquidity(
	senderAddr sdk.AccAddress, poolId uint64, lowerPrice, upperPrice sdk.Dec,
	desiredAmt0, desiredAmt1, minAmt0, minAmt1 sdk.Int) *MsgAddLiquidity {
	return &MsgAddLiquidity{
		Sender:         senderAddr.String(),
		PoolId:         poolId,
		LowerPrice:     lowerPrice,
		UpperPrice:     upperPrice,
		DesiredAmount0: desiredAmt0,
		DesiredAmount1: desiredAmt1,
		MinAmount0:     minAmt0,
		MinAmount1:     minAmt1,
	}
}

func (msg MsgAddLiquidity) Route() string { return RouterKey }
func (msg MsgAddLiquidity) Type() string  { return TypeMsgAddLiquidity }

func (msg MsgAddLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgAddLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgRemoveLiquidity(
	senderAddr sdk.AccAddress, positionId uint64, liquidity sdk.Dec,
	minAmt0, minAmt1 sdk.Int) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		Sender:     senderAddr.String(),
		PositionId: positionId,
		Liquidity:  liquidity,
		MinAmount0: minAmt0,
		MinAmount1: minAmt1,
	}
}

func (msg MsgRemoveLiquidity) Route() string { return RouterKey }
func (msg MsgRemoveLiquidity) Type() string  { return TypeMsgRemoveLiquidity }

func (msg MsgRemoveLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRemoveLiquidity) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgRemoveLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}
