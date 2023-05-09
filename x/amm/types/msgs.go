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
	_ sdk.Msg = (*MsgHarvest)(nil)
)

// Message types for the module
const (
	TypeMsgCreatePool               = "create_pool"
	TypeMsgAddLiquidity             = "add_liquidity"
	TypeMsgRemoveLiquidity          = "remove_liquidity"
	TypeMsgCollect                  = "collect"
	TypeMsgCreatePrivateFarmingPlan = "create_private_farming_plan"
	TypeMsgHarvest                  = "harvest"
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
	if msg.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pool is must not be 0")
	}
	if !msg.LowerPrice.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "lower price must be positive: %s", msg.LowerPrice)
	}
	if !msg.UpperPrice.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "upper price must be positive: %s", msg.UpperPrice)
	}
	if !msg.DesiredAmount0.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "desired amount 0 must be positive: %s", msg.DesiredAmount0)
	}
	if !msg.DesiredAmount1.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "desired amount 1 must be positive: %s", msg.DesiredAmount1)
	}
	if msg.MinAmount0.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "min amount 0 must not be negative: %s", msg.MinAmount0)
	}
	if msg.MinAmount1.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "min amount 1 must not be negative: %s", msg.MinAmount1)
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
	if msg.PositionId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "position is must not be 0")
	}
	if msg.Liquidity.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "liquidity must not be negative: %s", msg.Liquidity)
	}
	if msg.MinAmount0.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "min amount 0 must not be negative: %s", msg.MinAmount0)
	}
	if msg.MinAmount1.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "min amount 1 must not be negative: %s", msg.MinAmount1)
	}
	return nil
}

func NewMsgCollect(senderAddr sdk.AccAddress, positionId uint64, maxAmt0, maxAmt1 sdk.Int) *MsgCollect {
	return &MsgCollect{
		Sender:     senderAddr.String(),
		PositionId: positionId,
		MaxAmount0: maxAmt0,
		MaxAmount1: maxAmt1,
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
	if msg.MaxAmount0.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "max amount 0 must not be negative: %s", msg.MaxAmount0)
	}
	if msg.MaxAmount1.IsNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "max amount 1 must not be negative: %s", msg.MaxAmount1)
	}
	return nil
}

func NewMsgCreatePrivateFarmingPlan(
	senderAddr sdk.AccAddress, description string, rewardAllocations []RewardAllocation,
	startTime, endTime time.Time) *MsgCreatePrivateFarmingPlan {
	return &MsgCreatePrivateFarmingPlan{
		Sender:            senderAddr.String(),
		Description:       description,
		RewardAllocations: rewardAllocations,
		StartTime:         startTime,
		EndTime:           endTime,
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
	validAddr := RewardsPoolAddress // Chose random valid address
	dummyPlan := NewFarmingPlan(
		1, msg.Description, validAddr, validAddr,
		msg.RewardAllocations, msg.StartTime, msg.EndTime, true)
	if err := dummyPlan.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

func NewMsgHarvest(
	senderAddr sdk.AccAddress, positionId uint64) *MsgHarvest {
	return &MsgHarvest{
		Sender:     senderAddr.String(),
		PositionId: positionId,
	}
}

func (msg MsgHarvest) Route() string { return RouterKey }
func (msg MsgHarvest) Type() string  { return TypeMsgHarvest }

func (msg MsgHarvest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgHarvest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgHarvest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if msg.PositionId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "position id must not be 0")
	}
	return nil
}
