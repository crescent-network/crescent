package cli

// DONTCOVER

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPlanType         = "plan-type"
	FlagFarmingPoolAddr  = "farming-pool-addr"
	FlagTerminationAddr  = "termination-addr"
	FlagStakingCoinDenom = "staking-coin-denom"
	FlagTerminated       = "terminated"
	FlagAll              = "all"
)

// flagSetPlans returns the FlagSet used for farming plan related opertations.
func flagSetPlans() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagPlanType, "", "The plan type; private or public")
	fs.String(FlagFarmingPoolAddr, "", "The bech32 address of the farming pool account")
	fs.String(FlagTerminationAddr, "", "The bech32 address of the termination account")
	fs.String(FlagStakingCoinDenom, "", "The staking coin denom")
	fs.String(FlagTerminated, "", "Whether the plan is terminated or not (true/false)")

	return fs
}

// flagSetRewards returns the FlagSet used for farmer's rewards.
func flagSetRewards() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagStakingCoinDenom, "", "The staking coin denom")

	return fs
}

// flagSetHarvest returns the FlagSet used for harvest all staking coin denoms.
func flagSetHarvest() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Bool(FlagAll, false, "Harvest for all staking coin denoms")

	return fs
}
