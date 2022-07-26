package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	farmintypes "github.com/crescent-network/crescent/v2/x/farming/types"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
)

// BankKeeper defines the expected bank send keeper
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	SendCoinsFromModuleToModule(ctx sdk.Context, senderPool, recipientPool string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

// StakingKeeper expected staking keeper (noalias)
type StakingKeeper interface {
	Validator(sdk.Context, sdk.ValAddress) stakingtypes.ValidatorI
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingtypes.ValidatorI
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)

	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator

	GetLastTotalPower(ctx sdk.Context) sdk.Int
	GetLastValidatorPower(ctx sdk.Context, valAddr sdk.ValAddress) int64

	Delegation(sdk.Context, sdk.AccAddress, sdk.ValAddress) stakingtypes.DelegationI
	GetDelegation(ctx sdk.Context,
		delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool)
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation stakingtypes.DelegationI) (stop bool))
	Delegate(
		ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Int, tokenSrc stakingtypes.BondStatus,
		validator stakingtypes.Validator, subtractAccount bool,
	) (newShares sdk.Dec, err error)

	BondDenom(ctx sdk.Context) (res string)
	Unbond(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, shares sdk.Dec) (amount sdk.Int, err error)
	UnbondingTime(ctx sdk.Context) (res time.Duration)
	SetUnbondingDelegationEntry(
		ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
		creationHeight int64, minTime time.Time, balance sdk.Int,
	) stakingtypes.UnbondingDelegation
	InsertUBDQueue(ctx sdk.Context, ubd stakingtypes.UnbondingDelegation,
		completionTime time.Time)
	ValidateUnbondAmount(
		ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int,
	) (shares sdk.Dec, err error)
	GetAllUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.UnbondingDelegation
	BeginRedelegation(
		ctx sdk.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress, sharesAmount sdk.Dec,
	) (completionTime time.Time, err error)
	GetAllRedelegations(
		ctx sdk.Context, delegator sdk.AccAddress, srcValAddress, dstValAddress sdk.ValAddress,
	) []stakingtypes.Redelegation
	HasReceivingRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) bool
	BlockValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate
	HasMaxUnbondingDelegationEntries(ctx sdk.Context,
		delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) bool
}

// GovKeeper expected gov keeper (noalias)
type GovKeeper interface {
	Tally(ctx sdk.Context, proposal govtypes.Proposal) (passes bool, burnDeposits bool, tallyResults govtypes.TallyResult)
	GetVotes(ctx sdk.Context, proposalID uint64) (votes govtypes.Votes)
	GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (vote govtypes.Vote, found bool)
	GetProposal(ctx sdk.Context, proposalID uint64) (govtypes.Proposal, bool)
	SetProposal(ctx sdk.Context, proposal govtypes.Proposal)
	GetProposals(ctx sdk.Context) (proposals govtypes.Proposals)
	AddVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress, options govtypes.WeightedVoteOptions) error
}

// DistrKeeper expected distribution keeper (noalias)
type DistrKeeper interface {
	IncrementValidatorPeriod(ctx sdk.Context, val stakingtypes.ValidatorI) uint64
	CalculateDelegationRewards(ctx sdk.Context, val stakingtypes.ValidatorI, del stakingtypes.DelegationI, endingPeriod uint64) (rewards sdk.DecCoins)
	WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
}

// Liquidity expected liquidity keeper (noalias)
type LiquidityKeeper interface {
	GetPair(ctx sdk.Context, id uint64) (pair liquiditytypes.Pair, found bool)
	GetPool(ctx sdk.Context, id uint64) (pool liquiditytypes.Pool, found bool)
	GetPoolBalances(ctx sdk.Context, pool liquiditytypes.Pool) (rx sdk.Coin, ry sdk.Coin)
	GetPoolCoinSupply(ctx sdk.Context, pool liquiditytypes.Pool) sdk.Int
	IterateAllPools(ctx sdk.Context, cb func(pool liquiditytypes.Pool) (stop bool, err error)) error
}

// FarmingKeeper expected farming keeper (noalias)
type FarmingKeeper interface {
	IterateStakingsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress, cb func(stakingCoinDenom string, staking farmintypes.Staking) (stop bool))
	IterateQueuedStakingsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress, cb func(stakingCoinDenom string, endTime time.Time, queuedStaking farmintypes.QueuedStaking) (stop bool))
}

// SlashingKeeper expected slashing keeper (noalias)
type SlashingKeeper interface {
	IsTombstoned(ctx sdk.Context, consAddr sdk.ConsAddress) bool
}

// StakingHooks event hooks for staking validator object (noalias)
type StakingHooks interface {
	AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)                           // Must be called when a validator is created
	AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) // Must be called when a validator is deleted

	BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress)        // Must be called when a delegation is created
	BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) // Must be called when a delegation's shares are modified
	AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress)
	BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec)
}
