package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	farmingtypes "github.com/crescent-network/crescent/v2/x/farming/types"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
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

// FarmingKeeper defines the expected interface needed for the module.
type FarmingKeeper interface {
	Stake(ctx sdk.Context, farmerAcc sdk.AccAddress, amount sdk.Coins) error
	Unstake(ctx sdk.Context, farmerAcc sdk.AccAddress, amount sdk.Coins) error
	Harvest(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenoms []string) error
	Rewards(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string) sdk.Coins
	GetCurrentEpochDays(ctx sdk.Context) uint32
	GetStaking(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress) (staking farmingtypes.Staking, found bool)
	GetAllQueuedStakingAmountByFarmerAndDenom(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string) sdk.Int
}

// LiquidityKeeper defines the expected interface needed for the module.
type LiquidityKeeper interface {
	GetPool(ctx sdk.Context, id uint64) (pool liquiditytypes.Pool, found bool)
	Withdraw(ctx sdk.Context, msg *liquiditytypes.MsgWithdraw) (liquiditytypes.WithdrawRequest, error)
}
