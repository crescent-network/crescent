package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

func (k Keeper) LimitOrder(ctx sdk.Context, msg *types.MsgLimitOrder) (types.Order, error) {
	// TODO: keeper level validate
	tickAdjustedOfferCoin, price, err := k.ValidateMsgLimitOrder(ctx, msg)
	if err != nil {
		return types.Order{}, err
	}

	refundedCoin := msg.OfferCoin.Sub(tickAdjustedOfferCoin)
	pool, _ := k.GetBootstrapPool(ctx, msg.BootstrapPoolId)
	if err := k.bankKeeper.SendCoins(ctx, msg.GetOrderer(), pool.GetEscrowAddress(), sdk.NewCoins(tickAdjustedOfferCoin)); err != nil {
		return types.Order{}, err
	}

	orderId := k.getNextOrderIdWithUpdate(ctx, pool.Id)
	order := types.NewOrderForLimitOrder(msg, orderId, pool.Id, tickAdjustedOfferCoin, price, ctx.BlockHeight())
	k.SetOrder(ctx, order)
	k.SetOrderIndex(ctx, order)

	ctx.GasMeter().ConsumeGas(k.GetOrderExtraGas(ctx), "OrderExtraGas")

	// TODO: event emit
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLimitOrder,
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(msg.BootstrapPoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyOrderDirection, msg.Direction.String()),
			sdk.NewAttribute(types.AttributeKeyOfferCoin, tickAdjustedOfferCoin.String()),
			//sdk.NewAttribute(types.AttributeKeyDemandCoinDenom, msg.DemandCoinDenom),
			sdk.NewAttribute(types.AttributeKeyPrice, price.String()),
			//sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyOrderId, strconv.FormatUint(order.Id, 10)),
			//sdk.NewAttribute(types.AttributeKeyStageId, strconv.FormatUint(order.BatchId, 10)),
			sdk.NewAttribute(types.AttributeKeyRefundedCoins, refundedCoin.String()),
		),
	})

	return order, nil
}

// ValidateMsgLimitOrder validates types.MsgLimitOrder with state and returns
// calculated offer coin and price that is fit into ticks.
func (k Keeper) ValidateMsgLimitOrder(ctx sdk.Context, msg *types.MsgLimitOrder) (offerCoin sdk.Coin, price sdk.Dec, err error) {
	spendable := k.bankKeeper.SpendableCoins(ctx, msg.GetOrderer())
	if spendableAmt := spendable.AmountOf(msg.OfferCoin.Denom); spendableAmt.LT(msg.OfferCoin.Amount) {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds, "%s is smaller than %s",
			sdk.NewCoin(msg.OfferCoin.Denom, spendableAmt), msg.OfferCoin)
	}

	//tickPrec := k.GetTickPrecision(ctx)

	pool, found := k.GetBootstrapPool(ctx, msg.BootstrapPoolId)
	if !found {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "bootstrap pool %d not found", msg.BootstrapPoolId)
	}

	// TODO: checking IsActive
	if !pool.IsActive() {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrInactivePool, "bootstrap pool %d inactive", pool.Id)
	}

	//switch msg.Direction {
	//case types.OrderDirectionBuy:
	//	if msg.OfferCoin.Denom != pool.QuoteCoinDenom || msg.DemandCoinDenom != pool.BaseCoinDenom {
	//		return sdk.Coin{}, sdk.Dec{},
	//			sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
	//				msg.DemandCoinDenom, msg.OfferCoin.Denom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
	//	}
	//	price = amm.PriceToDownTick(msg.Price, int(tickPrec))
	//	offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, amm.OfferCoinAmount(amm.Buy, price, msg.Amount))
	//	if msg.OfferCoin.IsLT(offerCoin) {
	//		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
	//			types.ErrInsufficientOfferCoin, "%s is smaller than %s", msg.OfferCoin, offerCoin)
	//	}
	//case types.OrderDirectionSell:
	//	if msg.OfferCoin.Denom != pair.BaseCoinDenom || msg.DemandCoinDenom != pair.QuoteCoinDenom {
	//		return sdk.Coin{}, sdk.Dec{},
	//			sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
	//				msg.OfferCoin.Denom, msg.DemandCoinDenom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
	//	}
	//	price = amm.PriceToUpTick(msg.Price, int(tickPrec))
	//	offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount)
	//	if msg.OfferCoin.Amount.LT(msg.Amount) {
	//		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
	//			types.ErrInsufficientOfferCoin, "%s is smaller than %s", msg.OfferCoin, sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount))
	//	}
	//}
	//if types.IsTooSmallOrderAmount(msg.Amount, price) {
	//	return sdk.Coin{}, sdk.Dec{}, types.ErrTooSmallOrder
	//}

	// TODO: fix tick
	//return offerCoin, price, nil
	return msg.OfferCoin, msg.Price, nil
}

