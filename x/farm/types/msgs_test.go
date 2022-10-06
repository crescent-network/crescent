package types_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func TestMsgCreatePrivatePlan(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCreatePrivatePlan)
		expectedErr string // empty means no error
	}{
		{
			"happy case",
			func(msg *types.MsgCreatePrivatePlan) {},
			"",
		},
		{
			"invalid creator",
			func(msg *types.MsgCreatePrivatePlan) {
				msg.Creator = "invalidaddr"
			},
			"invalid creator address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"too long description",
			func(msg *types.MsgCreatePrivatePlan) {
				msg.Description = strings.Repeat("x", types.MaxPlanDescriptionLen+1)
			},
			"too long plan description, maximum 200: invalid request",
		},
		{
			"empty reward allocations",
			func(msg *types.MsgCreatePrivatePlan) {
				msg.RewardAllocations = []types.RewardAllocation{}
			},
			"invalid reward allocations: empty reward allocations: invalid request",
		},
		{
			"zero pair id",
			func(msg *types.MsgCreatePrivatePlan) {
				msg.RewardAllocations = []types.RewardAllocation{
					{
						PairId:        0,
						RewardsPerDay: utils.ParseCoins("1_000000stake"),
					},
				}
			},
			"invalid reward allocations: pair id must not be zero: invalid request",
		},
		{
			"invalid rewards per day",
			func(msg *types.MsgCreatePrivatePlan) {
				msg.RewardAllocations = []types.RewardAllocation{
					{
						PairId:        1,
						RewardsPerDay: sdk.Coins{utils.ParseCoin("0stake")},
					},
				}
			},
			"invalid reward allocations: invalid rewards per day: coin 0stake amount is not positive: invalid request",
		},
		{
			"duplicate pair id",
			func(msg *types.MsgCreatePrivatePlan) {
				msg.RewardAllocations = []types.RewardAllocation{
					{
						PairId:        1,
						RewardsPerDay: sdk.Coins{utils.ParseCoin("100_000000stake")},
					},
					{
						PairId:        1,
						RewardsPerDay: sdk.Coins{utils.ParseCoin("200_000000stake")},
					},
				}
			},
			"invalid reward allocations: duplicate pair id: 1: invalid request",
		},
		{
			"invalid start/end time",
			func(msg *types.MsgCreatePrivatePlan) {
				msg.StartTime = utils.ParseTime("2023-01-01T00:00:00Z")
				msg.EndTime = utils.ParseTime("2023-01-01T00:00:00Z")
			},
			"end time must be after start time: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgCreatePrivatePlan(
				utils.TestAddress(0), "Farming Plan",
				[]types.RewardAllocation{
					{
						PairId:        1,
						RewardsPerDay: utils.ParseCoins("100_000000stake"),
					},
					{
						PairId:        2,
						RewardsPerDay: utils.ParseCoins("200_000000stake"),
					},
				},
				utils.ParseTime("2022-01-01T00:00:00Z"),
				utils.ParseTime("2023-01-01T00:00:00Z"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgCreatePrivatePlan, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetCreatorAddress(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgFarm(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgFarm)
		expectedErr string // empty means no error
	}{
		{
			"happy case",
			func(msg *types.MsgFarm) {},
			"",
		},
		{
			"invalid farmer",
			func(msg *types.MsgFarm) {
				msg.Farmer = "invalidaddr"
			},
			"invalid farmer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"negative coin",
			func(msg *types.MsgFarm) {
				msg.Coin = sdk.Coin{Denom: "pool1", Amount: sdk.NewInt(-1)}
			},
			"invalid coin: negative coin amount: -1: invalid request",
		},
		{
			"zero coin",
			func(msg *types.MsgFarm) {
				msg.Coin = utils.ParseCoin("0pool1")
			},
			"non-positive coin: 0pool1: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgFarm(utils.TestAddress(0), utils.ParseCoin("1000_000000pool1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgFarm, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetFarmerAddress(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgUnfarm(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgUnfarm)
		expectedErr string // empty means no error
	}{
		{
			"happy case",
			func(msg *types.MsgUnfarm) {},
			"",
		},
		{
			"invalid farmer",
			func(msg *types.MsgUnfarm) {
				msg.Farmer = "invalidaddr"
			},
			"invalid farmer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"negative bond",
			func(msg *types.MsgUnfarm) {
				msg.Coin = sdk.Coin{Denom: "pool1", Amount: sdk.NewInt(-1)}
			},
			"invalid coin: negative coin amount: -1: invalid request",
		},
		{
			"zero bond",
			func(msg *types.MsgUnfarm) {
				msg.Coin = utils.ParseCoin("0pool1")
			},
			"non-positive coin: 0pool1: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgUnfarm(utils.TestAddress(0), utils.ParseCoin("1000_000000pool1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgUnfarm, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetFarmerAddress(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgHarvest(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgHarvest)
		expectedErr string // empty means no error
	}{
		{
			"happy case",
			func(msg *types.MsgHarvest) {},
			"",
		},
		{
			"invalid farmer",
			func(msg *types.MsgHarvest) {
				msg.Farmer = "invalidaddr"
			},
			"invalid farmer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid denom",
			func(msg *types.MsgHarvest) {
				msg.Denom = "invalid!"
			},
			"invalid denom: invalid denom: invalid!: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgHarvest(utils.TestAddress(0), "pool1")
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgHarvest, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetFarmerAddress(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
