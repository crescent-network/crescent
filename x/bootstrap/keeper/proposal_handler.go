package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

// HandleBootstrapProposal is a handler for executing a market maker proposal.
func HandleBootstrapProposal(ctx sdk.Context, k Keeper, p *types.BootstrapProposal) error {
	// keeper level validation logic
	if k.bankKeeper.GetSupply(ctx, p.QuoteCoinDenom).Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "denom %s has no supply", p.QuoteCoinDenom)
	}

	_, found := k.liquidityKeeper.GetPair(ctx, p.PairId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", p.PairId)
	}

	_, found = k.liquidityKeeper.GetPool(ctx, p.PoolId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool %d not found", p.PoolId)
	}

	proposer := p.GetProposer()
	// TODO: TBD along vesting method
	// check proposer address is not vesting account
	bacc := k.accountKeeper.GetAccount(ctx, proposer)
	_, ok := bacc.(exported.VestingAccount)
	if ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "proposer %s must not vesting account", p.ProposerAddress)
	}

	bp := types.NewBootstrapPool(k.GetLastBootstrapPoolId(ctx)+1, p.BaseCoinDenom, p.QuoteCoinDenom, p.MinPrice, p.MaxPrice, proposer)

	// TODO: make stage schedules StartTime, NumOfStages, StageDuration

	// escrow OfferCoins
	err := k.bankKeeper.SendCoins(ctx, proposer, bp.GetEscrowAddress(), p.OfferCoins)
	if err != nil {
		return err
	}

	// TODO: Set BootstrapPool

	// TODO: set initial orders, Set, store

	// TODO: collecting creation fee
	//creationFees := sdk.NewCoins()

	return nil
}

