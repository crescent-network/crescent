package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	// authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	// SetModuleAccount(ctx sdk.Context, macc authtypes.ModuleAccountI)
	// GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error // TODO: mint from module account to airdrop source address
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	// MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}
