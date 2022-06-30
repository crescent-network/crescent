package keeper

// DONTCOVER

// Although written in msg_server_test.go, it is approached at the keeper level rather than at the msgServer level
// so is not included in the coverage.

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the liquidstaking MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	newShares, bTokenMintAmount, err := k.Keeper.LiquidStake(ctx, types.LiquidStakingProxyAcc, msg.GetDelegator(), msg.Amount)
	if err != nil {
		return nil, err
	}

	liquidBondDenom := k.LiquidBondDenom(ctx)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgLiquidStake,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyNewShares, newShares.String()),
			sdk.NewAttribute(types.AttributeKeyBTokenMintedAmount, sdk.Coin{Denom: liquidBondDenom, Amount: bTokenMintAmount}.String()),
		),
	})
	return &types.MsgLiquidStakeResponse{}, nil
}

func (k msgServer) LiquidUnstake(goCtx context.Context, msg *types.MsgLiquidUnstake) (*types.MsgLiquidUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	completionTime, unbondingAmount, _, unbondedAmount, err := k.Keeper.LiquidUnstake(ctx, types.LiquidStakingProxyAcc, msg.GetDelegator(), msg.Amount)
	if err != nil {
		return nil, err
	}

	bondDenom := k.stakingKeeper.BondDenom(ctx)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgLiquidUnstake,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyUnbondingAmount, sdk.Coin{Denom: bondDenom, Amount: unbondingAmount}.String()),
			sdk.NewAttribute(types.AttributeKeyUnbondedAmount, sdk.Coin{Denom: bondDenom, Amount: unbondedAmount}.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
	})
	return &types.MsgLiquidUnstakeResponse{
		CompletionTime: completionTime,
	}, nil
}
