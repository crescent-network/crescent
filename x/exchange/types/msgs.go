package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgCreateMarket)(nil)
	_ sdk.Msg = (*MsgPlaceLimitOrder)(nil)
	_ sdk.Msg = (*MsgPlaceBatchLimitOrder)(nil)
	_ sdk.Msg = (*MsgPlaceMMLimitOrder)(nil)
	_ sdk.Msg = (*MsgPlaceMMBatchLimitOrder)(nil)
	_ sdk.Msg = (*MsgPlaceMarketOrder)(nil)
	_ sdk.Msg = (*MsgCancelOrder)(nil)
	_ sdk.Msg = (*MsgCancelAllOrders)(nil)
	_ sdk.Msg = (*MsgSwapExactAmountIn)(nil)
)

// Message types for the module
const (
	TypeMsgCreateMarket           = "create_market"
	TypeMsgPlaceLimitOrder        = "place_limit_order"
	TypeMsgPlaceBatchLimitOrder   = "place_batch_limit_order"
	TypeMsgPlaceMMLimitOrder      = "place_mm_limit_order"
	TypeMsgPlaceMMBatchLimitOrder = "place_mm_batch_limit_order"
	TypeMsgPlaceMarketOrder       = "place_market_order"
	TypeMsgCancelOrder            = "cancel_order"
	TypeMsgCancelAllOrders        = "cancel_all_orders"
	TypeMsgSwapExactAmountIn      = "swap_exact_amount_in"
)

func NewMsgCreateMarket(
	senderAddr sdk.AccAddress, baseDenom, quoteDenom string) *MsgCreateMarket {
	return &MsgCreateMarket{
		Sender:     senderAddr.String(),
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
	}
}

func (msg MsgCreateMarket) Route() string { return RouterKey }
func (msg MsgCreateMarket) Type() string  { return TypeMsgCreateMarket }

func (msg MsgCreateMarket) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateMarket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreateMarket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if err := sdk.ValidateDenom(msg.BaseDenom); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid base denom: %v", err)
	}
	if err := sdk.ValidateDenom(msg.QuoteDenom); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid quote denom: %v", err)
	}
	if msg.BaseDenom == msg.QuoteDenom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "base denom and quote denom must not be same: %s", msg.BaseDenom)
	}
	return nil
}

func NewMsgPlaceLimitOrder(
	senderAddr sdk.AccAddress, marketId uint64, isBuy bool,
	price, qty sdk.Dec, lifespan time.Duration) *MsgPlaceLimitOrder {
	return &MsgPlaceLimitOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Price:    price,
		Quantity: qty,
		Lifespan: lifespan,
	}
}

func (msg MsgPlaceLimitOrder) Route() string { return RouterKey }
func (msg MsgPlaceLimitOrder) Type() string  { return TypeMsgPlaceLimitOrder }

func (msg MsgPlaceLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceLimitOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceLimitOrder) ValidateBasic() error {
	return ValidateLimitOrderMsg(msg.Sender, msg.MarketId, msg.IsBuy, msg.Price, msg.Quantity, msg.Lifespan)
}

func NewMsgPlaceBatchLimitOrder(
	senderAddr sdk.AccAddress, marketId uint64, isBuy bool,
	price, qty sdk.Dec, lifespan time.Duration) *MsgPlaceBatchLimitOrder {
	return &MsgPlaceBatchLimitOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Price:    price,
		Quantity: qty,
		Lifespan: lifespan,
	}
}

func (msg MsgPlaceBatchLimitOrder) Route() string { return RouterKey }
func (msg MsgPlaceBatchLimitOrder) Type() string  { return TypeMsgPlaceBatchLimitOrder }

func (msg MsgPlaceBatchLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceBatchLimitOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceBatchLimitOrder) ValidateBasic() error {
	return ValidateLimitOrderMsg(msg.Sender, msg.MarketId, msg.IsBuy, msg.Price, msg.Quantity, msg.Lifespan)
}

func NewMsgPlaceMMLimitOrder(
	senderAddr sdk.AccAddress, marketId uint64, isBuy bool,
	price, qty sdk.Dec, lifespan time.Duration) *MsgPlaceMMLimitOrder {
	return &MsgPlaceMMLimitOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Price:    price,
		Quantity: qty,
		Lifespan: lifespan,
	}
}

func (msg MsgPlaceMMLimitOrder) Route() string { return RouterKey }
func (msg MsgPlaceMMLimitOrder) Type() string  { return TypeMsgPlaceMMLimitOrder }

func (msg MsgPlaceMMLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceMMLimitOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceMMLimitOrder) ValidateBasic() error {
	return ValidateLimitOrderMsg(msg.Sender, msg.MarketId, msg.IsBuy, msg.Price, msg.Quantity, msg.Lifespan)
}

