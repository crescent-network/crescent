package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func NewPool(id uint64, denom0, denom1 string, tickSpacing uint32, reserveAddr sdk.AccAddress) Pool {
	return Pool{
		Id:               id,
		Denom0:           denom0,
		Denom1:           denom1,
		TickSpacing:      tickSpacing,
		ReserveAddress:   reserveAddr.String(),
		CurrentTick:      0,
		CurrentSqrtPrice: sdk.ZeroDec(), // not initialized
		CurrentLiquidity: sdk.ZeroInt(),
		Initialized:      false,
	}
}

func DerivePoolReserveAddress(poolId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("PoolReserveAddress/%d", poolId)))
}