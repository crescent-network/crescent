package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v5/types"
)

func NewPool(id uint64, denom0, denom1 string, tickSpacing uint32, reserveAddr sdk.AccAddress) Pool {
	return Pool{
		Id:             id,
		Denom0:         denom0,
		Denom1:         denom1,
		TickSpacing:    tickSpacing,
		ReserveAddress: reserveAddr.String(),
	}
}

func (pool Pool) DenomIn(isBuy bool) string {
	if isBuy {
		return pool.Denom0
	}
	return pool.Denom1
}

func (pool Pool) DenomOut(isBuy bool) string {
	if isBuy {
		return pool.Denom1
	}
	return pool.Denom0
}

func NewPoolState(tick int32, sqrtPrice sdk.Dec) PoolState {
	return PoolState{
		CurrentTick:      tick,
		CurrentSqrtPrice: sqrtPrice,
		CurrentLiquidity: utils.ZeroDec,
	}
}

func DerivePoolReserveAddress(poolId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("PoolReserveAddress/%d", poolId)))
}
