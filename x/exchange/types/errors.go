package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrSwapNotEnoughInput     = sdkerrors.Register(ModuleName, 2, "not enough swap input amount")
	ErrSwapNotEnoughOutput    = sdkerrors.Register(ModuleName, 3, "not enough swap output amount")
	ErrSwapNotEnoughLiquidity = sdkerrors.Register(ModuleName, 4, "not enough liquidity in the market")
	ErrOrderPriceOutOfRange   = sdkerrors.Register(ModuleName, 5, "order price out of range")
	ErrMaxNumMMOrdersExceeded = sdkerrors.Register(ModuleName, 6, "number of MM orders exceeded the limit")
)
