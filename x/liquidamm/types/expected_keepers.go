package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
)

// AccountKeeper defines the expected interface needed for the module.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

// BankKeeper defines the expected interface needed for the module.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type AMMKeeper interface {
	LookupPool(ctx sdk.Context, poolId uint64) (found bool)
	GetPool(ctx sdk.Context, poolId uint64) (pool ammtypes.Pool, found bool)
	GetPositionByParams(
		ctx sdk.Context, ownerAddr sdk.AccAddress, poolId uint64, lowerTick, upperTick int32) (position ammtypes.Position, found bool)
	AddLiquidity(
		ctx sdk.Context, ownerAddr, fromAddr sdk.AccAddress, poolId uint64,
		lowerPrice, upperPrice sdk.Dec, desiredAmt sdk.Coins) (position ammtypes.Position, liquidity sdk.Int, amt sdk.Coins, err error)
	RemoveLiquidity(
		ctx sdk.Context, ownerAddr, toAddr sdk.AccAddress,
		positionId uint64, liquidity sdk.Int) (position ammtypes.Position, amt sdk.Coins, err error)
	Collect(
		ctx sdk.Context, ownerAddr, toAddr sdk.AccAddress, positionId uint64, amt sdk.Coins) error
	CollectibleCoins(ctx sdk.Context, positionId uint64) (fee, farmingRewards sdk.Coins, err error)

	// used in simulation

	IterateAllPools(ctx sdk.Context, cb func(pool ammtypes.Pool) (stop bool))
	MustGetPoolState(ctx sdk.Context, poolId uint64) ammtypes.PoolState
}
