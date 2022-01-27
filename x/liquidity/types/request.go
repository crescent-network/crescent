package types

import (
	"fmt"
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
		AcceptedCoins:  nil,
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
	if err := req.DepositCoins.Validate(); err != nil {
		return fmt.Errorf("invalid deposit coins: %w", err)
	}
	if len(req.DepositCoins) != 2 {
		return fmt.Errorf("wrong number of deposit coins: %d", len(req.DepositCoins))
	}
	if err := req.AcceptedCoins.Validate(); err != nil {
		return fmt.Errorf("invalid accepted coins: %w", err)
	}
	if len(req.AcceptedCoins) != 0 && len(req.AcceptedCoins) != 2 {
		return fmt.Errorf("wrong number of accepted coins: %d", len(req.AcceptedCoins))
	}
	if err := req.MintedPoolCoin.Validate(); err != nil {
		return fmt.Errorf("invalid minted pool coin %s: %w", req.MintedPoolCoin, err)
	}
	if !req.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", req.Status)
	}
	return nil
}

func NewWithdrawRequest(msg *MsgWithdrawBatch, id uint64, msgHeight int64) WithdrawRequest {
	return WithdrawRequest{
		Id:             id,
		PoolId:         msg.PoolId,
		MsgHeight:      msgHeight,
		Withdrawer:     msg.Withdrawer,
		PoolCoin:       msg.PoolCoin,
		WithdrawnCoins: nil,
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
	if err := req.WithdrawnCoins.Validate(); err != nil {
		return fmt.Errorf("invalid withdrawn coins: %w", err)
	}
	if len(req.WithdrawnCoins) != 0 && len(req.WithdrawnCoins) != 2 {
		return fmt.Errorf("wrong number of withdrawn coins: %d", len(req.WithdrawnCoins))
	}
	if !req.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", req.Status)
	}
	return nil
}

func NewSwapRequestForLimitOrder(msg *MsgLimitOrderBatch, id uint64, pair Pair, offerCoin sdk.Coin, expireAt time.Time, msgHeight int64) SwapRequest {
	return SwapRequest{
		Id:                 id,
		PairId:             pair.Id,
		MsgHeight:          msgHeight,
		Orderer:            msg.Orderer,
		Direction:          msg.Direction,
		OfferCoin:          offerCoin,
		RemainingOfferCoin: offerCoin,
		ReceivedCoin:       sdk.NewCoin(msg.DemandCoinDenom, sdk.ZeroInt()),
		Price:              msg.Price,
		Amount:             msg.Amount,
		OpenAmount:         msg.Amount,
		BatchId:            pair.CurrentBatchId,
		ExpireAt:           expireAt,
		Status:             SwapRequestStatusNotExecuted,
	}
}

func NewSwapRequestForMarketOrder(msg *MsgMarketOrderBatch, id uint64, pair Pair, price sdk.Dec, expireAt time.Time, msgHeight int64) SwapRequest {
	return SwapRequest{
		Id:                 id,
		PairId:             pair.Id,
		MsgHeight:          msgHeight,
		Orderer:            msg.Orderer,
		Direction:          msg.Direction,
		OfferCoin:          msg.OfferCoin,
		RemainingOfferCoin: msg.OfferCoin,
		ReceivedCoin:       sdk.NewCoin(msg.DemandCoinDenom, sdk.ZeroInt()),
		Price:              price,
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
	if err := req.OfferCoin.Validate(); err != nil {
		return fmt.Errorf("invalid offer coin %s: %w", req.OfferCoin, err)
	}
	if req.OfferCoin.IsZero() {
		return fmt.Errorf("offer coin must not be 0")
	}
	if err := req.RemainingOfferCoin.Validate(); err != nil {
		return fmt.Errorf("invalid remaining offer coin %s: %w", req.RemainingOfferCoin, err)
	}
	if err := req.ReceivedCoin.Validate(); err != nil {
		return fmt.Errorf("invalid received coin %s: %w", req.ReceivedCoin, err)
	}
	if !req.Price.IsPositive() {
		return fmt.Errorf("price must be positive: %s", req.Price)
	}
	if !req.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive: %s", req.Amount)
	}
	if req.OpenAmount.IsNegative() {
		return fmt.Errorf("open amount must not be negative: %s", req.OpenAmount)
	}
	if req.BatchId == 0 {
		return fmt.Errorf("batch id must not be 0")
	}
	if req.ExpireAt.IsZero() {
		return fmt.Errorf("no expiration info")
	}
	if !req.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", req.Status)
	}
	return nil
}

func NewCancelOrderRequest(
	msg *MsgCancelOrderBatch, id uint64, pair Pair, msgHeight int64) CancelOrderRequest {
	return CancelOrderRequest{
		Id:            id,
		PairId:        pair.Id,
		MsgHeight:     msgHeight,
		Orderer:       msg.Orderer,
		SwapRequestId: msg.SwapRequestId,
		BatchId:       pair.CurrentBatchId,
		Status:        RequestStatusNotExecuted,
	}
}

func (req CancelOrderRequest) Validate() error {
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
	if !req.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", req.Status)
	}
	return nil
}

func (status RequestStatus) IsValid() bool {
	return status == RequestStatusNotExecuted || status == RequestStatusSucceeded || status == RequestStatusFailed
}

func (status RequestStatus) ShouldBeDeleted() bool {
	return status == RequestStatusSucceeded || status == RequestStatusFailed
}

func (status SwapRequestStatus) IsValid() bool {
	return status == SwapRequestStatusNotExecuted || status == SwapRequestStatusNotMatched ||
		status == SwapRequestStatusPartiallyMatched || status == SwapRequestStatusCompleted ||
		status == SwapRequestStatusCanceled || status == SwapRequestStatusExpired
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
