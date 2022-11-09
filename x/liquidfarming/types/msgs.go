package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

var (
	_ sdk.Msg = (*MsgLiquidFarm)(nil)
	_ sdk.Msg = (*MsgLiquidUnfarm)(nil)
	_ sdk.Msg = (*MsgLiquidUnfarmAndWithdraw)(nil)
	_ sdk.Msg = (*MsgPlaceBid)(nil)
	_ sdk.Msg = (*MsgRefundBid)(nil)
	_ sdk.Msg = (*MsgAdvanceAuction)(nil)
)

// Message types for the module
const (
	TypeMsgLiquidFarm              = "liquid_farm"
	TypeMsgLiquidUnfarm            = "liquid_unfarm"
	TypeMsgLiquidUnfarmAndWithdraw = "liquid_unfarm_and_withdraw"
	TypeMsgPlaceBid                = "place_bid"
	TypeMsgRefundBid               = "refund_bid"
	TypeMsgAdvanceAuction          = "advance_auction"
)

// NewMsgLiquidFarm creates a new MsgLiquidFarm
func NewMsgLiquidFarm(poolId uint64, farmer string, farmingCoin sdk.Coin) *MsgLiquidFarm {
	return &MsgLiquidFarm{
		PoolId:      poolId,
		Farmer:      farmer,
		FarmingCoin: farmingCoin,
	}
}

func (msg MsgLiquidFarm) Route() string { return RouterKey }

func (msg MsgLiquidFarm) Type() string { return TypeMsgLiquidFarm }

func (msg MsgLiquidFarm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address: %v", err)
	}
	if msg.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid pool id")
	}
	if err := msg.FarmingCoin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid farming coin: %v", err)
	}
	if !msg.FarmingCoin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "farming coin must be positive")
	}
	poolCoinDenom := liquiditytypes.PoolCoinDenom(msg.PoolId)
	if poolCoinDenom != msg.FarmingCoin.Denom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected denom %s, but got %s", poolCoinDenom, msg.FarmingCoin.Denom)
	}
	return nil
}

func (msg MsgLiquidFarm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgLiquidFarm) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgLiquidFarm) GetFarmer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgLiquidUnfarm creates a new MsgLiquidUnfarm
func NewMsgLiquidUnfarm(poolId uint64, farmer string, unfarmingCoin sdk.Coin) *MsgLiquidUnfarm {
	return &MsgLiquidUnfarm{
		PoolId:        poolId,
		Farmer:        farmer,
		UnfarmingCoin: unfarmingCoin,
	}
}

func (msg MsgLiquidUnfarm) Route() string { return RouterKey }

func (msg MsgLiquidUnfarm) Type() string { return TypeMsgLiquidUnfarm }

func (msg MsgLiquidUnfarm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address: %v", err)
	}
	if msg.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid pool id")
	}
	if err := msg.UnfarmingCoin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid unfarming coin: %v", err)
	}
	if !msg.UnfarmingCoin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "unfarming coin must be positive")
	}
	expCoinDenom := LiquidFarmCoinDenom(msg.PoolId)
	if msg.UnfarmingCoin.Denom != expCoinDenom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected denom %s, but got %s", expCoinDenom, msg.UnfarmingCoin.Denom)
	}
	return nil
}

func (msg MsgLiquidUnfarm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgLiquidUnfarm) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgLiquidUnfarm) GetFarmer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgLiquidUnfarmAndWithdraw creates a new MsgLiquidUnfarmAndWithdraw
func NewMsgLiquidUnfarmAndWithdraw(poolId uint64, farmer string, unfarmingCoin sdk.Coin) *MsgLiquidUnfarmAndWithdraw {
	return &MsgLiquidUnfarmAndWithdraw{
		PoolId:        poolId,
		Farmer:        farmer,
		UnfarmingCoin: unfarmingCoin,
	}
}

func (msg MsgLiquidUnfarmAndWithdraw) Route() string { return RouterKey }

func (msg MsgLiquidUnfarmAndWithdraw) Type() string { return TypeMsgLiquidUnfarmAndWithdraw }

func (msg MsgLiquidUnfarmAndWithdraw) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address: %v", err)
	}
	if msg.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid pool id")
	}
	if err := msg.UnfarmingCoin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid unfarming coin: %v", err)
	}
	if !msg.UnfarmingCoin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "unfarming coin must be positive")
	}
	expCoinDenom := LiquidFarmCoinDenom(msg.PoolId)
	if msg.UnfarmingCoin.Denom != expCoinDenom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected denom %s, but got %s", expCoinDenom, msg.UnfarmingCoin.Denom)
	}
	return nil
}

func (msg MsgLiquidUnfarmAndWithdraw) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgLiquidUnfarmAndWithdraw) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgLiquidUnfarmAndWithdraw) GetFarmer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgPlaceBid creates a new MsgPlaceBid
func NewMsgPlaceBid(auctionId uint64, poolId uint64, bidder string, biddingCoin sdk.Coin) *MsgPlaceBid {
	return &MsgPlaceBid{
		AuctionId:   auctionId,
		PoolId:      poolId,
		Bidder:      bidder,
		BiddingCoin: biddingCoin,
	}
}

func (msg MsgPlaceBid) Route() string { return RouterKey }

func (msg MsgPlaceBid) Type() string { return TypeMsgPlaceBid }

func (msg MsgPlaceBid) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Bidder); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid bidder address: %v", err)
	}
	if msg.AuctionId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid auction id")
	}
	if msg.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid pool id")
	}
	if err := msg.BiddingCoin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid bidding coin: %v", err)
	}
	if !msg.BiddingCoin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "bidding amount must be positive")
	}
	poolCoinDenom := liquiditytypes.PoolCoinDenom(msg.PoolId)
	if poolCoinDenom != msg.BiddingCoin.Denom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected denom %s, but got %s", poolCoinDenom, msg.BiddingCoin.Denom)
	}
	return nil
}

func (msg MsgPlaceBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceBid) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Bidder)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceBid) GetBidder() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Bidder)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgRefundBid creates a new MsgRefundBid
func NewMsgRefundBid(auctionId uint64, poolId uint64, bidder string) *MsgRefundBid {
	return &MsgRefundBid{
		AuctionId: auctionId,
		PoolId:    poolId,
		Bidder:    bidder,
	}
}

func (msg MsgRefundBid) Route() string { return RouterKey }

func (msg MsgRefundBid) Type() string { return TypeMsgRefundBid }

func (msg MsgRefundBid) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Bidder); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid bidder address: %v", err)
	}
	if msg.AuctionId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid auction id")
	}
	if msg.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid pool id")
	}
	return nil
}

func (msg MsgRefundBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRefundBid) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Bidder)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgRefundBid) GetBidder() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Bidder)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgAdvanceAuction creates a new MsgAdvanceAuction.
func NewMsgAdvanceAuction(requesterAcc sdk.AccAddress) *MsgAdvanceAuction {
	return &MsgAdvanceAuction{
		Requester: requesterAcc.String(),
	}
}

func (msg MsgAdvanceAuction) Route() string { return RouterKey }

func (msg MsgAdvanceAuction) Type() string { return TypeMsgAdvanceAuction }

func (msg MsgAdvanceAuction) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Requester); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid requester address %q: %v", msg.Requester, err)
	}
	return nil
}

func (msg MsgAdvanceAuction) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgAdvanceAuction) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
