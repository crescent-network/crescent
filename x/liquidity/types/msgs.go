package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgCreatePair)(nil)
	_ sdk.Msg = (*MsgCreatePool)(nil)
	_ sdk.Msg = (*MsgDepositBatch)(nil)
	_ sdk.Msg = (*MsgWithdrawBatch)(nil)
	_ sdk.Msg = (*MsgLimitOrderBatch)(nil)
	_ sdk.Msg = (*MsgMarketOrderBatch)(nil)
	_ sdk.Msg = (*MsgCancelOrderBatch)(nil)
)

// Message types for the liquidity module
const (
	TypeMsgCreatePair       = "create_pair"
	TypeMsgCreatePool       = "create_pool"
	TypeMsgDepositBatch     = "deposit_batch"
	TypeMsgWithdrawBatch    = "withdraw_batch"
	TypeMsgLimitOrderBatch  = "limit_order_batch"
	TypeMsgMarketOrderBatch = "market_order_batch"
	TypeMsgCancelOrderBatch = "cancel_order_batch"
)

// NewMsgCreatePair returns a new MsgCreatePair.
func NewMsgCreatePair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string) *MsgCreatePair {
	return &MsgCreatePair{
		Creator:        creator.String(),
		BaseCoinDenom:  baseCoinDenom,
		QuoteCoinDenom: quoteCoinDenom,
	}
}

func (msg MsgCreatePair) Route() string { return RouterKey }

func (msg MsgCreatePair) Type() string { return TypeMsgCreatePair }

func (msg MsgCreatePair) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address: %v", err)
	}
	if err := sdk.ValidateDenom(msg.BaseCoinDenom); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := sdk.ValidateDenom(msg.QuoteCoinDenom); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

func (msg MsgCreatePair) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreatePair) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreatePair) GetCreator() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgCreatePool creates a new MsgCreatePool.
func NewMsgCreatePool(
	creator sdk.AccAddress,
	pairId uint64,
	depositCoins sdk.Coins,
) *MsgCreatePool {
	return &MsgCreatePool{
		Creator:      creator.String(),
		PairId:       pairId,
		DepositCoins: depositCoins,
	}
}

func (msg MsgCreatePool) Route() string { return RouterKey }

func (msg MsgCreatePool) Type() string { return TypeMsgCreatePool }

func (msg MsgCreatePool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address: %v", err)
	}
	if msg.PairId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pair id must not be 0")
	}
	if err := msg.DepositCoins.Validate(); err != nil {
		return err
	}
	if len(msg.DepositCoins) != 2 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "wrong number of deposit coins: %d", len(msg.DepositCoins))
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
	depositCoins sdk.Coins,
) *MsgDepositBatch {
	return &MsgDepositBatch{
		Depositor:    depositor.String(),
		PoolId:       poolId,
		DepositCoins: depositCoins,
	}
}

func (msg MsgDepositBatch) Route() string { return RouterKey }

func (msg MsgDepositBatch) Type() string { return TypeMsgDepositBatch }

func (msg MsgDepositBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid depositor address: %v", err)
	}
	if msg.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pool id must not be 0")
	}
	if err := msg.DepositCoins.Validate(); err != nil {
		return err
	}
	if len(msg.DepositCoins) != 2 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "wrong number of deposit coins: %d", len(msg.DepositCoins))
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
	if msg.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pool id must not be 0")
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

// NewMsgLimitOrderBatch creates a new MsgLimitOrderBatch.
func NewMsgLimitOrderBatch(
	orderer sdk.AccAddress,
	pairId uint64,
	dir SwapDirection,
	offerCoin sdk.Coin,
	demandCoinDenom string,
	price sdk.Dec,
	amt sdk.Int,
	orderLifespan time.Duration,
) *MsgLimitOrderBatch {
	return &MsgLimitOrderBatch{
		Orderer:         orderer.String(),
		PairId:          pairId,
		Direction:       dir,
		OfferCoin:       offerCoin,
		DemandCoinDenom: demandCoinDenom,
		Price:           price,
		Amount:          amt,
		OrderLifespan:   orderLifespan,
	}
}

