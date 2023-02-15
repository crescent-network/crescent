package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	utils "github.com/crescent-network/crescent/v4/types"
	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

// HandleBootstrapProposal is a handler for executing a market maker proposal.
func HandleBootstrapProposal(ctx sdk.Context, k Keeper, p *types.BootstrapProposal) error {
	// keeper level validation logic

	// TODO: supply checking could be skipped
	if k.bankKeeper.GetSupply(ctx, p.QuoteCoinDenom).Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "denom %s has no supply", p.QuoteCoinDenom)
	}
	if k.bankKeeper.GetSupply(ctx, p.BaseCoinDenom).Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "denom %s has no supply", p.BaseCoinDenom)
	}

	pair, found := k.liquidityKeeper.GetPair(ctx, p.PairId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", p.PairId)
	}

	// TODO: need to checking spec
	if pair.BaseCoinDenom != p.BaseCoinDenom || pair.QuoteCoinDenom != p.QuoteCoinDenom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid pair, offer coin or quote coin miss matched")
	}

	_, found = k.liquidityKeeper.GetPool(ctx, p.PoolId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool %d not found", p.PoolId)
	}

	params := k.GetParams(ctx)
	// check is the quote denom in whitelist
	if !utils.Contains(params.QuoteCoinWhitelist, p.QuoteCoinDenom) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quote denom %s is not whitelisted", p.QuoteCoinDenom)
	}

	// TODO: TBD along vesting method
	// check proposer address is not vesting account
	proposer := p.GetProposer()
	bacc := k.accountKeeper.GetAccount(ctx, proposer)
	_, ok := bacc.(exported.VestingAccount)
	if ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "proposer %s must not vesting account", p.ProposerAddress)
	}

	bp := types.NewBootstrapPool(k.GetLastBootstrapPoolId(ctx)+1, p.BaseCoinDenom, p.QuoteCoinDenom, p.MinPrice, p.MaxPrice, proposer)

	// TODO: make stage schedules StartTime, NumOfStages, StageDuration

	// escrow offer coins
	err := k.bankKeeper.SendCoins(ctx, proposer, bp.GetEscrowAddress(), p.OfferCoins)
	if err != nil {
		return err
	}

	//  collecting creation fee
	creationFee := utils.CoinsMul(p.OfferCoins, params.CreationFeeRate)
	// TODO: global fee collector or pool fee collector?
	err = k.bankKeeper.SendCoins(ctx, proposer, bp.GetFeeCollector(), creationFee)
	if err != nil {
		return err
	}

	// Set bootstrap pool
	k.SetBootstrapPool(ctx, bp)
	k.SetLastBootstrapPoolId(ctx, bp.Id)
	// TODO: set

	// TODO: set initial orders, Set, store

	// TODO: event emit
	//		ctx.EventManager().EmitEvents(sdk.Events{
	//			sdk.NewEvent(
	//				types.EventTypeCreateBootstrapPool,
	//				sdk.NewAttribute(types.AttributeKeyPoolId, fmt.Sprintf("%d", bp.PoolId)),
	//			),
	//		})

	return nil
}
