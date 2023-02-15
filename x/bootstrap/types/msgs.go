package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v4/x/liquidity/amm"
)

var (
	_ sdk.Msg = (*MsgLimitOrder)(nil)
	_ sdk.Msg = (*MsgModifyOrder)(nil)
)

// Message types for the bootstrap module
const (
	TypeMsgLimitOrder  = "limit_order"
	TypeMsgModifyOrder = "modify_order"
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
	if msg.BootstrapPoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pool id must not be 0")
	}
	if msg.Direction != OrderDirectionBuy && msg.Direction != OrderDirectionSell {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid order direction: %s", msg.Direction)
	}
	// TODO: DemandCoinDenom
	//if err := sdk.ValidateDenom(msg.OfferCoin.Denom); err != nil {
	//	return sdkerrors.Wrap(err, "invalid demand coin denom")
	//}

	if !msg.Price.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "price must be positive")
	}
	if err := msg.OfferCoin.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid offer coin")
	}
	if msg.OfferCoin.Amount.LT(amm.MinCoinAmount) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "offer coin %s is smaller than the min amount %s", msg.OfferCoin, amm.MinCoinAmount)
	}
	if msg.OfferCoin.Amount.GT(amm.MaxCoinAmount) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "offer coin %s is bigger than the max amount %s", msg.OfferCoin, amm.MaxCoinAmount)
	}
	//if msg.Amount.LT(amm.MinCoinAmount) {
	//	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "order amount %s is smaller than the min amount %s", msg.Amount, amm.MinCoinAmount)
	//}
	//if msg.Amount.GT(amm.MaxCoinAmount) {
	//	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "order amount %s is bigger than the max amount %s", msg.Amount, amm.MaxCoinAmount)
	//}
	//var minOfferCoin sdk.Coin
	//switch msg.Direction {
	//case OrderDirectionBuy:
	//	minOfferCoin = sdk.NewCoin(msg.OfferCoin.Denom, amm.OfferCoinAmount(amm.Buy, msg.Price, msg.Amount))
	//case OrderDirectionSell:
	//	minOfferCoin = sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount)
	//}
	//if msg.OfferCoin.IsLT(minOfferCoin) {
	//	return sdkerrors.Wrapf(ErrInsufficientOfferCoin, "%s is less than %s", msg.OfferCoin, minOfferCoin)
	//}
	//if msg.OfferCoin.Denom == msg.DemandCoinDenom {
	//	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin denom and demand coin denom must not be same")
	//}
	return nil
}

func (msg MsgLimitOrder) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
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

// NewMsgModifyOrder modifies existing limit order.
func NewMsgModifyOrder(
	orderer sdk.AccAddress,
	bootstrapPoolId uint64,
	orderId uint64,
	offerCoin sdk.Coin,
	price sdk.Dec,
) *MsgModifyOrder {
	return &MsgModifyOrder{
		Orderer:         orderer.String(),
		BootstrapPoolId: bootstrapPoolId,
		OrderId:         orderId,
		OfferCoin:       offerCoin,
		Price:           price,
	}
}

func (msg MsgModifyOrder) Route() string { return RouterKey }

func (msg MsgModifyOrder) Type() string { return TypeMsgModifyOrder }

func (msg MsgModifyOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Orderer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address %q: %v", msg.Orderer, err)
	}
	// TODO: bootstrap id, direction, offercoin, price
	if msg.BootstrapPoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pool id must not be 0")
	}
	// TODO: DemandCoinDenom
	//if err := sdk.ValidateDenom(msg.OfferCoin.Denom); err != nil {
	//	return sdkerrors.Wrap(err, "invalid demand coin denom")
	//}

	if !msg.Price.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "price must be positive")
	}
	if err := msg.OfferCoin.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid offer coin")
	}
	if msg.OfferCoin.Amount.LT(amm.MinCoinAmount) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "offer coin %s is smaller than the min amount %s", msg.OfferCoin, amm.MinCoinAmount)
	}
	if msg.OfferCoin.Amount.GT(amm.MaxCoinAmount) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "offer coin %s is bigger than the max amount %s", msg.OfferCoin, amm.MaxCoinAmount)
	}
	//if msg.Amount.LT(amm.MinCoinAmount) {
	//	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "order amount %s is smaller than the min amount %s", msg.Amount, amm.MinCoinAmount)
	//}
	//if msg.Amount.GT(amm.MaxCoinAmount) {
	//	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "order amount %s is bigger than the max amount %s", msg.Amount, amm.MaxCoinAmount)
	//}
	//var minOfferCoin sdk.Coin
	//switch msg.Direction {
	//case OrderDirectionBuy:
	//	minOfferCoin = sdk.NewCoin(msg.OfferCoin.Denom, amm.OfferCoinAmount(amm.Buy, msg.Price, msg.Amount))
	//case OrderDirectionSell:
	//	minOfferCoin = sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount)
	//}
	//if msg.OfferCoin.IsLT(minOfferCoin) {
	//	return sdkerrors.Wrapf(ErrInsufficientOfferCoin, "%s is less than %s", msg.OfferCoin, minOfferCoin)
	//}
	//if msg.OfferCoin.Denom == msg.DemandCoinDenom {
	//	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin denom and demand coin denom must not be same")
	//}
	return nil
}

func (msg MsgModifyOrder) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgModifyOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgModifyOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgModifyOrder) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}
