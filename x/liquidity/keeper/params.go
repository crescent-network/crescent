package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func (k Keeper) GetMaxPriceLimitRatio(ctx sdk.Context) (ratio sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyMaxPriceLimitRatio, &ratio)
	return
}

func (k Keeper) GetTickPrecision(ctx sdk.Context) (tickPrec uint32) {
	k.paramSpace.Get(ctx, types.KeyTickPrecision, &tickPrec)
	return
}

func (k Keeper) GetFeeCollector(ctx sdk.Context) sdk.AccAddress {
	var feeCollectorAddr string
	k.paramSpace.Get(ctx, types.KeyFeeCollectorAddress, &feeCollectorAddr)
	addr, err := sdk.AccAddressFromBech32(feeCollectorAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

func (k Keeper) GetDustCollector(ctx sdk.Context) sdk.AccAddress {
	var dustCollectorAddr string
	k.paramSpace.Get(ctx, types.KeyDustCollectorAddress, &dustCollectorAddr)
	addr, err := sdk.AccAddressFromBech32(dustCollectorAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

func (k Keeper) GetMinInitialPoolCoinSupply(ctx sdk.Context) (i sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyMinInitialPoolCoinSupply, &i)
	return
}

func (k Keeper) GetPoolCreationFee(ctx sdk.Context) (fee sdk.Coins) {
	k.paramSpace.Get(ctx, types.KeyPoolCreationFee, &fee)
	return
}