func NewMsgPlaceMMBatchLimitOrder(
	senderAddr sdk.AccAddress, marketId uint64, isBuy bool,
	price, qty sdk.Dec, lifespan time.Duration) *MsgPlaceMMBatchLimitOrder {
	return &MsgPlaceMMBatchLimitOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Price:    price,
		Quantity: qty,
		Lifespan: lifespan,
	}
}

func (msg MsgPlaceMMBatchLimitOrder) Route() string { return RouterKey }
func (msg MsgPlaceMMBatchLimitOrder) Type() string  { return TypeMsgPlaceMMBatchLimitOrder }

func (msg MsgPlaceMMBatchLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceMMBatchLimitOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceMMBatchLimitOrder) ValidateBasic() error {
	return ValidateLimitOrderMsg(msg.Sender, msg.MarketId, msg.IsBuy, msg.Price, msg.Quantity, msg.Lifespan)
}

func NewMsgPlaceMarketOrder(
	senderAddr sdk.AccAddress, marketId uint64,
	isBuy bool, qty sdk.Dec) *MsgPlaceMarketOrder {
	return &MsgPlaceMarketOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Quantity: qty,
	}
}

func (msg MsgPlaceMarketOrder) Route() string { return RouterKey }
func (msg MsgPlaceMarketOrder) Type() string  { return TypeMsgPlaceMarketOrder }

func (msg MsgPlaceMarketOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceMarketOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceMarketOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.MarketId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market id must not be 0")
	}
	if !msg.Quantity.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quantity must be positive: %s", msg.Quantity)
	}
	if !msg.Quantity.TruncateDec().Equal(msg.Quantity) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quantity must be an integer: %s", msg.Quantity)
	}
	return nil
}

func NewMsgCancelOrder(senderAddr sdk.AccAddress, orderId uint64) *MsgCancelOrder {
	return &MsgCancelOrder{
		Sender:  senderAddr.String(),
		OrderId: orderId,
	}
}

func (msg MsgCancelOrder) Route() string { return RouterKey }
func (msg MsgCancelOrder) Type() string  { return TypeMsgCancelOrder }

func (msg MsgCancelOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCancelOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCancelOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.OrderId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "order id must not be 0")
	}
	return nil
}

func NewMsgCancelAllOrders(senderAddr sdk.AccAddress, marketId uint64) *MsgCancelAllOrders {
	return &MsgCancelAllOrders{
		Sender:   senderAddr.String(),
		MarketId: marketId,
	}
}

func (msg MsgCancelAllOrders) Route() string { return RouterKey }
func (msg MsgCancelAllOrders) Type() string  { return TypeMsgCancelAllOrders }

func (msg MsgCancelAllOrders) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCancelAllOrders) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCancelAllOrders) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.MarketId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market id must not be 0")
	}
	return nil
}

func NewMsgSwapExactAmountIn(
	senderAddr sdk.AccAddress, routes []uint64, input, minOutput sdk.DecCoin) *MsgSwapExactAmountIn {
	return &MsgSwapExactAmountIn{
		Sender:    senderAddr.String(),
		Routes:    routes,
		Input:     input,
		MinOutput: minOutput,
	}
}

func (msg MsgSwapExactAmountIn) Route() string { return RouterKey }
func (msg MsgSwapExactAmountIn) Type() string  { return TypeMsgSwapExactAmountIn }

func (msg MsgSwapExactAmountIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSwapExactAmountIn) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgSwapExactAmountIn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if len(msg.Routes) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "routes must not be empty")
	}
	for _, marketId := range msg.Routes {
		if marketId == 0 {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market id must not be 0")
		}
	}
	if err := msg.Input.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid input: %v", err)
	}
	if !msg.Input.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "input must be positive: %s", msg.Input)
	}
	if !msg.Input.Amount.TruncateDec().Equal(msg.Input.Amount) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "input amount must be integer: %s", msg.Input)
	}
	if err := msg.MinOutput.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid min output: %v", err)
	}
	if !msg.MinOutput.Amount.TruncateDec().Equal(msg.MinOutput.Amount) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "min output amount must be integer: %s", msg.MinOutput)
	}
	return nil
}

func ValidateLimitOrderMsg(
	sender string, marketId uint64, isBuy bool, price, qty sdk.Dec, lifespan time.Duration) error {
	if _, err := sdk.AccAddressFromBech32(sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if marketId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market id must not be 0")
	}
	if price.LT(MinPrice) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price is lower than the min price; %s < %s", price, MinPrice)
	}
	if price.GT(MaxPrice) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price is higher than the max price; %s > %s", price, MaxPrice)
	}
	if _, valid := ValidateTickPrice(price); !valid {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid price tick: %s", price)
	}
	if !qty.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quantity must be positive: %s", qty)
	}
	if !qty.TruncateDec().Equal(qty) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quantity must be an integer: %s", qty)
	}
	if lifespan < 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "lifespan must not be negative: %v", lifespan)
	}
	return nil
}
