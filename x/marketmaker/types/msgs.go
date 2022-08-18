package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgApplyMarketMaker)(nil)
	_ sdk.Msg = (*MsgClaimIncentives)(nil)
)

// Message types for the marketmaker module
const (
	TypeMsgApplyMarketMaker = "apply_market_maker"
	TypeMsgClaimIncentives  = "claim_incentives"
)

// NewMsgApplyMarketMaker creates a new MsgApplyMarketMaker.
func NewMsgApplyMarketMaker(
	marketMaker sdk.AccAddress,
	pairIds []uint64,
) *MsgApplyMarketMaker {
	return &MsgApplyMarketMaker{
		Address: marketMaker.String(),
		PairIds: pairIds,
	}
}

func (msg MsgApplyMarketMaker) Route() string { return RouterKey }

func (msg MsgApplyMarketMaker) Type() string { return TypeMsgApplyMarketMaker }

func (msg MsgApplyMarketMaker) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address %q: %v", msg.Address, err)
	}
	if len(msg.PairIds) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pair ids must not be empty")
	}
	pairMap := make(map[uint64]struct{})
	for _, pair := range msg.PairIds {
		if _, ok := pairMap[pair]; ok {
			return sdkerrors.Wrapf(ErrInvalidPairId, "duplicated pair id %d", pair)
		}
		pairMap[pair] = struct{}{}
	}
	return nil
}

func (msg MsgApplyMarketMaker) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgApplyMarketMaker) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgApplyMarketMaker) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgClaimIncentives creates a new MsgClaimIncentives.
func NewMsgClaimIncentives(
	marketMaker sdk.AccAddress,
) *MsgClaimIncentives {
	return &MsgClaimIncentives{
		Address: marketMaker.String(),
	}
}

func (msg MsgClaimIncentives) Route() string { return RouterKey }

func (msg MsgClaimIncentives) Type() string { return TypeMsgClaimIncentives }

func (msg MsgClaimIncentives) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address %q: %v", msg.Address, err)
	}
	return nil
}

func (msg MsgClaimIncentives) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgClaimIncentives) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgClaimIncentives) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}
