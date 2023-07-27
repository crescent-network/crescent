package cli

// DONTCOVER

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagRewardsAuctionStatus = "status"
)

// flagSetRewardsAuctions returns the FlagSet used for farming plan related opertations.
func flagSetRewardsAuctions() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagRewardsAuctionStatus, "", "The rewards auction status; AUCTION_STATUS_STARTED, AUCTION_STATUS_FINISHED, or AUCTION_STATUS_SKIPPED")

	return fs
}
