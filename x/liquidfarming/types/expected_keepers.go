package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	farmtypes "github.com/crescent-network/crescent/v3/x/farm/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// AccountKeeper defines the expected interface needed for the module.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
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

// FarmKeeper defines the expected interface needed for the module.
type FarmKeeper interface {
	Farm(ctx sdk.Context, farmerAddr sdk.AccAddress, coin sdk.Coin) (withdrawnRewards sdk.Coins, err error)
	Unfarm(ctx sdk.Context, farmerAddr sdk.AccAddress, coin sdk.Coin) (withdrawnRewards sdk.Coins, err error)
	Harvest(ctx sdk.Context, farmerAddr sdk.AccAddress, denom string) (withdrawnRewards sdk.Coins, err error)
	Rewards(ctx sdk.Context, farmerAddr sdk.AccAddress, denom string) sdk.DecCoins
	GetFarm(ctx sdk.Context, denom string) (farm farmtypes.Farm, found bool)
	GetPosition(ctx sdk.Context, farmerAddr sdk.AccAddress, denom string) (position farmtypes.Position, found bool)
}

// LiquidityKeeper defines the expected interface needed for the module.
type LiquidityKeeper interface {
	GetPool(ctx sdk.Context, id uint64) (pool liquiditytypes.Pool, found bool)
	Withdraw(ctx sdk.Context, msg *liquiditytypes.MsgWithdraw) (liquiditytypes.WithdrawRequest, error)
}
