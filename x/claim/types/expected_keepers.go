package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	liqtypes "github.com/cosmosquad-labs/squad/x/liquidity/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// FarmingKeeper defines the expected interface needed to check the condition.
type FarmingKeeper interface {
	GetAllQueuedCoinsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress) sdk.Coins
	GetAllStakedCoinsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress) sdk.Coins
}

// LiquidityKeeper defines the expected interface needed to check the condition.
type LiquidityKeeper interface {
	GetAllOrders(ctx sdk.Context) (reqs []liqtypes.Order)
}
