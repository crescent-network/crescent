package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// SwapBatch handles types.MsgSwapBatch and stores it.
func (k Keeper) SwapBatch(ctx sdk.Context, msg *types.MsgSwapBatch) error {
	params := k.GetParams(ctx)
	if price := types.PriceToTick(msg.Price, int(params.TickPrecision)); !msg.Price.Equal(price) {
		return types.ErrInvalidPriceTick
	}

	var pair types.Pair
	pair, found := k.GetPairByDenoms(ctx, msg.XCoinDenom, msg.YCoinDenom)
	if !found {
		pair = k.CreatePair(ctx, msg.XCoinDenom, msg.YCoinDenom)
	}

	if pair.LastPrice != nil {
		lastPrice := *pair.LastPrice
		switch {
		case msg.Price.GT(lastPrice):
			priceLimit := msg.Price.Mul(sdk.OneDec().Add(params.MaxPriceLimitRatio))
			if msg.Price.GT(priceLimit) {
				return types.ErrPriceOutOfRange
			}
		case msg.Price.LT(lastPrice):
			priceLimit := msg.Price.Mul(sdk.OneDec().Sub(params.MaxPriceLimitRatio))
			if msg.Price.LT(priceLimit) {
				return types.ErrPriceOutOfRange
			}
		}
	}

	if err := k.bankKeeper.SendCoins(ctx, msg.GetOrderer(), pair.GetEscrowAddress(), sdk.NewCoins(msg.OfferCoin)); err != nil {
		return err
	}

	requestId := k.GetNextSwapRequestIdWithUpdate(ctx, pair)
	req := types.SwapRequest{
		Id:              requestId,
		PairId:          pair.Id,
		MsgHeight:       ctx.BlockHeight(),
		BatchId:         pair.CurrentBatchId,
		Orderer:         msg.Orderer,
		Direction:       msg.GetDirection(),
		Price:           msg.Price,
		RemainingAmount: msg.OfferCoin.Amount,
		ReceivedAmount:  sdk.ZeroInt(),
	}
	k.SetSwapRequest(ctx, pair.Id, req)

	// TODO: need to emit an event?

	return nil
}

// CancelSwapBatch handles types.MsgCancelSwapBatch and stores it.
func (k Keeper) CancelSwapBatch(ctx sdk.Context, msg *types.MsgCancelSwapBatch) error {
	swapReq, found := k.GetSwapRequest(ctx, msg.PairId, msg.SwapRequestId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "swap request with id %d in pair %d not found", msg.SwapRequestId, msg.PairId)
	}

	if msg.Orderer != swapReq.Orderer {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "mismatching orderer")
	}

	pair, found := k.GetPair(ctx, msg.PairId)
	if !found { // TODO: will it ever happen?
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair with id %d not found", msg.PairId)
	}

	requestId := k.GetNextCancelSwapRequestIdWithUpdate(ctx, pair)
	req := types.CancelSwapRequest{
		Id:            requestId,
		PairId:        msg.PairId,
		MsgHeight:     ctx.BlockHeight(),
		Orderer:       msg.Orderer,
		SwapRequestId: msg.SwapRequestId,
		BatchId:       pair.CurrentBatchId,
	}
	k.SetCancelSwapRequest(ctx, msg.PairId, req)

	return nil
}
