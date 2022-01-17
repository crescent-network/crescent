package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewDepositRequest(msg *MsgDepositBatch, pool Pool, id uint64, msgHeight int64) DepositRequest {
	return DepositRequest{
		Id:             id,
		PoolId:         msg.PoolId,
		MsgHeight:      msgHeight,
		Depositor:      msg.Depositor,
		XCoin:          msg.XCoin,
		AcceptedXCoin:  sdk.NewCoin(msg.XCoin.Denom, sdk.ZeroInt()),
		YCoin:          msg.YCoin,
		AcceptedYCoin:  sdk.NewCoin(msg.YCoin.Denom, sdk.ZeroInt()),
		MintedPoolCoin: sdk.NewCoin(pool.PoolCoinDenom, sdk.ZeroInt()),
		Succeeded:      false,
		ToBeDeleted:    false,
	}
}

func (req DepositRequest) GetDepositor() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		panic(err)
	}
	return addr
}

func (req DepositRequest) Validate() error {
	if req.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if req.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if req.MsgHeight == 0 { // TODO: is this check correct?
		return fmt.Errorf("message height must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(req.Depositor); err != nil {
		return fmt.Errorf("invalid depositor address %s: %w", req.Depositor, err)
	}
	if err := req.XCoin.Validate(); err != nil {
		return fmt.Errorf("invalid x coin %s: %w", req.XCoin, err)
	}
	if req.XCoin.IsZero() {
		return fmt.Errorf("x coin must not be 0")
	}
	if err := req.AcceptedXCoin.Validate(); err != nil {
		return fmt.Errorf("invalid accepted x coin %s: %w", req.AcceptedXCoin, err)
	}
	if err := req.YCoin.Validate(); err != nil {
		return fmt.Errorf("invalid y coin %s: %w", req.YCoin, err)
	}
	if req.YCoin.IsZero() {
		return fmt.Errorf("y coin must not be 0")
	}
	if err := req.AcceptedYCoin.Validate(); err != nil {
		return fmt.Errorf("invalid accepted y coin %s: %w", req.AcceptedYCoin, err)
	}
	if err := req.MintedPoolCoin.Validate(); err != nil {
		return fmt.Errorf("invalid minted pool coin %s: %w", req.MintedPoolCoin, err)
	}
	return nil
}

func NewWithdrawRequest(msg *MsgWithdrawBatch, pool Pool, id uint64, msgHeight int64) WithdrawRequest {
	return WithdrawRequest{
		Id:             id,
		PoolId:         msg.PoolId,
		MsgHeight:      msgHeight,
		Withdrawer:     msg.Withdrawer,
		PoolCoin:       msg.PoolCoin,
		WithdrawnXCoin: sdk.NewCoin(pool.XCoinDenom, sdk.ZeroInt()),
		WithdrawnYCoin: sdk.NewCoin(pool.YCoinDenom, sdk.ZeroInt()),
		Succeeded:      false,
		ToBeDeleted:    false,
	}
}

func (req WithdrawRequest) GetWithdrawer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.Withdrawer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (req WithdrawRequest) Validate() error {
	if req.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if req.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if req.MsgHeight == 0 { // TODO: is this check correct?
		return fmt.Errorf("message height must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(req.Withdrawer); err != nil {
		return fmt.Errorf("invalid withdrawer address %s: %w", req.Withdrawer, err)
	}
	if err := req.PoolCoin.Validate(); err != nil {
		return fmt.Errorf("invalid pool coin %s: %w", req.PoolCoin, err)
	}
	if req.PoolCoin.IsZero() {
		return fmt.Errorf("pool coin must not be 0")
	}
	if err := req.WithdrawnXCoin.Validate(); err != nil {
		return fmt.Errorf("invalid withdrawn x coin %s: %w", req.WithdrawnXCoin, err)
	}
	if err := req.WithdrawnYCoin.Validate(); err != nil {
		return fmt.Errorf("invalid withdrawn y coin %s: %w", req.WithdrawnYCoin, err)
	}
	return nil
}

func NewSwapRequest(msg *MsgSwapBatch, id uint64, pair Pair, canceledAt time.Time, msgHeight int64) SwapRequest {
	return SwapRequest{
		Id:            id,
		PairId:        pair.Id,
		MsgHeight:     msgHeight,
		Orderer:       msg.Orderer,
		Direction:     msg.GetDirection(),
		Price:         msg.Price,
		RemainingCoin: msg.OfferCoin,
		ReceivedCoin:  sdk.NewCoin(msg.DemandCoinDenom, sdk.ZeroInt()),
		BatchId:       pair.CurrentBatchId,
		CanceledAt:    canceledAt,
		Matched:       false,
		Canceled:      false,
		ToBeDeleted:   false,
	}
}

func (req SwapRequest) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (req SwapRequest) Validate() error {
	if req.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if req.PairId == 0 {
		return fmt.Errorf("pair id must not be 0")
	}
	if req.MsgHeight == 0 { // TODO: is this check correct?
		return fmt.Errorf("message height must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(req.Orderer); err != nil {
		return fmt.Errorf("invalid orderer address %s: %w", req.Orderer, err)
	}
	if req.Direction != SwapDirectionBuy && req.Direction != SwapDirectionSell {
		return fmt.Errorf("invalid direction: %s", req.Direction)
	}
	if !req.Price.IsPositive() {
		return fmt.Errorf("price must be positive: %s", req.Price)
	}
	if err := req.RemainingCoin.Validate(); err != nil {
		return fmt.Errorf("invalid remaining coin %s: %w", req.RemainingCoin, err)
	}
	if err := req.ReceivedCoin.Validate(); err != nil {
		return fmt.Errorf("invalid received coin %s: %w", req.ReceivedCoin, err)
	}
	if req.BatchId == 0 {
		return fmt.Errorf("batch id must not be 0")
	}
	if req.CanceledAt.IsZero() {
		return fmt.Errorf("no cancelation info")
	}
	return nil
}

func NewCancelSwapRequest(
	msg *MsgCancelSwapBatch, id uint64, pair Pair, msgHeight int64) CancelSwapRequest {
	return CancelSwapRequest{
		Id:            id,
		PairId:        pair.Id,
		MsgHeight:     msgHeight,
		Orderer:       msg.Orderer,
		SwapRequestId: msg.SwapRequestId,
		BatchId:       pair.CurrentBatchId,
		Succeeded:     false,
		ToBeDeleted:   false,
	}
}

func (req CancelSwapRequest) Validate() error {
	if req.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if req.PairId == 0 {
		return fmt.Errorf("pair id must not be 0")
	}
	if req.MsgHeight == 0 { // TODO: is this check correct?
		return fmt.Errorf("message height must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(req.Orderer); err != nil {
		return fmt.Errorf("invalid orderer address %s: %w", req.Orderer, err)
	}
	if req.SwapRequestId == 0 {
		return fmt.Errorf("swap request id must not be 0")
	}
	if req.BatchId == 0 {
		return fmt.Errorf("batch id must not be 0")
	}
	return nil
}
