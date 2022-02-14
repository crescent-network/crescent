package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

func (k Keeper) Claim(ctx sdk.Context, msg *types.MsgClaim) (types.ClaimRecord, error) {
	record, found := k.GetClaimRecord(ctx, msg.GetRecipient())
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "claim record not found")
	}

	// TODO: params.ExpiredDuration to block claim?

	notClaimedActions := int64(0)
	for _, action := range []bool{
		record.DepositActionClaimed,
		record.SwapActionClaimed,
		record.FarmingActionClaimed,
	} {
		if !action {
			notClaimedActions++
		}
	}

	// The recipient completed all the actions
	if notClaimedActions == 0 {
		// TODO: emit an event?
		return types.ClaimRecord{}, nil // return nil due to better client handling
	}

	skip := true
	switch msg.ActionType {
	case types.ActionTypeDeposit:
		if !record.DepositActionClaimed {
			record.DepositActionClaimed = true
			skip = false
		}
	case types.ActionTypeSwap:
		if !record.SwapActionClaimed {
			record.SwapActionClaimed = true
			skip = false
		}
	case types.ActionTypeFarming:
		if !record.FarmingActionClaimed {
			record.FarmingActionClaimed = true
			skip = false
		}
	}
	if skip {
		return types.ClaimRecord{}, nil // return nil due to better client handling
	}

	divisor := sdk.NewDec(notClaimedActions)
	amt, _ := sdk.NewDecCoinsFromCoins(record.ClaimableCoins...).QuoDecTruncate(divisor).TruncateDecimal()
	record.ClaimableCoins = record.ClaimableCoins.Sub(amt)
	k.SetClaimRecord(ctx, record)

	// Send claimable amounts to the recipient
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, msg.GetRecipient(), amt); err != nil {
		return types.ClaimRecord{}, sdkerrors.Wrap(err, "failed to send coins from module account to the recipient")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyRecipient, record.Address),
			sdk.NewAttribute(types.AttributeKeyInitialClaimableCoins, record.InitialClaimableCoins.String()),
			sdk.NewAttribute(types.AttributeKeyClaimableCoins, record.ClaimableCoins.String()),
			sdk.NewAttribute(types.AttributeKeyDepositActionClaimed, fmt.Sprint(record.DepositActionClaimed)),
			sdk.NewAttribute(types.AttributeKeySwapActionClaimed, fmt.Sprint(record.SwapActionClaimed)),
			sdk.NewAttribute(types.AttributeKeyFarmingActionClaimed, fmt.Sprint(record.FarmingActionClaimed)),
		),
	})

	return record, nil
}
