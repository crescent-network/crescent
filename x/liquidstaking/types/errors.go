package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// Sentinel errors for the liquidstaking module.
var (
	ErrActiveLiquidValidatorsNotExists = sdkerrors.Register(ModuleName, 2, "active liquid validators not exists")
	ErrInvalidDenom                    = sdkerrors.Register(ModuleName, 3, "invalid denom")
	ErrInvalidBondDenom                = sdkerrors.Register(ModuleName, 4, "invalid bond denom")
	ErrInvalidLiquidBondDenom          = sdkerrors.Register(ModuleName, 5, "invalid liquid bond denom")
	ErrNotImplementedYet               = sdkerrors.Register(ModuleName, 6, "not implemented yet")
	ErrLessThanMinLiquidStakingAmount  = sdkerrors.Register(ModuleName, 7, "staking amount should be over params.min_liquid_staking_amount")
	ErrInvalidBTokenSupply             = sdkerrors.Register(ModuleName, 8, "invalid liquid bond denom supply")
	ErrInvalidActiveLiquidValidators   = sdkerrors.Register(ModuleName, 9, "invalid active liquid validators")
	ErrLiquidValidatorsNotExists       = sdkerrors.Register(ModuleName, 10, "liquid validators not exists")
	ErrInsufficientProxyAccBalance     = sdkerrors.Register(ModuleName, 11, "insufficient liquid tokens or balance of proxy account, need to wait for new liquid validator to be added or unbonding of proxy account to be completed")
	ErrTooSmallLiquidStakingAmount     = sdkerrors.Register(ModuleName, 12, "liquid staking amount is too small, the result becomes zero")
	ErrTooSmallLiquidUnstakingAmount   = sdkerrors.Register(ModuleName, 13, "liquid unstaking amount is too small, the result becomes zero")
)
