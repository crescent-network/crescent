package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

// x/liquidity module sentinel errors
var (
	ErrInsufficientDepositAmount = sdkerrors.Register(ModuleName, 2, "insufficient deposit amount")
	ErrPairAlreadyExists         = sdkerrors.Register(ModuleName, 3, "pair already exists")
	ErrPoolAlreadyExists         = sdkerrors.Register(ModuleName, 4, "pool already exists")
	ErrWrongPoolCoinDenom        = sdkerrors.Register(ModuleName, 5, "wrong pool coin denom")
	ErrInvalidCoinDenom          = sdkerrors.Register(ModuleName, 6, "invalid coin denom")
	ErrInvalidPriceTick          = sdkerrors.Register(ModuleName, 7, "price not fit into ticks")
	ErrPriceOutOfRange           = sdkerrors.Register(ModuleName, 8, "price out of range limit")
	ErrTooLongOrderLifespan      = sdkerrors.Register(ModuleName, 9, "order lifespan is too long")
	ErrDisabledPool              = sdkerrors.Register(ModuleName, 10, "disabled pool")
	ErrWrongPair                 = sdkerrors.Register(ModuleName, 11, "wrong denom pair")
)
