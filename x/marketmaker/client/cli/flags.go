package cli

// DONTCOVER

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagAddress  = "address"
	FlagPairId   = "pair-id"
	FlagEligible = "eligible"
)

func flagSetMarketMakers() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagAddress, "", "The market maker address")
	fs.String(FlagPairId, "", "The pair id")
	fs.String(FlagEligible, "", "Whether the market maker is eligible or not")

	return fs
}
