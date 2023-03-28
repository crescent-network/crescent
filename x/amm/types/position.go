package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewPosition(id, poolId uint64, ownerAddr sdk.AccAddress, lowerTick, upperTick int32) Position {
	return Position{
		Id:        id,
		PoolId:    poolId,
		Owner:     ownerAddr.String(),
		LowerTick: lowerTick,
		UpperTick: upperTick,
		Liquidity: sdk.ZeroInt(),
	}
}
