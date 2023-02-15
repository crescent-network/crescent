package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

func (k Keeper) LimitOrder(ctx sdk.Context, orderer sdk.AccAddress, poolId uint64, direction types.OrderDirection,
	offerCoin sdk.Coin, price sdk.Dec) error {

	// TODO:

	return nil
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
