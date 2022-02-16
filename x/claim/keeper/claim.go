package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

func (k Keeper) Claim(ctx sdk.Context, msg *types.MsgClaim) (types.ClaimRecord, error) {
	airdrop, found := k.GetAirdrop(ctx, msg.AirdropId)
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "airdrop not found")
	}

	record, found := k.GetClaimRecordByRecipient(ctx, airdrop.AirdropId, msg.GetRecipient())
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "claim record not found")
	}

	endTime := k.GetEndTime(ctx, airdrop.AirdropId)
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
	}

	// The recipient completed all the actions and it returns nil on purpose
	// for better client handling; it prevents from multiple txs getting failed
	if unclaimedActions == 0 {
		// TODO: consider emitting events since it returns nil?
		return types.ClaimRecord{}, nil
	}

	// The recipient already completed the action and it returns nil on purpose
	// for better client handling; it prevents from multiple txs getting failed
	claimed, found := actionsMap[msg.ActionType]
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "action type not found")
	}
	if claimed {
		// TODO: consider emitting events since it returns nil?
		return types.ClaimRecord{}, nil
	}

	// Use divisor to send a proportional amount of the claimable amount to the recipient
	// When unclaimedActions is 1, send all the remaining amount to the recipient
	switch unclaimedActions {
	case 1:
		amt := record.ClaimableCoins
		record.ClaimableCoins = record.ClaimableCoins.Sub(amt)

		if err := k.bankKeeper.SendCoins(ctx, airdrop.GetSourceAddress(), record.GetRecipient(), amt); err != nil {
			return types.ClaimRecord{}, sdkerrors.Wrap(err, "failed to send coins to the recipient")
		}

	default:
		divisor := sdk.NewDec(unclaimedActions)
		amt, _ := sdk.NewDecCoinsFromCoins(record.ClaimableCoins...).QuoDecTruncate(divisor).TruncateDecimal()
		record.ClaimableCoins = record.ClaimableCoins.Sub(amt)

		if err := k.bankKeeper.SendCoins(ctx, airdrop.GetSourceAddress(), record.GetRecipient(), amt); err != nil {
			return types.ClaimRecord{}, sdkerrors.Wrap(err, "failed to transfer coins to the recipient")
		}
	}

	k.SetClaimRecord(ctx, record)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyAirdropId, fmt.Sprint(record.AirdropId)),
			sdk.NewAttribute(types.AttributeKeyRecipient, record.Recipient),
			sdk.NewAttribute(types.AttributeKeyInitialClaimableCoins, record.InitialClaimableCoins.String()),
			sdk.NewAttribute(types.AttributeKeyClaimableCoins, record.ClaimableCoins.String()),
			sdk.NewAttribute(types.AttributeKeyActionType, msg.ActionType.String()),
		),
	})

	return record, nil
}

// TerminateAirdrop terminates the airdrop and transfer the remaining coins to the termination address.
func (k Keeper) TerminateAirdrop(ctx sdk.Context, airdrop types.Airdrop) error {
	amt := k.bankKeeper.GetAllBalances(ctx, airdrop.GetSourceAddress())

	if err := k.bankKeeper.SendCoins(ctx, airdrop.GetSourceAddress(), airdrop.GetTerminationAddress(), amt); err != nil {
		return sdkerrors.Wrap(err, "failed to transfer the remaining coins to the termination address")
	}
	return nil
}
