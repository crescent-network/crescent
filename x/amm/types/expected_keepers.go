package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

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
	InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error

	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type ExchangeKeeper interface {
	GetSpotMarket(ctx sdk.Context, marketId string) (market exchangetypes.SpotMarket, found bool)
	PlaceSpotLimitOrder(ctx sdk.Context, ordererAddr sdk.AccAddress, market exchangetypes.SpotMarket, isBuy bool, priceLimit sdk.Dec, qty sdk.Int) (order exchangetypes.SpotOrder, execQuote sdk.Int, err error)
	CancelSpotOrder(ctx sdk.Context, senderAddr sdk.AccAddress, marketId string, orderId uint64) (order exchangetypes.SpotOrder, err error)
}
