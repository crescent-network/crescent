package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

// GetNextAirdropIdWithUpdate increments airdrop id by one and set it.
func (k Keeper) GetNextAirdropIdWithUpdate(ctx sdk.Context) uint64 {
	id := k.GetLastAirdropId(ctx) + 1
	k.SetAirdropId(ctx, id)
	return id
}

func (k Keeper) Claim(ctx sdk.Context, msg *types.MsgClaim) (types.ClaimRecord, error) {
	airdrop, found := k.GetAirdrop(ctx, msg.AirdropId)
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "airdrop not found")
	}

	record, found := k.GetClaimRecordByRecipient(ctx, airdrop.AirdropId, msg.GetRecipient())
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "claim record not found")
	}

	endTime, err := k.GetEndTime(ctx, airdrop.AirdropId)
	if err != nil {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "end time not found") // TODO: better way to handle this?
	}

	if !endTime.After(ctx.BlockTime()) {
		return types.ClaimRecord{}, types.ErrTerminatedAirdrop
	}

	// Increment a number of unclaimed actions and add values
	// to actionsMap to sanity check the already completed action
	actionsMap := make(map[types.ActionType]bool, len(record.Actions))
	unclaimedActions := int64(0)
	for i, action := range record.Actions {
		if !action.Claimed {
			unclaimedActions++
		}

		actionsMap[action.ActionType] = action.Claimed

		// Update the claimed status
		if action.ActionType == msg.ActionType {
			if !action.Claimed {
				record.Actions[i].Claimed = true
			}
		}

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeClaim,
				sdk.NewAttribute(types.AttributeKeyActionType, action.ActionType.String()),
				sdk.NewAttribute(types.AttributeKeyClaimed, fmt.Sprint(action.Claimed)),
			),
		})
	}

	// The recipient completed all the actions and it returns nil on purpose
	// for better client handling; it prevents from multiple txs getting failed
	if unclaimedActions == 0 {
		return types.ClaimRecord{}, nil
	}

	// The recipient already completed the action and it returns nil on purpose
	// for better client handling; it prevents from multiple txs getting failed
	claimed, found := actionsMap[msg.ActionType]
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "action type not found")
	}
	if claimed {
		return types.ClaimRecord{}, nil
	}

	switch unclaimedActions {
	case 2, 3:
		divisor := sdk.NewDec(unclaimedActions)
		amt, _ := sdk.NewDecCoinsFromCoins(record.ClaimableCoins...).QuoDecTruncate(divisor).TruncateDecimal()
		record.ClaimableCoins = record.ClaimableCoins.Sub(amt)

		if err := k.bankKeeper.SendCoins(ctx, airdrop.GetSourceAddress(), record.GetRecipient(), amt); err != nil {
			return types.ClaimRecord{}, sdkerrors.Wrap(err, "failed to send coins to the recipient")
		}
	default: // Send all the remaining coins to the recipient
		amt := record.ClaimableCoins
		record.ClaimableCoins = record.ClaimableCoins.Sub(amt)

		if err := k.bankKeeper.SendCoins(ctx, airdrop.GetSourceAddress(), record.GetRecipient(), amt); err != nil {
			return types.ClaimRecord{}, sdkerrors.Wrap(err, "failed to send coins to the recipient")
		}
	}

	k.SetClaimRecord(ctx, record)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyRecipient, record.Recipient),
			sdk.NewAttribute(types.AttributeKeyInitialClaimableCoins, record.InitialClaimableCoins.String()),
			sdk.NewAttribute(types.AttributeKeyClaimableCoins, record.ClaimableCoins.String()),
		),
	})

	return record, nil
}
