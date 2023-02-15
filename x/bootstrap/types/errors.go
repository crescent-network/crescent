package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// bootstrap module sentinel errors
var (
	ErrPoolAlreadyExists     = sdkerrors.Register(ModuleName, 2, "pool already exists")
	ErrWrongPoolCoinDenom    = sdkerrors.Register(ModuleName, 3, "wrong pool coin denom")
	ErrInvalidCoinDenom      = sdkerrors.Register(ModuleName, 4, "invalid coin denom")
	ErrNoLastPrice           = sdkerrors.Register(ModuleName, 5, "cannot make a market order to a pair with no last price")
	ErrInsufficientOfferCoin = sdkerrors.Register(ModuleName, 6, "insufficient offer coin")
	ErrPriceOutOfRange       = sdkerrors.Register(ModuleName, 7, "price out of range limit")
	ErrInactivePool          = sdkerrors.Register(ModuleName, 8, "inactive pool")
	ErrWrongPair             = sdkerrors.Register(ModuleName, 9, "wrong denom pair")
	ErrAlreadyCanceled       = sdkerrors.Register(ModuleName, 10, "the order is already canceled")
	ErrDuplicatePoolId       = sdkerrors.Register(ModuleName, 11, "duplicate pool id presents in the pool id list")
	ErrTooSmallOrder         = sdkerrors.Register(ModuleName, 12, "too small order")
	ErrTooLargePool          = sdkerrors.Register(ModuleName, 13, "too large pool")
	ErrTooManyPools          = sdkerrors.Register(ModuleName, 14, "too many pools in the pair")
	ErrPriceNotOnTicks       = sdkerrors.Register(ModuleName, 15, "price is not on ticks")
)
