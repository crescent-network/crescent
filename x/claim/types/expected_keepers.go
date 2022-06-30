package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
	liquidstakingtypes "github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// AccountKeeper is the expected x/auth module keeper.
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	// MintCoins is used only for simulation test codes
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// DistrKeeper is the keeper of the distribution store
type DistrKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type GovKeeper interface {
	IterateProposals(ctx sdk.Context, cb func(proposal govtypes.Proposal) (stop bool))
	GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (vote govtypes.Vote, found bool)
	// Note that this function is only used before the UpgradeHeight defined in app/upgrades/v1.1.0
	IterateAllVotes(ctx sdk.Context, cb func(vote govtypes.Vote) (stop bool))
}

// LiquidityKeeper defines the expected interface needed to check the condition.
type LiquidityKeeper interface {
	GetDepositRequestsByDepositor(ctx sdk.Context, depositor sdk.AccAddress) (reqs []liquiditytypes.DepositRequest)
	GetOrdersByOrderer(ctx sdk.Context, orderer sdk.AccAddress) (orders []liquiditytypes.Order)
}

type LiquidStakingKeeper interface {
	GetParams(ctx sdk.Context) (params liquidstakingtypes.Params)
}
