package cli

// DONTCOVER

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPlanType         = "plan-type"
	FlagFarmingPoolAddr  = "farming-pool-addr"
	FlagRewardPoolAddr   = "reward-pool-addr"
	FlagTerminationAddr  = "termination-addr"
	FlagStakingCoinDenom = "staking-coin-denom"
	FlagFarmerAddr       = "farmer-addr"
)

func flagSetPlans() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagPlanType, "", "The plan type; private or public")
	fs.String(FlagFarmingPoolAddr, "", "The bech32 address of the farming pool account")
	fs.String(FlagRewardPoolAddr, "", "The bech32 address of the reward pool account")
	fs.String(FlagTerminationAddr, "", "The bech32 address of the termination account")
	fs.String(FlagStakingCoinDenom, "", "The staking coin denom")

	return fs
}

func flagSetStaking() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagStakingCoinDenom, "", "The staking coin denom")
	fs.String(FlagFarmerAddr, "", "The bech32 address of the farmer account")

	return fs
}

func flagSetRewards() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagStakingCoinDenom, "", "The staking coin denom")
	fs.String(FlagFarmerAddr, "", "The bech32 address of the farmer account")

	return fs
}
