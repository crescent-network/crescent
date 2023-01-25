package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO:

// RegisterInvariants registers all bootstrap invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	//ir.RegisterRoute(types.ModuleName, "deposit-reserved-amount",
	//	DepositReservedAmountInvariant(k))
	//ir.RegisterRoute(types.ModuleName, "incentive-reserved-amount",
	//	IncentiveReservedAmountInvariant(k))
	//ir.RegisterRoute(types.ModuleName, "deposit-record",
	//	DepositRecordsInvariant(k))
}

// AllInvariants runs all invariants of the bootstrap module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, inv := range []func(Keeper) sdk.Invariant{
			//DepositReservedAmountInvariant,
			//IncentiveReservedAmountInvariant,
			//DepositRecordsInvariant,
		} {
			res, stop := inv(k)(ctx)
			if stop {
				return res, stop
			}
		}
		return "", false
	}
}

//// DepositReservedAmountInvariant checks that the balance of StakingReserveAcc greater than the amount of staked, Queued coins in all staking objects.
//func DepositReservedAmountInvariant(k Keeper) sdk.Invariant {
//	return func(ctx sdk.Context) (string, bool) {
//		err := k.ValidateDepositReservedAmount(ctx)
//		broken := err != nil
//		return sdk.FormatInvariant(types.ModuleName, "deposit reserved amount",
//			"the balance of DepositReserveAcc less than the amount of deposited in all deposit objects",
//		), broken
//	}
//}
//
//// IncentiveReservedAmountInvariant checks that the balance of StakingReserveAcc greater than the amount of staked, Queued coins in all staking objects.
//func IncentiveReservedAmountInvariant(k Keeper) sdk.Invariant {
//	return func(ctx sdk.Context) (string, bool) {
//		err := k.ValidateIncentiveReservedAmount(ctx, k.GetAllIncentives(ctx))
//		broken := err != nil
//		return sdk.FormatInvariant(types.ModuleName, "claimable incentive reserved amount",
//			"the balance of ClaimableIncentiveReserveAcc less than the amount queued in all incentive objects",
//		), broken
//	}
//}
//
//// DepositRecordsInvariant checks that the invariants for pair of deposit records with not eligible market maker.
//func DepositRecordsInvariant(k Keeper) sdk.Invariant {
//	return func(ctx sdk.Context) (string, bool) {
//		err := types.ValidateDepositRecords(k.GetAllBootstraps(ctx), k.GetAllDepositRecords(ctx))
//		broken := err != nil
//		return sdk.FormatInvariant(types.ModuleName, "deposit record",
//			"the deposit record not matched with not eligible market maker",
//		), broken
//	}
//}
