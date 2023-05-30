package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

// AccountKeeper defines the expected keeper interface of the auth module.
// Some methods are used only in simulation tests.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the expected keeper interface of the bank module.
// Some methods are used only in simulation tests.
type BankKeeper interface {
	HasSupply(ctx sdk.Context, denom string) bool
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error

	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type ExchangeKeeper interface {
	GetMarket(ctx sdk.Context, marketId uint64) (market exchangetypes.Market, found bool)
}

type MarkerKeeper interface {
	GetLastBlockTime(ctx sdk.Context) (t time.Time, found bool)
}
