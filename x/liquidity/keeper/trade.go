package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/crescent-network/crescent/v3/x/liquidity/types"
)

var (
	LastTradeIdKey               = []byte{0xa2}
	TradeKeyPrefix               = []byte{0xa3}
	TradeIndexKeyPrefix          = []byte{0xa4}
	TradeIndexByOrdererKeyPrefix = []byte{0xa8}
)

func GetTradeKey(id uint64) []byte {
	return append(TradeKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func GetTradeIndexKey(pairId, orderId uint64, height int64) []byte {
	return append(append(append(TradeIndexKeyPrefix,
		sdk.Uint64ToBigEndian(pairId)...),
		sdk.Uint64ToBigEndian(orderId)...),
		sdk.Uint64ToBigEndian(uint64(height))...)
}

func GetTradeIndexByOrdererKey(ordererAddr sdk.AccAddress, pairId, orderId uint64, height int64) []byte {
	return append(append(append(append(TradeIndexByOrdererKeyPrefix,
		address.MustLengthPrefix(ordererAddr)...),
		sdk.Uint64ToBigEndian(pairId)...),
		sdk.Uint64ToBigEndian(orderId)...),
		sdk.Uint64ToBigEndian(uint64(height))...)
}

func (k Keeper) GetLastTradeId() (lastTradeId uint64, found bool) {
	store := dbadapter.Store{DB: k.offChainDB}
	bz := store.Get(LastTradeIdKey)
	if bz == nil {
		return
	}
	return sdk.BigEndianToUint64(bz), true
}

func (k Keeper) SetLastTradeId(id uint64) {
	store := dbadapter.Store{DB: k.offChainDB}
	store.Set(LastTradeIdKey, sdk.Uint64ToBigEndian(id))
}

func (k Keeper) SetTrade(trade types.Trade) {
	store := dbadapter.Store{DB: k.offChainDB}
	bz := k.cdc.MustMarshal(&trade)
	store.Set(GetTradeKey(trade.Id), bz)
	store.Set(GetTradeIndexKey(trade.Order.PairId, trade.Order.Id, trade.Height), sdk.Uint64ToBigEndian(trade.Id))
	store.Set(GetTradeIndexByOrdererKey(
		trade.Order.GetOrderer(), trade.Order.PairId, trade.Order.Id, trade.Height),
		sdk.Uint64ToBigEndian(trade.Id))
}

func (k Keeper) GetTrade(id uint64) (trade types.Trade, found bool) {
	store := dbadapter.Store{DB: k.offChainDB}
	bz := store.Get(GetTradeKey(id))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &trade)
	return trade, true
}

func (k Querier) Trades(c context.Context, req *types.QueryTradesRequest) (*types.QueryTradesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if k.offChainDB == nil {
		return nil, status.Error(codes.Unavailable, "trades endpoint is disabled for this node")
	}

	var (
		keyPrefix []byte
	)
	if req.Orderer != "" {
		ordererAddr, err := sdk.AccAddressFromBech32(req.Orderer)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid orderer: %v", err)
		}
		keyPrefix = append(TradeIndexByOrdererKeyPrefix, address.MustLengthPrefix(ordererAddr)...)
		if req.PairId != 0 {
			keyPrefix = append(keyPrefix, sdk.Uint64ToBigEndian(req.PairId)...)
		} else if req.OrderId != 0 {
			return nil, status.Error(codes.InvalidArgument, "pair id must be specified with order id")
		}
		if req.OrderId != 0 {
			keyPrefix = append(keyPrefix, sdk.Uint64ToBigEndian(req.OrderId)...)
		}
	} else if req.PairId != 0 {
		keyPrefix = append(TradeIndexKeyPrefix, sdk.Uint64ToBigEndian(req.PairId)...)
		if req.OrderId != 0 {
			keyPrefix = append(keyPrefix, sdk.Uint64ToBigEndian(req.OrderId)...)
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "orderer or pair id must be specified")
	}

	store := dbadapter.Store{DB: k.offChainDB}
	tradeStore := prefix.NewStore(store, keyPrefix)

	var trades []types.Trade
	pageRes, err := query.Paginate(tradeStore, req.Pagination, func(key, value []byte) error {
		trade, _ := k.GetTrade(sdk.BigEndianToUint64(value))
		trades = append(trades, trade)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTradesResponse{Trades: trades, Pagination: pageRes}, nil
}

func (k Querier) Trade(c context.Context, req *types.QueryTradeRequest) (*types.QueryTradeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if k.offChainDB == nil {
		return nil, status.Error(codes.Unavailable, "trade endpoint is disabled for this node")
	}

	trade, found := k.GetTrade(req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "trade not found")
	}

	return &types.QueryTradeResponse{Trade: trade}, nil
}
