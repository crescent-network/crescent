package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
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

func NewSwapRequest(msg *MsgSwapBatch, id uint64, pair Pair, expireAt time.Time, msgHeight int64) SwapRequest {
	return SwapRequest{
		Id:                 id,
		PairId:             pair.Id,
		MsgHeight:          msgHeight,
		Orderer:            msg.Orderer,
		Direction:          msg.Direction,
		OfferCoin:          msg.OfferCoin,
		RemainingOfferCoin: msg.OfferCoin,
		ReceivedCoin:       sdk.NewCoin(msg.DemandCoinDenom, sdk.ZeroInt()),
		Price:              msg.Price,
		Amount:             msg.Amount,
		OpenAmount:         msg.Amount,
		BatchId:            pair.CurrentBatchId,
		ExpireAt:           expireAt,
		Status:             SwapRequestStatusNotExecuted,
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
	return status == SwapRequestStatusNotExecuted ||
		status == SwapRequestStatusNotMatched ||
		status == SwapRequestStatusPartiallyMatched
}

func (status SwapRequestStatus) IsCanceledOrExpired() bool {
	return status == SwapRequestStatusCanceled || status == SwapRequestStatusExpired
}

func (status SwapRequestStatus) ShouldBeDeleted() bool {
	return status == SwapRequestStatusCompleted || status.IsCanceledOrExpired()
}

// MustMarshalDepositRequest returns the DepositRequest bytes. Panics if fails.
func MustMarshalDepositRequest(cdc codec.BinaryCodec, msg DepositRequest) []byte {
	return cdc.MustMarshal(&msg)
}

// UnmarshalDepositRequest returns the DepositRequest from bytes.
func UnmarshalDepositRequest(cdc codec.BinaryCodec, value []byte) (msg DepositRequest, err error) {
	err = cdc.Unmarshal(value, &msg)
	return msg, err
}

// MustUnmarshalDepositRequest returns the DepositRequest from bytes.
// It throws panic if it fails.
func MustUnmarshalDepositRequest(cdc codec.BinaryCodec, value []byte) DepositRequest {
	msg, err := UnmarshalDepositRequest(cdc, value)
	if err != nil {
		panic(err)
	}
	return msg
}

// MustMarshaWithdrawRequest returns the WithdrawRequest bytes.
// It throws panic if it fails.
func MustMarshaWithdrawRequest(cdc codec.BinaryCodec, msg WithdrawRequest) []byte {
	return cdc.MustMarshal(&msg)
}

// UnmarshalWithdrawRequest returns the WithdrawRequest from bytes.
func UnmarshalWithdrawRequest(cdc codec.BinaryCodec, value []byte) (msg WithdrawRequest, err error) {
	err = cdc.Unmarshal(value, &msg)
	return msg, err
}

// MustUnmarshalWithdrawRequest returns the WithdrawRequest from bytes.
// It throws panic if it fails.
func MustUnmarshalWithdrawRequest(cdc codec.BinaryCodec, value []byte) WithdrawRequest {
	msg, err := UnmarshalWithdrawRequest(cdc, value)
	if err != nil {
		panic(err)
	}
	return msg
}

// MustMarshaSwapRequest returns the SwapRequest bytes.
// It throws panic if it fails.
func MustMarshaSwapRequest(cdc codec.BinaryCodec, msg SwapRequest) []byte {
	return cdc.MustMarshal(&msg)
}

// UnmarshalSwapRequest returns the SwapRequest from bytes.
func UnmarshalSwapRequest(cdc codec.BinaryCodec, value []byte) (msg SwapRequest, err error) {
	err = cdc.Unmarshal(value, &msg)
	return msg, err
}

// MustUnmarshalSwapRequest returns the SwapRequest from bytes.
// It throws panic if it fails.
func MustUnmarshalSwapRequest(cdc codec.BinaryCodec, value []byte) SwapRequest {
	msg, err := UnmarshalSwapRequest(cdc, value)
	if err != nil {
		panic(err)
	}
	return msg
}
