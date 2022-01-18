package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewDepositRequest(msg *MsgDepositBatch, pool Pool, id uint64, msgHeight int64) DepositRequest {
	return DepositRequest{
		Id:             id,
		PoolId:         msg.PoolId,
		MsgHeight:      msgHeight,
		Depositor:      msg.Depositor,
		DepositCoins:   msg.DepositCoins,
		AcceptedCoins:  sdk.Coins{},
		MintedPoolCoin: sdk.NewCoin(pool.PoolCoinDenom, sdk.ZeroInt()),
		Status:         RequestStatusNotExecuted,
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
		Id:             id,
		PoolId:         msg.PoolId,
		MsgHeight:      msgHeight,
		Withdrawer:     msg.Withdrawer,
		PoolCoin:       msg.PoolCoin,
		WithdrawnCoins: sdk.Coins{},
		Status:         RequestStatusNotExecuted,
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
		Status:        SwapRequestStatusNotExecuted,
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
		Status:        RequestStatusNotExecuted,
	}
}

func (status RequestStatus) ShouldBeDeleted() bool {
	return status == RequestStatusSucceeded || status == RequestStatusFailed
}

func (status SwapRequestStatus) IsMatchable() bool {
	return status == SwapRequestStatusNotExecuted || status == SwapRequestStatusExecuted
}

func (status SwapRequestStatus) IsCanceledOrExpired() bool {
	return status == SwapRequestStatusCanceled || status == SwapRequestStatusExpired
}

func (status SwapRequestStatus) ShouldBeDeleted() bool {
	return status.IsCanceledOrExpired()
}