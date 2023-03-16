package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func TestRewardsAuctionValidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(*types.RewardsAuction)
		expectedErr string
	}{
		{
			"happy case",
			func(auction *types.RewardsAuction) {
				registry := codectypes.NewInterfaceRegistry()
				types.RegisterInterfaces(registry)
				cdc := codec.NewProtoCodec(registry)
				bz := types.MustMarshalRewardsAuction(cdc, *auction)
				newAuction := types.MustUnmarshalRewardsAuction(cdc, bz)
				require.EqualValues(t, auction, &newAuction)
			},
			"",
		},
		{
			"invalid pool id",
			func(auction *types.RewardsAuction) {
				auction.PoolId = 0
			},
			"pool id must not be 0",
		},
		{
			"invalid bidding coin denom",
			func(auction *types.RewardsAuction) {
				auction.BiddingCoinDenom = ""
			},
			"denom must not be empty",
		},
		{
			"invalid bidding coin denom",
			func(auction *types.RewardsAuction) {
				auction.BiddingCoinDenom = "123!@#$%"
			},
			"invalid coin denom",
		},
		{
			"invalid paying reserve address",
			func(auction *types.RewardsAuction) {
				auction.PayingReserveAddress = "invalidaddr"
			},
			"invalid paying reserve address decoding bech32 failed: invalid separator index -1",
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
				auction.Status = types.AuctionStatusNil
			},
			"invalid auction status",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auction := types.NewRewardsAuction(
				1,
				1,
				utils.ParseTime("0001-01-01T00:00:00Z"),
				utils.ParseTime("9999-12-31T00:00:00Z"),
			)
			auction.SetStatus(types.AuctionStatusStarted)
			auction.SetWinner("")
			auction.SetWinningAmount(utils.ParseCoin("100000pool1"))
			auction.SetRewards(utils.ParseCoins("100000denom1"))
			auction.SetFees(utils.ParseCoins("10denom1"))
			auction.SetFeeRate(utils.ParseDec("0.05"))
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
		malleate    func(*types.Bid)
		expectedErr string
	}{
		{
			"happy case",
			func(b *types.Bid) {},
			"",
		},
		{
			"invalid pool id",
			func(b *types.Bid) {
				b.PoolId = 0
			},
			"pool id must not be 0",
		},
		{
			"invalid bidder",
			func(b *types.Bid) {
				b.Bidder = "invalidaddr"
			},
			"invalid bidder address decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid bidding amount",
			func(b *types.Bid) {
				b.Amount = utils.ParseCoin("0pool1")
			},
			"amount must be positive value",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bid := types.NewBid(
				1,
				sdk.AccAddress(crypto.AddressHash([]byte("address1"))).String(),
				utils.ParseCoin("100000000pool1"),
			)
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
