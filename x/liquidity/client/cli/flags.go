package cli

// DONTCOVER

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPairId        = "pair-id"
	FlagXCoinDenom    = "x-coin-denom"
	FlagYCoinDenom    = "y-coin-denom"
	FlagPoolCoinDenom = "pool-coin-denom"
	FlagReserveAcc    = "reserve-acc"
)

func flagSetPools() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagXCoinDenom, "", "The X coin denomination")
	fs.String(FlagYCoinDenom, "", "The Y coin denomination")
	fs.String(FlagPairId, "", "The pair id")

	return fs
}

func flagSetPool() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagPoolCoinDenom, "", "The denomination of the pool coin")
	fs.String(FlagReserveAcc, "", "The Bech32 address of the reserve account")

	return fs
}

func flagSetPairs() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagXCoinDenom, "", "The X coin denomination")
	fs.String(FlagYCoinDenom, "", "The Y coin denomination")

	return fs
}
