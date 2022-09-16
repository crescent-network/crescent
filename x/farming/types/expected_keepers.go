package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// BankKeeper defines the expected bank send keeper
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
	BlockedAddr(ctx sdk.Context, addr sdk.AccAddress) bool
	AddBlockedAddr(ctx sdk.Context, addr sdk.AccAddress)
	RemoveBlockedAddr(ctx sdk.Context, addr sdk.AccAddress)
	// MintCoins is used only for simulation test codes
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetModuleAddress(name string) sdk.AccAddress
}

// Event hooks
// These can be utilized to communicate between a farming keeper and other keepers from other modules.
// The other keepers must implement this interface, which then the farming keeper can call.

// FarmingHooks event hooks for the farming object (noalias)
type FarmingHooks interface {
	AfterAllocateRewards(ctx sdk.Context) // Must be called when farming rewards are allocated
}