func (k Keeper) ModifyOrder(ctx sdk.Context, msg *types.MsgModifyOrder) error {
	// TODO: keeper level validate
	//tickAdjustedOfferCoin, price, err := k.ValidateMsgModifyOrder(ctx, msg)
	//if err != nil {
	//	return types.Order{}, err
	//}

	// TODO:

	return nil
}

// ValidateMsgModifyOrder validates types.MsgModifyOrder with state and returns
// calculated offer coin and price that is fit into ticks.
func (k Keeper) ValidateMsgModifyOrder(ctx sdk.Context, msg *types.MsgModifyOrder) (offerCoin sdk.Coin, price sdk.Dec, err error) {
	//spendable := k.bankKeeper.SpendableCoins(ctx, msg.GetOrderer())
	//if spendableAmt := spendable.AmountOf(msg.OfferCoin.Denom); spendableAmt.LT(msg.OfferCoin.Amount) {
	//	return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
	//		sdkerrors.ErrInsufficientFunds, "%s is smaller than %s",
	//		sdk.NewCoin(msg.OfferCoin.Denom, spendableAmt), msg.OfferCoin)
	//}

	//tickPrec := k.GetTickPrecision(ctx)

	pool, found := k.GetBootstrapPool(ctx, msg.BootstrapPoolId)
	if !found {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "bootstrap pool %d not found", msg.BootstrapPoolId)
	}

	// TODO: checking IsActive
	if !pool.IsActive() {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrInactivePool, "bootstrap pool %d inactive", pool.Id)
	}

	order, found := k.GetOrder(ctx, pool.Id, msg.OrderId)
	if !found {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "order %d not found", msg.OrderId)
	}

	if order.Orderer != msg.Orderer {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "orderer %s not matched", order.Orderer)
	}

	if msg.BootstrapPoolId != order.BootstrapPoolId {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "bootstrap pool not matched")
	}

	// The denom of offer coin is not the same as the bootstrap coin denom
	if offerCoin.Denom != order.OfferCoin.Denom {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "offer coin denom not matched")
	}

	// The price is less than the original amount of price
	if price.LT(order.Price) {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price should be higher than original order")
	}

	// The amount of offer coin is less than the original amount of offer coin
	if offerCoin.Amount.LT(order.OfferCoin.Amount) {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "offer coin amount should be more than original order")
	}

	//- Both offer coin and price are the same as those of the original order
	if offerCoin.Amount.Equal(order.OfferCoin.Amount) && price.Equal(order.Price) {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "offer coin amount and price is same with original order")
	}

	// TODO: add demand denom, amount
	//- Direction is not the same as the original one

	//switch msg.Direction {
	//case types.OrderDirectionBuy:
	//	if msg.OfferCoin.Denom != pool.QuoteCoinDenom || msg.DemandCoinDenom != pool.BaseCoinDenom {
	//		return sdk.Coin{}, sdk.Dec{},
	//			sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
	//				msg.DemandCoinDenom, msg.OfferCoin.Denom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
	//	}
	//	price = amm.PriceToDownTick(msg.Price, int(tickPrec))
	//	offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, amm.OfferCoinAmount(amm.Buy, price, msg.Amount))
	//	if msg.OfferCoin.IsLT(offerCoin) {
	//		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
	//			types.ErrInsufficientOfferCoin, "%s is smaller than %s", msg.OfferCoin, offerCoin)
	//	}
	//case types.OrderDirectionSell:
	//	if msg.OfferCoin.Denom != pair.BaseCoinDenom || msg.DemandCoinDenom != pair.QuoteCoinDenom {
	//		return sdk.Coin{}, sdk.Dec{},
	//			sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
	//				msg.OfferCoin.Denom, msg.DemandCoinDenom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
	//	}
	//	price = amm.PriceToUpTick(msg.Price, int(tickPrec))
	//	offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount)
	//	if msg.OfferCoin.Amount.LT(msg.Amount) {
	//		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
	//			types.ErrInsufficientOfferCoin, "%s is smaller than %s", msg.OfferCoin, sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount))
	//	}
	//}
	//if types.IsTooSmallOrderAmount(msg.Amount, price) {
	//	return sdk.Coin{}, sdk.Dec{}, types.ErrTooSmallOrder
	//}

	// TODO: fix tick
	//return offerCoin, price, nil
	return msg.OfferCoin, msg.Price, nil
}

func (k Keeper) Vesting(ctx sdk.Context, returnAddr sdk.AccAddress, originalVesting sdk.Coins, startTime int64, periods vestingtypes.Periods) {
	//var account authtypes.AccountI
	bacc := k.accountKeeper.GetAccount(ctx, returnAddr)
	fmt.Println(bacc.GetPubKey(), bacc.GetSequence(), bacc.GetAccountNumber(), bacc.GetAddress())

	// TODO: send

	_, ok := bacc.(exported.VestingAccount)
	if ok {
		panic("already vested")
	}

	acc := vestingtypes.NewPeriodicVestingAccount(bacc.(*authtypes.BaseAccount), originalVesting, startTime, periods)
	k.accountKeeper.SetAccount(ctx, acc)
}

