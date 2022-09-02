package types

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewDepositRequest returns a new DepositRequest.
func NewDepositRequest(msg *MsgDeposit, pool Pool, id uint64, msgHeight int64) DepositRequest {
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

// Validate validates DepositRequest for genesis.
func (req DepositRequest) Validate() error {
	if req.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if req.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if req.MsgHeight == 0 {
		return fmt.Errorf("message height must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(req.Depositor); err != nil {
		return fmt.Errorf("invalid depositor address %s: %w", req.Depositor, err)
	}
	if err := req.DepositCoins.Validate(); err != nil {
		return fmt.Errorf("invalid deposit coins: %w", err)
	}
	if len(req.DepositCoins) == 0 || len(req.DepositCoins) > 2 {
		return fmt.Errorf("wrong number of deposit coins: %d", len(req.DepositCoins))
	}
	if err := req.AcceptedCoins.Validate(); err != nil {
		return fmt.Errorf("invalid accepted coins: %w", err)
	}
	if len(req.AcceptedCoins) > 2 {
		return fmt.Errorf("wrong number of accepted coins: %d", len(req.AcceptedCoins))
	}
	for _, coin := range req.AcceptedCoins {
		if req.DepositCoins.AmountOf(coin.Denom).IsZero() {
			return fmt.Errorf("mismatching denom pair between deposit coins and accepted coins")
		}
	}
	if err := req.MintedPoolCoin.Validate(); err != nil {
		return fmt.Errorf("invalid minted pool coin %s: %w", req.MintedPoolCoin, err)
	}
	if !req.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", req.Status)
	}
	return nil
}

// SetStatus sets the request's status.
// SetStatus is to easily find locations where the status is changed.
func (req *DepositRequest) SetStatus(status RequestStatus) {
	req.Status = status
}

// NewWithdrawRequest returns a new WithdrawRequest.
func NewWithdrawRequest(msg *MsgWithdraw, id uint64, msgHeight int64) WithdrawRequest {
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

// Validate validates WithdrawRequest for genesis.
func (req WithdrawRequest) Validate() error {
	if req.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if req.PoolId == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if req.MsgHeight == 0 {
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
	if len(req.WithdrawnCoins) > 2 {
		return fmt.Errorf("wrong number of withdrawn coins: %d", len(req.WithdrawnCoins))
	}
	if !req.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", req.Status)
	}
	return nil
}

// SetStatus sets the request's status.
// SetStatus is to easily find locations where the status is changed.
func (req *WithdrawRequest) SetStatus(status RequestStatus) {
	req.Status = status
}

// NewOrderForLimitOrder returns a new Order from MsgLimitOrder.
func NewOrderForLimitOrder(msg *MsgLimitOrder, id uint64, pair Pair, offerCoin sdk.Coin, price sdk.Dec, expireAt time.Time, msgHeight int64) Order {
	return Order{
		Type:               OrderTypeLimit,
		Id:                 id,
		PairId:             pair.Id,
		MsgHeight:          msgHeight,
		Orderer:            msg.Orderer,
		Direction:          msg.Direction,
		OfferCoin:          offerCoin,
		RemainingOfferCoin: offerCoin,
		ReceivedCoin:       sdk.NewCoin(msg.DemandCoinDenom, sdk.ZeroInt()),
		Price:              price,
		Amount:             msg.Amount,
		OpenAmount:         msg.Amount,
		BatchId:            pair.CurrentBatchId,
		ExpireAt:           expireAt,
		Status:             OrderStatusNotExecuted,
	}
}

// NewOrderForMarketOrder returns a new Order from MsgMarketOrder.
func NewOrderForMarketOrder(msg *MsgMarketOrder, id uint64, pair Pair, offerCoin sdk.Coin, price sdk.Dec, expireAt time.Time, msgHeight int64) Order {
	return Order{
		Type:               OrderTypeMarket,
		Id:                 id,
		PairId:             pair.Id,
		MsgHeight:          msgHeight,
		Orderer:            msg.Orderer,
		Direction:          msg.Direction,
		OfferCoin:          offerCoin,
		RemainingOfferCoin: offerCoin,
		ReceivedCoin:       sdk.NewCoin(msg.DemandCoinDenom, sdk.ZeroInt()),
		Price:              price,
		Amount:             msg.Amount,
		OpenAmount:         msg.Amount,
		BatchId:            pair.CurrentBatchId,
		ExpireAt:           expireAt,
		Status:             OrderStatusNotExecuted,
	}
}

func NewOrder(
	typ OrderType, id uint64, pair Pair, orderer sdk.AccAddress,
	offerCoin sdk.Coin, price sdk.Dec, amt sdk.Int, expireAt time.Time, msgHeight int64) Order {
	var (
		dir             OrderDirection
		demandCoinDenom string
	)
	if offerCoin.Denom == pair.BaseCoinDenom {
		dir = OrderDirectionSell
		demandCoinDenom = pair.QuoteCoinDenom
	} else {
		dir = OrderDirectionBuy
		demandCoinDenom = pair.BaseCoinDenom
	}
	return Order{
		Type:               typ,
		Id:                 id,
		PairId:             pair.Id,
		MsgHeight:          msgHeight,
		Orderer:            orderer.String(),
		Direction:          dir,
		OfferCoin:          offerCoin,
		RemainingOfferCoin: offerCoin,
		ReceivedCoin:       sdk.NewCoin(demandCoinDenom, sdk.ZeroInt()),
		Price:              price,
		Amount:             amt,
		OpenAmount:         amt,
		BatchId:            pair.CurrentBatchId,
		ExpireAt:           expireAt,
		Status:             OrderStatusNotExecuted,
	}
}

func (order Order) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(order.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}

// Validate validates Order for genesis.
func (order Order) Validate() error {
	if order.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if order.PairId == 0 {
		return fmt.Errorf("pair id must not be 0")
	}
	if order.MsgHeight == 0 {
		return fmt.Errorf("message height must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(order.Orderer); err != nil {
		return fmt.Errorf("invalid orderer address %s: %w", order.Orderer, err)
	}
	if order.Direction != OrderDirectionBuy && order.Direction != OrderDirectionSell {
		return fmt.Errorf("invalid direction: %s", order.Direction)
	}
	if err := order.OfferCoin.Validate(); err != nil {
		return fmt.Errorf("invalid offer coin %s: %w", order.OfferCoin, err)
	}
	if order.OfferCoin.IsZero() {
		return fmt.Errorf("offer coin must not be 0")
	}
	if err := order.RemainingOfferCoin.Validate(); err != nil {
		return fmt.Errorf("invalid remaining offer coin %s: %w", order.RemainingOfferCoin, err)
	}
	if order.OfferCoin.Denom != order.RemainingOfferCoin.Denom {
		return fmt.Errorf("offer coin denom %s != remaining offer coin denom %s", order.OfferCoin.Denom, order.RemainingOfferCoin.Denom)
	}
	if err := order.ReceivedCoin.Validate(); err != nil {
		return fmt.Errorf("invalid received coin %s: %w", order.ReceivedCoin, err)
	}
	if !order.Price.IsPositive() {
		return fmt.Errorf("price must be positive: %s", order.Price)
	}
	if !order.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive: %s", order.Amount)
	}
	if order.OpenAmount.IsNegative() {
		return fmt.Errorf("open amount must not be negative: %s", order.OpenAmount)
	}
	if order.BatchId == 0 {
		return fmt.Errorf("batch id must not be 0")
	}
	if order.ExpireAt.IsZero() {
		return fmt.Errorf("no expiration info")
	}
	if !order.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", order.Status)
	}
	return nil
}

// ExpiredAt returns whether the order should be deleted at given time.
func (order Order) ExpiredAt(t time.Time) bool {
	return !order.ExpireAt.After(t)
}

// SetStatus sets the order's status.
// SetStatus is to easily find locations where the status is changed.
func (order *Order) SetStatus(status OrderStatus) {
	order.Status = status
}

// IsValid returns true if the RequestStatus is one of:
// RequestStatusNotExecuted, RequestStatusSucceeded, RequestStatusFailed.
func (status RequestStatus) IsValid() bool {
	switch status {
	case RequestStatusNotExecuted, RequestStatusSucceeded, RequestStatusFailed:
		return true
	default:
		return false
	}
}

// ShouldBeDeleted returns true if the RequestStatus is one of:
// RequestStatusSucceeded, RequestStatusFailed.
func (status RequestStatus) ShouldBeDeleted() bool {
	switch status {
	case RequestStatusSucceeded, RequestStatusFailed:
		return true
	default:
		return false
	}
}

// IsValid returns true if the OrderStatus is one of:
// OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched,
// OrderStatusCompleted, OrderStatusCanceled, OrderStatusExpired.
func (status OrderStatus) IsValid() bool {
	switch status {
	case OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched,
		OrderStatusCompleted, OrderStatusCanceled, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// IsMatchable returns true if the OrderStatus is one of:
// OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched.
func (status OrderStatus) IsMatchable() bool {
	switch status {
	case OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched:
		return true
	default:
		return false
	}
}

// CanBeExpired has the same condition as IsMatchable.
func (status OrderStatus) CanBeExpired() bool {
	return status.IsMatchable()
}

// CanBeCanceled returns true if the OrderStatus is one of:
// OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched.
func (status OrderStatus) CanBeCanceled() bool {
	switch status {
	case OrderStatusNotExecuted, OrderStatusNotMatched, OrderStatusPartiallyMatched:
		return true
	default:
		return false
	}
}

// IsCanceledOrExpired returns true if the OrderStatus is one of:
// OrderStatusCanceled, OrderStatusExpired.
func (status OrderStatus) IsCanceledOrExpired() bool {
	switch status {
	case OrderStatusCanceled, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// ShouldBeDeleted returns true if the OrderStatus is one of:
// OrderStatusCompleted, OrderStatusCanceled, OrderStatusExpired.
func (status OrderStatus) ShouldBeDeleted() bool {
	return status == OrderStatusCompleted || status.IsCanceledOrExpired()
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

// MustMarshaOrder returns the Order bytes.
// It throws panic if it fails.
func MustMarshaOrder(cdc codec.BinaryCodec, order Order) []byte {
	return cdc.MustMarshal(&order)
}

// UnmarshalOrder returns the Order from bytes.
func UnmarshalOrder(cdc codec.BinaryCodec, value []byte) (order Order, err error) {
	err = cdc.Unmarshal(value, &order)
	return order, err
}

// MustUnmarshalOrder returns the Order from bytes.
// It throws panic if it fails.
func MustUnmarshalOrder(cdc codec.BinaryCodec, value []byte) Order {
	msg, err := UnmarshalOrder(cdc, value)
	if err != nil {
		panic(err)
	}
	return msg
}
