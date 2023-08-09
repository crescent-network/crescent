package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func TestRewardsAuction_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(auction *types.RewardsAuction)
		expectedErr string
	}{
		{
			"happy case",
			func(auction *types.RewardsAuction) {},
			"",
		},
		{
			"invalid public position id",
			func(auction *types.RewardsAuction) {
				auction.PublicPositionId = 0
			},
			"public position id must not be 0",
		},
		{
			"invalid id",
			func(auction *types.RewardsAuction) {
				auction.Id = 0
			},
			"id must not be 0",
		},
		{
			"invalid start and end time",
			func(auction *types.RewardsAuction) {
				auction.StartTime = utils.ParseTime("9999-12-31T00:00:00Z")
				auction.EndTime = utils.ParseTime("0001-01-01T00:00:00Z")
			},
			"end time must be set after the start time",
		},
		{
			"invalid auction status",
			func(auction *types.RewardsAuction) {
				auction.Status = 10
			},
			"invalid auction status: 10",
		},
		{
			"invalid winning bid",
			func(auction *types.RewardsAuction) {
				auction.WinningBid = &types.Bid{
					PublicPositionId: 0,
					RewardsAuctionId: 0,
					Bidder:           "",
					Share:            sdk.Coin{},
				}
			},
			"invalid winning bid: public position id must not be 0",
		},
		{
			"invalid rewards",
			func(auction *types.RewardsAuction) {
				auction.Rewards = sdk.Coins{utils.ParseCoin("0uatom")}
			},
			"invalid rewards: coin 0uatom amount is not positive",
		},
		{
			"invalid fees",
			func(auction *types.RewardsAuction) {
				auction.Fees = sdk.Coins{utils.ParseCoin("0uatom")}
			},
			"invalid fees: coin 0uatom amount is not positive",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auction := types.NewRewardsAuction(
				1, 2, utils.ParseTime("2023-05-01T00:00:00Z"), utils.ParseTime("2023-05-01T01:00:00Z"),
				types.AuctionStatusStarted)
			winningBid := types.NewBid(1, 2, utils.TestAddress(1), utils.ParseCoin("10000lashare1"))
			auction.SetWinningBid(&winningBid)
			auction.SetRewards(utils.ParseCoins("100000uatom"))
			auction.SetFees(utils.ParseCoins("300uatom"))
			tc.malleate(&auction)
			err := auction.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestBidValidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(bid *types.Bid)
		expectedErr string
	}{
		{
			"happy case",
			func(bid *types.Bid) {},
			"",
		},
		{
			"invalid public position id",
			func(bid *types.Bid) {
				bid.PublicPositionId = 0
			},
			"public position id must not be 0",
		},
		{
			"invalid auction id",
			func(bid *types.Bid) {
				bid.RewardsAuctionId = 0
			},
			"rewards auction id must not be 0",
		},
		{
			"invalid bidder",
			func(bid *types.Bid) {
				bid.Bidder = "invalidaddr"
			},
			"invalid bidder address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid share",
			func(bid *types.Bid) {
				bid.Share = utils.ParseCoin("0lashare1")
			},
			"share amount must be positive: 0lashare1",
		},
		{
			"invalid share denom",
			func(bid *types.Bid) {
				bid.Share = utils.ParseCoin("10000lashare2")
			},
			"share denom must be lashare1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bid := types.NewBid(1, 2, utils.TestAddress(1), utils.ParseCoin("10000lashare1"))
			tc.malleate(&bid)
			err := bid.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
