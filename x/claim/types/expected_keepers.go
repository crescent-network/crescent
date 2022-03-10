package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	liqtypes "github.com/cosmosquad-labs/squad/x/liquidity/types"
	lstypes "github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// DistrKeeper is the keeper of the distribution store
type DistrKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type GovKeeper interface {
	GetProposals(ctx sdk.Context) (proposals govtypes.Proposals)
	GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (vote govtypes.Vote, found bool)
	GetAllVotes(ctx sdk.Context) (votes govtypes.Votes)
}

// LiquidityKeeper defines the expected interface needed to check the condition.
type LiquidityKeeper interface {
	GetDepositRequestsByDepositor(ctx sdk.Context, depositor sdk.AccAddress) (reqs []liqtypes.DepositRequest)
	GetOrdersByOrderer(ctx sdk.Context, orderer sdk.AccAddress) (orders []liqtypes.Order)
}

type LiquidStakingKeeper interface {
	GetParams(ctx sdk.Context) (params lstypes.Params)
}
