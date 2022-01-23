package cli

// DONTCOVER

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPairId        = "pair-id"
	FlagDisabled      = "disabled"
	FlagPoolCoinDenom = "pool-coin-denom"
	FlagReserveAcc    = "reserve-acc"
	FlagDenoms        = "denoms"
)

func flagSetPools() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagPairId, "", "The pair id")
	fs.String(FlagDisabled, "", "Whether the pool is disabled or not")

	return fs
}

func flagSetPool() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagPoolCoinDenom, "", "The denomination of the pool coin")
	fs.String(FlagReserveAcc, "", "The bech-32 encoded address of the reserve account")

	return fs
}

func flagSetPairs() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.StringSlice(FlagDenoms, []string{}, "Coin denominations to query")

	return fs
}
