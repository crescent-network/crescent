package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// AccountKeeper defines the expected keeper interface of the auth module.
// Some methods are used only in simulation tests.
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

// BankKeeper defines the expected keeper interface of the bank module.
// Some methods are used only in simulation tests.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error

	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// LiquidityKeeper defines the expected keeper interface of the liquidity module.
type LiquidityKeeper interface {
	GetPair(ctx sdk.Context, id uint64) (pair liquiditytypes.Pair, found bool)
	GetAllPairs(ctx sdk.Context) (pairs []liquiditytypes.Pair)
	IteratePoolsByPair(ctx sdk.Context, pairId uint64, cb func(pool liquiditytypes.Pool) (stop bool, err error)) error
}
