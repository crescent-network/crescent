package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) CreateSharedPosition(
	ctx sdk.Context, creatorAddr sdk.AccAddress, poolId uint64,
	lowerPrice, upperPrice sdk.Dec, desiredAmt sdk.Coins) (sharedPosition types.SharedPosition, liquidity sdk.Int, amt sdk.Coins, err error) {
	return
}

func (k Keeper) MintPositionShare(
	ctx sdk.Context, ownerAddr sdk.AccAddress, positionId uint64, desiredAmt sdk.Coins) (liquidity sdk.Int, amt sdk.Coins, err error) {
	return
}

func (k Keeper) BurnPositionShare(
	ctx sdk.Context, ownerAddr sdk.AccAddress, positionId uint64, liquidity sdk.Int) (amt sdk.Coins, err error) {
	return
}