//func (k Keeper) ApplyBootstrap(ctx sdk.Context, mmAddr sdk.AccAddress, pairIds []uint64) error {
//	params := k.GetParams(ctx)
//	incentivePairsMap := params.IncentivePairsMap()
//
//	totalDepositAmt := sdk.Coins{}
//	for _, pairId := range pairIds {
//		// Fail if the same market maker already exists
//		_, found := k.GetBootstrapPool(ctx, mmAddr, pairId)
//		if found {
//			return types.ErrAlreadyExistBootstrap
//		}
//		totalDepositAmt = totalDepositAmt.Add(params.DepositAmount...)
//
//		// fail for pairs that are not registered as incentive pairs on params
//		if _, ok := incentivePairsMap[pairId]; !ok {
//			return types.ErrUnregisteredPairId
//		}
//	}
//
//	// total deposit amount = deposit amount * number of pairs
//	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, mmAddr, types.ModuleName, totalDepositAmt)
//	if err != nil {
//		return err
//	}
//
//	// create market maker, deposit object for each pair
//	for _, pairId := range pairIds {
//		k.SetDeposit(ctx, mmAddr, pairId, params.DepositAmount)
//		k.SetBootstrapPool(ctx, types.Bootstrap{
//			Address:  mmAddr.String(),
//			PairId:   pairId,
//			Eligible: false,
//		})
//	}
//
//	ctx.EventManager().EmitEvents(sdk.Events{
//		sdk.NewEvent(
//			types.EventTypeApplyBootstrap,
//			sdk.NewAttribute(types.AttributeKeyAddress, mmAddr.String()),
//			sdk.NewAttribute(types.AttributeKeyPairIds, strings.Trim(strings.Replace(fmt.Sprint(pairIds), " ", ",", -1), "[]")),
//		),
//		sdk.NewEvent(
//			sdk.EventTypeMessage,
//			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
//			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeApplyBootstrap),
//			sdk.NewAttribute(sdk.AttributeKeySender, mmAddr.String()),
//		),
//	})
//
//	return nil
//}
//
//func (k Keeper) ClaimIncentives(ctx sdk.Context, mmAddr sdk.AccAddress) error {
//	incentive, found := k.GetIncentive(ctx, mmAddr)
//	if !found {
//		return types.ErrEmptyClaimableIncentive
//	}
//
//	if err := k.bankKeeper.SendCoins(ctx, types.ClaimableIncentiveReserveAcc, mmAddr, incentive.Claimable); err != nil {
//		return err
//	}
//
//	k.DeleteIncentive(ctx, mmAddr)
//
//	ctx.EventManager().EmitEvents(sdk.Events{
//		sdk.NewEvent(
//			types.EventTypeClaimIncentives,
//			sdk.NewAttribute(types.AttributeKeyAddress, mmAddr.String()),
//			sdk.NewAttribute(sdk.AttributeKeyAmount, incentive.Claimable.String()),
//		),
//		sdk.NewEvent(
//			sdk.EventTypeMessage,
//			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
//			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeClaimIncentives),
//			sdk.NewAttribute(sdk.AttributeKeySender, mmAddr.String()),
//		),
//	})
//	return nil
//}
//
//func (k Keeper) ValidateDepositReservedAmount(ctx sdk.Context) error {
//	mmCount := 0
//	depositCount := 0
//	var totalAmt sdk.Coins
//	k.IterateBootstraps(ctx, func(mm types.Bootstrap) (stop bool) {
//		if !mm.Eligible {
//			mmCount += 1
//		}
//		return false
//	})
//	k.IterateDeposits(ctx, func(id types.Deposit) (stop bool) {
//		depositCount += 1
//		totalAmt = totalAmt.Add(id.Amount...)
//		return false
//	})
//	if mmCount != depositCount {
//		return fmt.Errorf("market maker number differs from the actual value; have %d, want %d", mmCount, depositCount)
//	}
//
//	if !totalAmt.Empty() {
//		reserveBalance := k.bankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
//		if !reserveBalance.IsAllGTE(totalAmt) {
//			return fmt.Errorf("DepositReserveAcc differs from the actual value; have %s, want %s", reserveBalance, totalAmt)
//		}
//	}
//	return nil
//}
//
//func (k Keeper) ValidateIncentiveReservedAmount(ctx sdk.Context, incentives []types.Incentive) error {
//	var totalClaimable sdk.Coins
//	for _, record := range incentives {
//		totalClaimable = totalClaimable.Add(record.Claimable...)
//	}
//	if !totalClaimable.Empty() {
//		reserveBalance := k.bankKeeper.GetAllBalances(ctx, types.ClaimableIncentiveReserveAcc)
//		if !reserveBalance.IsAllGTE(totalClaimable) {
//			return fmt.Errorf("ClaimableIncentiveReserveAcc differs from the actual value; have %s, want %s", reserveBalance, totalClaimable)
//		}
//	}
//	return nil
//}
