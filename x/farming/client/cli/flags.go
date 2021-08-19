package cli

// DONTCOVER

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPlanType          = "plan-type"
	FlagFarmingPoolAddr   = "farming-pool-addr"
	FlagTerminationAddr   = "termination-addr"
	FlagFarmerAddr        = "farmer-addr"
	FlagStakingCoinDenom  = "staking-coin-denom"
	FlagStakingCoinDenoms = "staking-coin-denoms"
)

func flagSetPlans() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagPlanType, "", "The plan type; private or public")
	fs.String(FlagFarmingPoolAddr, "", "The bech32 address of the farming pool account")
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

func flagSetHarvest() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.StringSlice(FlagStakingCoinDenoms, []string{""}, "The staking coin denoms to harvest farming rewards")

	return fs
}
