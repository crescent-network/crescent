package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func TestGenesisState_Validate(t *testing.T) {
	validPayingReserveAddr := sdk.AccAddress(crypto.AddressHash([]byte("validPayingReserveAddr")))
	validBidder := sdk.AccAddress(crypto.AddressHash([]byte("validBidder")))
	validDenom := "denom1"
	validPoolId := uint64(1)
	validAuctionId := uint64(1)

	for _, tc := range []struct {
		name        string
		malleate    func(genState *types.GenesisState)
		expectedErr string
	}{
		{
			"default is valid",
			func(genState *types.GenesisState) {},
			"",
		},
		{
			"valid liquid farm",
			func(genState *types.GenesisState) {
				genState.LiquidFarms = []types.LiquidFarm{
					{
						PoolId:        1,
						MinFarmAmount: sdk.ZeroInt(),
						MinBidAmount:  sdk.ZeroInt(),
						FeeRate:       sdk.ZeroDec(),
					},
				}
			},
			"",
		},
		{
			"invalid liquid farm: pool id",
			func(genState *types.GenesisState) {
				genState.LiquidFarms = []types.LiquidFarm{
					{
						PoolId: 0,
					},
				}
			},
			"invalid liquid farm pool id must not be 0",
		},
		{
			"invalid rewards auction: pool id",
			func(genState *types.GenesisState) {
				genState.RewardsAuctions = []types.RewardsAuction{
					{
						PoolId:               0,
						BiddingCoinDenom:     validDenom,
						PayingReserveAddress: validPayingReserveAddr.String(),
						StartTime:            utils.ParseTime("2022-08-01T00:00:00Z"),
						EndTime:              utils.ParseTime("2022-08-02T00:00:00Z"),
						Status:               types.AuctionStatusStarted,
					},
				}
			},
			"pool id must not be 0",
		},
		{
			"invalid rewards auction: bidding coin denom",
			func(genState *types.GenesisState) {
				genState.RewardsAuctions = []types.RewardsAuction{
					{
						PoolId:               validPoolId,
						BiddingCoinDenom:     "123!@#$%",
						PayingReserveAddress: validPayingReserveAddr.String(),
						StartTime:            utils.ParseTime("2022-08-01T00:00:00Z"),
						EndTime:              utils.ParseTime("2022-08-02T00:00:00Z"),
						Status:               types.AuctionStatusStarted,
					},
				}
			},
			"invalid coin denom",
		},
		{
			"invalid rewards auction: auction status",
			func(genState *types.GenesisState) {
				genState.RewardsAuctions = []types.RewardsAuction{
					{
						PoolId:               validPoolId,
						BiddingCoinDenom:     validDenom,
						PayingReserveAddress: validPayingReserveAddr.String(),
						StartTime:            utils.ParseTime("2022-08-01T00:00:00Z"),
						EndTime:              utils.ParseTime("2022-08-02T00:00:00Z"),
						Status:               types.AuctionStatusNil,
					},
				}
			},
			"invalid auction status",
		},
		{
			"invalid bids: pool id",
			func(genState *types.GenesisState) {
				genState.Bids = []types.Bid{
					{
						PoolId: 0,
						Bidder: validBidder.String(),
						Amount: utils.ParseCoin("1000000pool1"),
					},
				}
			},
			"pool id must not be 0",
		},
		{
			"invalid bids: bid amount",
			func(genState *types.GenesisState) {
				genState.Bids = []types.Bid{
					{
						PoolId: validPoolId,
						Bidder: validBidder.String(),
						Amount: utils.ParseCoin("0pool1"),
					},
				}
			},
			"amount must be positive value",
		},
		{
			"invalid winning bid records: auction id",
			func(genState *types.GenesisState) {
				genState.WinningBidRecords = []types.WinningBidRecord{
					{
						AuctionId: 0,
						WinningBid: types.Bid{
							PoolId: validPoolId,
							Bidder: validBidder.String(),
							Amount: utils.ParseCoin("1000000pool1"),
						},
					},
				}
			},
			"auction id must not be 0",
		},
		{
			"invalid winning bid records: pool id",
			func(genState *types.GenesisState) {
				genState.WinningBidRecords = []types.WinningBidRecord{
					{
						AuctionId: validAuctionId,
						WinningBid: types.Bid{
							PoolId: 0,
							Bidder: validBidder.String(),
							Amount: utils.ParseCoin("1000000pool1"),
						},
					},
				}
			},
			"invalid winning bid: pool id must not be 0",
		},
		{
			"invalid winning bid records: pool id",
			func(genState *types.GenesisState) {
				genState.WinningBidRecords = []types.WinningBidRecord{
					{
						AuctionId: validAuctionId,
						WinningBid: types.Bid{
							PoolId: validPoolId,
							Bidder: validBidder.String(),
							Amount: utils.ParseCoin("0pool1"),
						},
					},
				}
			},
			"invalid winning bid: amount must be positive value",
		},
		{
			"invalid winning bid records: duplicate winning bid",
			func(genState *types.GenesisState) {
				genState.WinningBidRecords = []types.WinningBidRecord{
					{
						AuctionId: validAuctionId,
						WinningBid: types.Bid{
							PoolId: validPoolId,
							Bidder: validBidder.String(),
							Amount: utils.ParseCoin("1000000pool1"),
						},
					},
					{
						AuctionId: validAuctionId,
						WinningBid: types.Bid{
							PoolId: validPoolId,
							Bidder: validBidder.String(),
							Amount: utils.ParseCoin("1000000pool1"),
						},
					},
				}
			},
			"multiple winning bids at auction 1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesis()
			genState.RewardsAuctions = []types.RewardsAuction{}
			genState.Bids = []types.Bid{}
			genState.WinningBidRecords = []types.WinningBidRecord{}
			tc.malleate(genState)
			err := genState.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
