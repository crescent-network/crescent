package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

// AccountKeeper defines the expected keeper interface of the auth module.
// Some methods are used only in simulation tests.
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the expected keeper interface of the bank module.
// Some methods are used only in simulation tests.
type BankKeeper interface {
	HasSupply(ctx sdk.Context, denom string) bool
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	// Just for simulations
	MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type ExchangeKeeper interface {
	GetMarket(ctx sdk.Context, marketId uint64) (market exchangetypes.Market, found bool)
	LookupMarket(ctx sdk.Context, marketId uint64) (found bool)
	IterateAllMarkets(ctx sdk.Context, cb func(market exchangetypes.Market) (stop bool))
	MustGetMarketState(ctx sdk.Context, marketId uint64) (marketState exchangetypes.MarketState)
}

type MarkerKeeper interface {
	GetLastBlockTime(ctx sdk.Context) (t time.Time, found bool)
}
