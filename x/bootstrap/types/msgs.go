package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgLimitOrder)(nil)
	// TODO: add modify order
)

// Message types for the bootstrap module
const (
	TypeMsgLimitOrder = "limit_order"
)

// NewMsgLimitOrder creates a new limit order.
func NewMsgLimitOrder(
	orderer sdk.AccAddress,
	bootstrapPoolId uint64,
	direction OrderDirection,
	offerCoin sdk.Coin,
	price sdk.Dec,
) *MsgLimitOrder {
	return &MsgLimitOrder{
		Orderer:         orderer.String(),
		BootstrapPoolId: bootstrapPoolId,
		Direction:       direction,
		OfferCoin:       offerCoin,
		Price:           price,
	}
}

func (msg MsgLimitOrder) Route() string { return RouterKey }

func (msg MsgLimitOrder) Type() string { return TypeMsgLimitOrder }

func (msg MsgLimitOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Orderer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address %q: %v", msg.Orderer, err)
	}
	// TODO: bootstrap id, direction, offercoin, price
	//if len(msg.PairIds) == 0 {
	//	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pair ids must not be empty")
	//}
	//pairMap := make(map[uint64]struct{})
	//for _, pair := range msg.PairIds {
	//	if _, ok := pairMap[pair]; ok {
	//		return sdkerrors.Wrapf(ErrInvalidPairId, "duplicated pair id %d", pair)
	//	}
	//	pairMap[pair] = struct{}{}
	//}
	return nil
}

func (msg MsgLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgLimitOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgLimitOrder) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}
