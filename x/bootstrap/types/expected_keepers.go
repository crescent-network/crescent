package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
)

// BankKeeper defines the expected bank send keeper
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	// MintCoins is used only for simulation test codes
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)
	GetModuleAddress(name string) sdk.AccAddress
}

type LiquidityKeeper interface {
	GetPair(ctx sdk.Context, id uint64) (pair liquiditytypes.Pair, found bool)
	GetPool(ctx sdk.Context, id uint64) (pool liquiditytypes.Pool, found bool)
}
