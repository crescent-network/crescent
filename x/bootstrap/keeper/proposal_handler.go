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
	if k.bankKeeper.GetSupply(ctx, p.BaseCoinDenom).Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "denom %s has no supply", p.BaseCoinDenom)
	}
	if k.bankKeeper.GetSupply(ctx, p.QuoteCoinDenom).Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "denom %s has no supply", p.QuoteCoinDenom)
	}

	// check the pair exist
	pair, found := k.liquidityKeeper.GetPair(ctx, p.PairId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", p.PairId)
	}

	// the pair's base coin or quote coin must be bootstrap pool's base coin
	if pair.BaseCoinDenom != p.BaseCoinDenom && pair.QuoteCoinDenom != p.BaseCoinDenom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid pair, the pair's base coin or quote coin must be bootstrap pool's base coin")
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

	// TODO: could be deleted
	if ctx.BlockTime().After(p.StartTime) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "start time %s must be after block time", p.StartTime)
	}

	bp := types.NewBootstrapPool(k.GetLastBootstrapPoolId(ctx)+1, p.BaseCoinDenom, p.QuoteCoinDenom, p.PairId, p.MinPrice, p.MaxPrice, p.GetStages(), proposer, params)
	if err := bp.Validate(); err != nil {
		return err
	}

	// escrow offer coins
	err := k.bankKeeper.SendCoins(ctx, proposer, bp.GetEscrowAddress(), p.OfferCoins)
	if err != nil {
		return err
	}

	//  collecting creation fee
	// TODO: only sell coins or all offer coins
	creationFee := utils.CoinsMul(p.OfferCoins, params.CreationFeeRate)
	// TODO: global fee collector or pool fee collector?
	err = k.bankKeeper.SendCoins(ctx, proposer, bp.GetFeeCollector(), creationFee)
	if err != nil {
		return err
	}

	// escrow InitialTradingFeeRate only sell offer coins +@
	// TODO: need to tracking? seperated escrow account?
	initialTradingFee := sdk.Coins{}
	protocolFee := sdk.Coins{}
	for _, io := range p.InitialOrders {
		if io.Direction == types.OrderDirectionSell {
			initialTradingFee = initialTradingFee.Add(utils.CoinMul(io.OfferCoin, params.InitialTradingFeeRate))
			protocolFee = protocolFee.Add(utils.CoinMul(io.OfferCoin, params.ProtocolFeeRate))
		}
	}
	// TODO: escrow to bp.Escrow? or sperated? or feecollecor?
	err = k.bankKeeper.SendCoins(ctx, proposer, bp.GetEscrowAddress(), initialTradingFee)
	if err != nil {
		return err
	}
	err = k.bankKeeper.SendCoins(ctx, proposer, bp.GetEscrowAddress(), protocolFee)
	if err != nil {
		return err
	}

	// TODO: escrow ProtocolFee, to where?

	// Set bootstrap pool
	k.SetBootstrapPool(ctx, bp)
	k.SetLastBootstrapPoolId(ctx, bp.Id)

	// TODO: set initial orders, Set, store

	// TODO: refactor PlaceInitialOrder with cached context
	for _, io := range p.InitialOrders {
		orderId := k.getNextOrderIdWithUpdate(ctx, bp.Id)
		// TODO: add fee field for invariant checking?
		order := types.NewOrderForInitialOrder(io, orderId, bp.Id, ctx.BlockHeight(), bp.ProposerAddress)
		k.SetOrder(ctx, order)
		k.SetOrderIndex(ctx, order)
	}

	// TODO: event emit
	//		ctx.EventManager().EmitEvents(sdk.Events{
	//			sdk.NewEvent(
	//				types.EventTypeCreateBootstrapPool,
	//				sdk.NewAttribute(types.AttributeKeyPoolId, fmt.Sprintf("%d", bp.PoolId)),
	//			),
	//		})

	return nil
}
