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

	record, found := k.GetClaimRecordByRecipient(ctx, airdrop.Id, msg.GetRecipient())
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "claim record not found")
	}

	endTime := k.GetEndTime(ctx, airdrop.Id)
	if !endTime.After(ctx.BlockTime()) {
		return types.ClaimRecord{}, types.ErrTerminatedAirdrop
	}

	if record.ClaimedConditions[msg.ConditionType] {
		return types.ClaimRecord{}, types.ErrAlreadyClaimed
	}

	//
	// TODO: sanity check whether or not if the receipient has executed deposit, swap, and farming stake
	//

	unclaimedNum := int64(0)
	for _, claimed := range record.ClaimedConditions {
		if !claimed {
			unclaimedNum++
		}
	}

	claimableCoins := record.GetClaimableCoinsForCondition(unclaimedNum)

	if err := k.bankKeeper.SendCoins(ctx, airdrop.GetSourceAddress(), record.GetRecipient(), claimableCoins); err != nil {
		return types.ClaimRecord{}, sdkerrors.Wrap(err, "failed to transfer coins to the recipient")
	}

	record.ClaimableCoins = record.ClaimableCoins.Sub(claimableCoins)
	record.ClaimedConditions[msg.ConditionType] = true
	k.SetClaimRecord(ctx, record)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyAirdropId, fmt.Sprint(record.AirdropId)),
			sdk.NewAttribute(types.AttributeKeyRecipient, record.Recipient),
			sdk.NewAttribute(types.AttributeKeyInitialClaimableCoins, record.InitialClaimableCoins.String()),
			sdk.NewAttribute(types.AttributeKeyClaimableCoins, record.ClaimableCoins.String()),
			sdk.NewAttribute(types.AttributeKeyConditionType, msg.ConditionType.String()),
		),
	})

	return record, nil
}

// TerminateAirdrop terminates the airdrop and transfer the remaining coins to the termination address.
func (k Keeper) TerminateAirdrop(ctx sdk.Context, airdrop types.Airdrop) error {
	amt := k.bankKeeper.GetAllBalances(ctx, airdrop.GetSourceAddress())
	if !amt.IsZero() {
		if err := k.bankKeeper.SendCoins(ctx, airdrop.GetSourceAddress(), airdrop.GetTerminationAddress(), amt); err != nil {
			return sdkerrors.Wrap(err, "failed to transfer the remaining coins to the termination address")
		}
	}
	return nil
}