//// IncludeBootstraps is a handler for include applied and not eligible market makers.
//func (k Keeper) IncludeBootstraps(ctx sdk.Context, proposals []types.BootstrapHandle) error {
//	for _, p := range proposals {
//		mmAddr, err := sdk.AccAddressFromBech32(p.Address)
//		if err != nil {
//			return err
//		}
//		mm, found := k.GetBootstrap(ctx, mmAddr, p.PairId)
//		if !found {
//			return sdkerrors.Wrapf(types.ErrNotExistBootstrap, "%s is not a applied market maker", p.Address)
//		}
//		// fail when already eligible market maker
//		if mm.Eligible {
//			return sdkerrors.Wrapf(types.ErrInvalidInclusion, "%s is already eligible market maker", p.Address)
//		}
//		mm.Eligible = true
//		k.SetBootstrap(ctx, mm)
//
//		// refund deposit amount
//		err = k.RefundDeposit(ctx, mmAddr, p.PairId)
//		if err != nil {
//			return err
//		}
//
//		ctx.EventManager().EmitEvents(sdk.Events{
//			sdk.NewEvent(
//				types.EventTypeIncludeBootstrap,
//				sdk.NewAttribute(types.AttributeKeyAddress, p.Address),
//				sdk.NewAttribute(types.AttributeKeyPairId, fmt.Sprintf("%d", p.PairId)),
//			),
//		})
//	}
//	return nil
//}
//
//// ExcludeBootstraps is a handler for exclude eligible market makers.
//func (k Keeper) ExcludeBootstraps(ctx sdk.Context, proposals []types.BootstrapHandle) error {
//	for _, p := range proposals {
//		mmAddr, err := sdk.AccAddressFromBech32(p.Address)
//		if err != nil {
//			return err
//		}
//		mm, found := k.GetBootstrap(ctx, mmAddr, p.PairId)
//		if !found {
//			return sdkerrors.Wrapf(types.ErrNotExistBootstrap, "%s is not market maker", p.Address)
//		}
//
//		if !mm.Eligible {
//			return sdkerrors.Wrapf(types.ErrInvalidExclusion, "%s is not eligible market maker", p.Address)
//		}
//
//		k.DeleteBootstrap(ctx, mmAddr, p.PairId)
//
//		ctx.EventManager().EmitEvents(sdk.Events{
//			sdk.NewEvent(
//				types.EventTypeExcludeBootstrap,
//				sdk.NewAttribute(types.AttributeKeyAddress, p.Address),
//				sdk.NewAttribute(types.AttributeKeyPairId, fmt.Sprintf("%d", p.PairId)),
//			),
//		})
//	}
//	return nil
//}
//
//// RejectBootstraps is a handler for reject applied and not eligible market makers.
//func (k Keeper) RejectBootstraps(ctx sdk.Context, proposals []types.BootstrapHandle) error {
//	for _, p := range proposals {
//		mmAddr, err := sdk.AccAddressFromBech32(p.Address)
//		if err != nil {
//			return err
//		}
//
//		mm, found := k.GetBootstrap(ctx, mmAddr, p.PairId)
//		if !found {
//			return sdkerrors.Wrapf(types.ErrNotExistBootstrap, "%s is not market maker", p.Address)
//		}
//
//		if mm.Eligible {
//			return sdkerrors.Wrapf(types.ErrInvalidRejection, "%s is eligible market maker", p.Address)
//		}
//
//		k.DeleteBootstrap(ctx, mmAddr, p.PairId)
//
//		// refund deposit amount
//		err = k.RefundDeposit(ctx, mmAddr, p.PairId)
//		if err != nil {
//			return err
//		}
//
//		ctx.EventManager().EmitEvents(sdk.Events{
//			sdk.NewEvent(
//				types.EventTypeRejectBootstrap,
//				sdk.NewAttribute(types.AttributeKeyAddress, p.Address),
//				sdk.NewAttribute(types.AttributeKeyPairId, fmt.Sprintf("%d", p.PairId)),
//			),
//		})
//	}
//	return nil
//}
//
//// DistributeBootstrapIncentives is a handler for distribute incentives to eligible market makers.
//func (k Keeper) DistributeBootstrapIncentives(ctx sdk.Context, proposals []types.IncentiveDistribution) error {
//	params := k.GetParams(ctx)
//	totalIncentives := sdk.Coins{}
//	for _, p := range proposals {
//		totalIncentives = totalIncentives.Add(p.Amount...)
//
//		mm, found := k.GetBootstrap(ctx, p.GetAccAddress(), p.PairId)
//		if !found {
//			return types.ErrNotExistBootstrap
//		}
//		if !mm.Eligible {
//			return types.ErrNotEligibleBootstrap
//		}
//	}
//
//	budgetAcc := params.IncentiveBudgetAcc()
//	err := k.bankKeeper.SendCoins(ctx, budgetAcc, types.ClaimableIncentiveReserveAcc, totalIncentives)
//	if err != nil {
//		return err
//	}
//
//	for _, p := range proposals {
//		incentive, found := k.GetIncentive(ctx, p.GetAccAddress())
//		if !found {
//			incentive.Claimable = sdk.Coins{}
//		}
//		k.SetIncentive(ctx, types.Incentive{
//			Address:   p.Address,
//			Claimable: incentive.Claimable.Add(p.Amount...),
//		})
//	}
//
//	ctx.EventManager().EmitEvents(sdk.Events{
//		sdk.NewEvent(
//			types.EventTypeDistributeIncentives,
//			sdk.NewAttribute(types.AttributeKeyBudgetAddress, budgetAcc.String()),
//			sdk.NewAttribute(types.AttributeKeyTotalIncentives, totalIncentives.String()),
//		),
//	})
//	return nil
//}
//
//// RefundDeposit is a handler for refund deposit amount and delete deposit object.
//func (k Keeper) RefundDeposit(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64) error {
//	deposit, found := k.GetDeposit(ctx, mmAddr, pairId)
//	if !found {
//		return types.ErrInvalidDeposit
//	}
//	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mmAddr, deposit.Amount)
//	if err != nil {
//		return err
//	}
//	k.DeleteDeposit(ctx, mmAddr, pairId)
//	return nil
//}
