package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewDepositRequest(msg *MsgDepositBatch, id uint64, msgHeight int64) DepositRequest {
	return DepositRequest{
		Id:          id,
		PoolId:      msg.PoolId,
		MsgHeight:   msgHeight,
		Depositor:   msg.Depositor,
		XCoin:       msg.XCoin,
		YCoin:       msg.YCoin,
		Succeeded:   false,
		ToBeDeleted: false,
	}
}

func (req DepositRequest) GetDepositor() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		panic(err)
	}
	return addr
}

func NewWithdrawRequest(msg *MsgWithdrawBatch, id uint64, msgHeight int64) WithdrawRequest {
	return WithdrawRequest{
		Id:          id,
		PoolId:      msg.PoolId,
		MsgHeight:   msgHeight,
		Withdrawer:  msg.Withdrawer,
		PoolCoin:    msg.PoolCoin,
		Succeeded:   false,
		ToBeDeleted: false,
	}
}

func (req WithdrawRequest) GetWithdrawer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.Withdrawer)
	if err != nil {
		panic(err)
	}
	return addr
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
