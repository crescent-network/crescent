package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

// GetNextPairIdWithUpdate increments pair id by one and set it.
func (k Keeper) GetNextPairIdWithUpdate(ctx sdk.Context) uint64 {
	id := k.GetLastPairId(ctx) + 1
	k.SetLastPairId(ctx, id)
	return id
}

// GetNextSwapRequestIdWithUpdate increments the pair's last swap request
// id and returns it.
func (k Keeper) GetNextSwapRequestIdWithUpdate(ctx sdk.Context, pair types.Pair) uint64 {
	id := pair.LastSwapRequestId + 1
	pair.LastSwapRequestId = id
	k.SetPair(ctx, pair)
	return id
}

// GetNextCancelOrderRequestIdWithUpdate increments the pair's last cancel
// swap request id and returns it.
func (k Keeper) GetNextCancelOrderRequestIdWithUpdate(ctx sdk.Context, pair types.Pair) uint64 {
	id := pair.LastCancelOrderRequestId + 1
	pair.LastCancelOrderRequestId = id
	k.SetPair(ctx, pair)
	return id
}

// CreatePair handles types.MsgCreatePair and creates a pair.
func (k Keeper) CreatePair(ctx sdk.Context, msg *types.MsgCreatePair) (types.Pair, error) {
	params := k.GetParams(ctx)

	if _, found := k.GetPairByDenoms(ctx, msg.BaseCoinDenom, msg.QuoteCoinDenom); found {
		return types.Pair{}, types.ErrPairAlreadyExists
	}

	// Send the pair creation fee to the fee collector.
	feeCollectorAddr, _ := sdk.AccAddressFromBech32(params.FeeCollectorAddress)
	if err := k.bankKeeper.SendCoins(ctx, msg.GetCreator(), feeCollectorAddr, params.PairCreationFee); err != nil {
		return types.Pair{}, sdkerrors.Wrap(err, "insufficient pair creation fee")
	}

	id := k.GetNextPairIdWithUpdate(ctx)
	pair := types.NewPair(id, msg.BaseCoinDenom, msg.QuoteCoinDenom)
	k.SetPair(ctx, pair)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePair,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyBaseCoinDenom, msg.BaseCoinDenom),
			sdk.NewAttribute(types.AttributeKeyQuoteCoinDenom, msg.QuoteCoinDenom),
		),
	})

	return pair, nil
}