func (msg MsgLimitOrderBatch) Route() string { return RouterKey }

func (msg MsgLimitOrderBatch) Type() string { return TypeMsgLimitOrderBatch }

func (msg MsgLimitOrderBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Orderer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid orderer address: %v", err)
	}
	if msg.PairId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pair id must not be 0")
	}
	if msg.Direction != SwapDirectionBuy && msg.Direction != SwapDirectionSell {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unknown swap direction: %s", msg.Direction)
	}
	if !msg.Amount.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount must be positive: %s", msg.Amount)
	}
	if err := sdk.ValidateDenom(msg.DemandCoinDenom); err != nil {
		return sdkerrors.Wrap(err, "invalid demand coin denom")
	}
	if !msg.Price.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "price must be positive")
	}
	if err := msg.OfferCoin.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid offer coin")
	}
	if !msg.OfferCoin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin must be positive")
	}
	if msg.OfferCoin.Amount.LT(MinCoinAmount) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin is less than minimum coin amount")
	}
	if msg.Amount.LT(MinCoinAmount) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "base coin is less than minimum coin amount")
	}
	if msg.OfferCoin.Denom == msg.DemandCoinDenom {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin denom and demand coin denom must not be same")
	}
	return nil
}

func (msg MsgLimitOrderBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgLimitOrderBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgLimitOrderBatch) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgMarketOrderBatch creates a new MsgMarketOrderBatch.
func NewMsgMarketOrderBatch(
	orderer sdk.AccAddress,
	pairId uint64,
	dir SwapDirection,
	offerCoin sdk.Coin,
	demandCoinDenom string,
	amt sdk.Int,
	orderLifespan time.Duration,
) *MsgMarketOrderBatch {
	return &MsgMarketOrderBatch{
		Orderer:         orderer.String(),
		PairId:          pairId,
		Direction:       dir,
		OfferCoin:       offerCoin,
		DemandCoinDenom: demandCoinDenom,
		Amount:          amt,
		OrderLifespan:   orderLifespan,
	}
}

func (msg MsgMarketOrderBatch) Route() string { return RouterKey }

func (msg MsgMarketOrderBatch) Type() string { return TypeMsgMarketOrderBatch }

func (msg MsgMarketOrderBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Orderer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid orderer address: %v", err)
	}
	if msg.PairId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pair id must not be 0")
	}
	if msg.Direction != SwapDirectionBuy && msg.Direction != SwapDirectionSell {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unknown swap direction: %s", msg.Direction)
	}
	if !msg.Amount.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount must be positive: %s", msg.Amount)
	}
	if err := sdk.ValidateDenom(msg.DemandCoinDenom); err != nil {
		return sdkerrors.Wrap(err, "invalid demand coin denom")
	}
	if err := msg.OfferCoin.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid offer coin")
	}
	if !msg.OfferCoin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin must be positive")
	}
	if msg.OfferCoin.Amount.LT(MinCoinAmount) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin is less than minimum coin amount")
	}
	if msg.Amount.LT(MinCoinAmount) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "base coin is less than minimum coin amount")
	}
	if msg.OfferCoin.Denom == msg.DemandCoinDenom {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "offer coin denom and demand coin denom must not be same")
	}
	return nil
}

func (msg MsgMarketOrderBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgMarketOrderBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgMarketOrderBatch) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgCancelOrderBatch creates a new MsgCancelOrderBatch.
func NewMsgCancelOrderBatch(
	orderer sdk.AccAddress,
	pairId uint64,
	swapRequestId uint64,
) *MsgCancelOrderBatch {
	return &MsgCancelOrderBatch{
		SwapRequestId: swapRequestId,
		PairId:        pairId,
		Orderer:       orderer.String(),
	}
}

func (msg MsgCancelOrderBatch) Route() string { return RouterKey }

func (msg MsgCancelOrderBatch) Type() string { return TypeMsgCancelOrderBatch }

func (msg MsgCancelOrderBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Orderer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid orderer address: %v", err)
	}
	return nil
}

func (msg MsgCancelOrderBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCancelOrderBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCancelOrderBatch) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}
