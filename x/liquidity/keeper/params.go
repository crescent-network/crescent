package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

// GetBatchSize returns the current batch size parameter.
func (k Keeper) GetBatchSize(ctx sdk.Context) (batchSize uint32) {
	k.paramSpace.Get(ctx, types.KeyBatchSize, &batchSize)
	return
}

// GetTickPrecision returns the current tick precision parameter.
func (k Keeper) GetTickPrecision(ctx sdk.Context) (tickPrec uint32) {
	k.paramSpace.Get(ctx, types.KeyTickPrecision, &tickPrec)
	return
}

// GetFeeCollector returns the current fee collector address parameter.
func (k Keeper) GetFeeCollector(ctx sdk.Context) sdk.AccAddress {
	var feeCollectorAddr string
	k.paramSpace.Get(ctx, types.KeyFeeCollectorAddress, &feeCollectorAddr)
	addr, err := sdk.AccAddressFromBech32(feeCollectorAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

// GetDustCollector returns the current dust collector address parameter.
func (k Keeper) GetDustCollector(ctx sdk.Context) sdk.AccAddress {
	var dustCollectorAddr string
	k.paramSpace.Get(ctx, types.KeyDustCollectorAddress, &dustCollectorAddr)
	addr, err := sdk.AccAddressFromBech32(dustCollectorAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

// GetMinInitialPoolCoinSupply returns the current minimum pool coin supply
// parameter.
func (k Keeper) GetMinInitialPoolCoinSupply(ctx sdk.Context) (i sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyMinInitialPoolCoinSupply, &i)
	return
}

// GetPairCreationFee returns the current pair creation fee parameter.
func (k Keeper) GetPairCreationFee(ctx sdk.Context) (fee sdk.Coins) {
	k.paramSpace.Get(ctx, types.KeyPairCreationFee, &fee)
	return
}

// GetPoolCreationFee returns the current pool creation fee parameter.
func (k Keeper) GetPoolCreationFee(ctx sdk.Context) (fee sdk.Coins) {
	k.paramSpace.Get(ctx, types.KeyPoolCreationFee, &fee)
	return
}

// GetMinInitialDepositAmount returns the current minimum initial deposit
// amount parameter.
func (k Keeper) GetMinInitialDepositAmount(ctx sdk.Context) (amt sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyMinInitialDepositAmount, &amt)
	return
}

// GetMaxPriceLimitRatio returns the current maximum price limit ratio
// parameter.
func (k Keeper) GetMaxPriceLimitRatio(ctx sdk.Context) (ratio sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyMaxPriceLimitRatio, &ratio)
	return
}

// GetMaxOrderLifespan returns the current maximum order lifespan
// parameter.
func (k Keeper) GetMaxOrderLifespan(ctx sdk.Context) (maxLifespan time.Duration) {
	k.paramSpace.Get(ctx, types.KeyMaxOrderLifespan, &maxLifespan)
	return
}

// GetWithdrawFeeRate returns the current withdraw fee rate parameter.
func (k Keeper) GetWithdrawFeeRate(ctx sdk.Context) (feeRate sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyWithdrawFeeRate, &feeRate)
	return
}

// GetDepositExtraGas returns the current deposit extra gas parameter.
func (k Keeper) GetDepositExtraGas(ctx sdk.Context) (gas sdk.Gas) {
	k.paramSpace.Get(ctx, types.KeyDepositExtraGas, &gas)
	return
}

// GetWithdrawExtraGas returns the current withdraw extra gas parameter.
func (k Keeper) GetWithdrawExtraGas(ctx sdk.Context) (gas sdk.Gas) {
	k.paramSpace.Get(ctx, types.KeyWithdrawExtraGas, &gas)
	return
}

// GetOrderExtraGas returns the current order extra gas parameter.
func (k Keeper) GetOrderExtraGas(ctx sdk.Context) (gas sdk.Gas) {
	k.paramSpace.Get(ctx, types.KeyOrderExtraGas, &gas)
	return
}
