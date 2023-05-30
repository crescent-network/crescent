package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var (
	_ sdk.Msg = (*MsgCreatePool)(nil)
	_ sdk.Msg = (*MsgAddLiquidity)(nil)
	_ sdk.Msg = (*MsgRemoveLiquidity)(nil)
	_ sdk.Msg = (*MsgCollect)(nil)
	_ sdk.Msg = (*MsgCreatePrivateFarmingPlan)(nil)
)

// Message types for the module
const (
	TypeMsgCreatePool               = "create_pool"
	TypeMsgAddLiquidity             = "add_liquidity"
	TypeMsgRemoveLiquidity          = "remove_liquidity"
	TypeMsgCollect                  = "collect"
	TypeMsgCreatePrivateFarmingPlan = "create_private_farming_plan"
)

func NewMsgCreatePool(
	senderAddr sdk.AccAddress, marketId uint64, price sdk.Dec) *MsgCreatePool {
	return &MsgCreatePool{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		Price:    price,
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
	if !msg.Price.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price must be positive: %s", msg.Price)
	}
	if msg.Price.LT(exchangetypes.MinPrice) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price is lower than the min price %s", exchangetypes.MinPrice)
	}
	if msg.Price.GT(exchangetypes.MaxPrice) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price is higher than the max price %s", exchangetypes.MaxPrice)
	}
	return nil
}

func NewMsgAddLiquidity(
	senderAddr sdk.AccAddress, poolId uint64, lowerPrice, upperPrice sdk.Dec,
	desiredAmt sdk.Coins) *MsgAddLiquidity {
	return &MsgAddLiquidity{
		Sender:        senderAddr.String(),
		PoolId:        poolId,
		LowerPrice:    lowerPrice,
		UpperPrice:    upperPrice,
		DesiredAmount: desiredAmt,
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
	if msg.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pool is must not be 0")
	}
	if !msg.LowerPrice.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "lower price must be positive: %s", msg.LowerPrice)
	}
	if !msg.UpperPrice.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "upper price must be positive: %s", msg.UpperPrice)
	}
	if msg.LowerPrice.GTE(msg.UpperPrice) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "lower price must be lower than upper price")
	}
	if _, valid := exchangetypes.ValidateTickPrice(msg.LowerPrice); !valid {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid lower tick")
	}
	if _, valid := exchangetypes.ValidateTickPrice(msg.UpperPrice); !valid {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid upper tick")
	}
	if err := msg.DesiredAmount.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid desired amount: %v", err)
	}
	if len(msg.DesiredAmount) == 0 || len(msg.DesiredAmount) > 2 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid desired amount length: %d", len(msg.DesiredAmount))
	}
	return nil
}

func NewMsgRemoveLiquidity(
	senderAddr sdk.AccAddress, positionId uint64, liquidity sdk.Int) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		Sender:     senderAddr.String(),
		PositionId: positionId,
		Liquidity:  liquidity,
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
	if msg.PositionId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "position is must not be 0")
	}
	if !msg.Liquidity.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "liquidity must be positive: %s", msg.Liquidity)
	}
	return nil
}

func NewMsgCollect(senderAddr sdk.AccAddress, positionId uint64, amt sdk.Coins) *MsgCollect {
	return &MsgCollect{
		Sender:     senderAddr.String(),
		PositionId: positionId,
		Amount:     amt,
	}
}

func (msg MsgCollect) Route() string { return RouterKey }
func (msg MsgCollect) Type() string  { return TypeMsgCollect }

func (msg MsgCollect) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCollect) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCollect) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.PositionId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "position is must not be 0")
	}
	if err := msg.Amount.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid amount: %v", err)
	}
	return nil
}

func NewMsgCreatePrivateFarmingPlan(
	senderAddr sdk.AccAddress, description string, termAddr sdk.AccAddress, rewardAllocs []FarmingRewardAllocation,
	startTime, endTime time.Time) *MsgCreatePrivateFarmingPlan {
	return &MsgCreatePrivateFarmingPlan{
		Sender:             senderAddr.String(),
		Description:        description,
		TerminationAddress: termAddr.String(),
		RewardAllocations:  rewardAllocs,
		StartTime:          startTime,
		EndTime:            endTime,
	}
}

func (msg MsgCreatePrivateFarmingPlan) Route() string { return RouterKey }
func (msg MsgCreatePrivateFarmingPlan) Type() string  { return TypeMsgCreatePrivateFarmingPlan }

func (msg MsgCreatePrivateFarmingPlan) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreatePrivateFarmingPlan) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreatePrivateFarmingPlan) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	// Create a dummy plan with valid fields and utilize Validate() method
	// for user-provided data.
	farmingPoolAddr := RewardsPoolAddress // Chose random valid address
	dummyPlan := NewFarmingPlan(
		1, msg.Description, farmingPoolAddr, sdk.MustAccAddressFromBech32(msg.TerminationAddress),
		msg.RewardAllocations, msg.StartTime, msg.EndTime, true)
	if err := dummyPlan.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}
