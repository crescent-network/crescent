package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// marketmaker module sentinel errors
var (
	ErrAlreadyExistMarketMaker = sdkerrors.Register(ModuleName, 2, "already exist market maker")
	ErrEmptyClaimableIncentive = sdkerrors.Register(ModuleName, 3, "empty claimable incentives")
	ErrNotExistMarketMaker     = sdkerrors.Register(ModuleName, 4, "not exist market maker")
	ErrInvalidPairId           = sdkerrors.Register(ModuleName, 5, "invalid pair id")
	ErrUnregisteredPairId      = sdkerrors.Register(ModuleName, 6, "unregistered pair id")
	ErrInvalidDeposit          = sdkerrors.Register(ModuleName, 7, "invalid apply deposit")
	ErrInvalidInclusion        = sdkerrors.Register(ModuleName, 8, "invalid inclusion, already eligible")
	ErrInvalidExclusion        = sdkerrors.Register(ModuleName, 9, "invalid exclusion, not eligible")
	ErrInvalidRejection        = sdkerrors.Register(ModuleName, 10, "invalid rejection, already eligible")
	ErrNotEligibleMarketMaker  = sdkerrors.Register(ModuleName, 11, "invalid distribution, not eligible")
)
