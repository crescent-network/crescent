package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

func (k Keeper) Claim(ctx sdk.Context, msg *types.MsgClaim) (types.ClaimRecord, error) {
	endTime := k.GetEndTime(ctx, msg.AirdropId)
	if !endTime.After(ctx.BlockTime()) {
		return types.ClaimRecord{}, types.ErrTerminatedAirdrop
	}

	airdrop, found := k.GetAirdrop(ctx, msg.AirdropId)
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "airdrop not found")
	}

	record, found := k.GetClaimRecordByRecipient(ctx, airdrop.Id, msg.GetRecipient())
	if !found {
		return types.ClaimRecord{}, sdkerrors.Wrap(sdkerrors.ErrNotFound, "claim record not found")
	}

	for _, c := range record.ClaimedConditions {
		if c == msg.ConditionType {
			return types.ClaimRecord{}, types.ErrAlreadyClaimed
		}
	}

	// Vadliate whether or not the recipient has executed the condition
	if err := k.ValidateCondition(ctx, record.GetRecipient(), msg.ConditionType); err != nil {
		return types.ClaimRecord{}, err
	}

	claimableCoins := record.GetClaimableCoinsForCondition(airdrop.Conditions)

	if err := k.bankKeeper.SendCoins(ctx, airdrop.GetSourceAddress(), record.GetRecipient(), claimableCoins); err != nil {
		return types.ClaimRecord{}, sdkerrors.Wrap(err, "failed to transfer coins to the recipient")
	}

	record.ClaimableCoins = record.ClaimableCoins.Sub(claimableCoins)
	record.ClaimedConditions = append(record.ClaimedConditions, msg.ConditionType)
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

// ValidateCondition validates if the recipient has executed the condition.
func (k Keeper) ValidateCondition(ctx sdk.Context, recipient sdk.AccAddress, ct types.ConditionType) error {
	skip := false

	switch ct {
	case types.ConditionTypeDeposit:
		if len(k.liquidityKeeper.GetDepositRequestsByDepositor(ctx, recipient)) != 0 {
			skip = true
		}

	case types.ConditionTypeSwap:
		if len(k.liquidityKeeper.GetOrdersByOrderer(ctx, recipient)) != 0 {
			skip = true
		}

	case types.ConditionTypeFarming:
		queuedCoins := k.farmingKeeper.GetAllQueuedCoinsByFarmer(ctx, recipient)
		stakedCoins := k.farmingKeeper.GetAllStakedCoinsByFarmer(ctx, recipient)
		if !queuedCoins.IsZero() || !stakedCoins.IsZero() {
			skip = true
		}
	}

	if !skip {
		return types.ErrConditionRequired
	}

	return nil
}

// TerminateAirdrop terminates the airdrop and transfer the remaining coins to the community pool.
func (k Keeper) TerminateAirdrop(ctx sdk.Context, airdrop types.Airdrop) error {
	amt := k.bankKeeper.SpendableCoins(ctx, airdrop.GetSourceAddress())
	if !amt.IsZero() {
		if err := k.distrKeeper.FundCommunityPool(ctx, amt, airdrop.GetSourceAddress()); err != nil {
			return sdkerrors.Wrap(err, "failed to transfer the remaining coins to the community pool")
		}
	}
	return nil
}
