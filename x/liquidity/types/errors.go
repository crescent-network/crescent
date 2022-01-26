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
	ErrNoLastPrice               = sdkerrors.Register(ModuleName, 8, "cannot make a market order to a pair with no last price")
	ErrInsufficientOfferCoin     = sdkerrors.Register(ModuleName, 9, "insufficient offer coin")
	ErrPriceOutOfRange           = sdkerrors.Register(ModuleName, 10, "price out of range limit")
	ErrTooLongOrderLifespan      = sdkerrors.Register(ModuleName, 11, "order lifespan is too long")
	ErrDisabledPool              = sdkerrors.Register(ModuleName, 12, "disabled pool")
	ErrWrongPair                 = sdkerrors.Register(ModuleName, 13, "wrong denom pair")
)
