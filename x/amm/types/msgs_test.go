package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestMsgCreatePool_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCreatePool)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgCreatePool) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgCreatePool) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgCreatePool) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"invalid price",
			func(msg *types.MsgCreatePool) {
				msg.Price = utils.ParseDec("0")
			},
			"price must be positive: 0.000000000000000000: invalid request",
		},
		{
			"invalid price 2",
			func(msg *types.MsgCreatePool) {
				msg.Price = utils.ParseDec("-1.0")
			},
			"price must be positive: -1.000000000000000000: invalid request",
		},
		{
			"invalid price 3",
			func(msg *types.MsgCreatePool) {
				msg.Price = utils.ParseDec("0.000000000000000001")
			},
			"price is lower than the min price 0.000000000000010000: invalid request",
		},
		{
			"invalid price 4",
			func(msg *types.MsgCreatePool) {
				msg.Price = utils.ParseDec("100000000000000000000000000000000000000000")
			},
			"price is higher than the max price 1000000000000000000000000.000000000000000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgCreatePool(senderAddr, 1, utils.ParseDec("12.3456789"))
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgCreatePool, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgAddLiquidity_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgAddLiquidity)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgAddLiquidity) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgAddLiquidity) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid pool id",
			func(msg *types.MsgAddLiquidity) {
				msg.PoolId = 0
			},
			"pool id must not be 0: invalid request",
		},
		{
			"invalid lower price",
			func(msg *types.MsgAddLiquidity) {
				msg.LowerPrice = utils.ParseDec("0")
			},
			"lower price must be positive: 0.000000000000000000: invalid request",
		},
		{
			"invalid lower price 2",
			func(msg *types.MsgAddLiquidity) {
				msg.LowerPrice = utils.ParseDec("-1.0")
			},
			"lower price must be positive: -1.000000000000000000: invalid request",
		},
		{
			"too low lower price",
			func(msg *types.MsgAddLiquidity) {
				msg.LowerPrice = utils.ParseDec("0.000000000000000001")
			},
			"lower tick must not be lower than the minimum -1260000: invalid request",
		},
		{
			"invalid upper price",
			func(msg *types.MsgAddLiquidity) {
				msg.UpperPrice = utils.ParseDec("0")
			},
			"upper price must be positive: 0.000000000000000000: invalid request",
		},
		{
			"invalid upper price 2",
			func(msg *types.MsgAddLiquidity) {
				msg.UpperPrice = utils.ParseDec("-1")
			},
			"upper price must be positive: -1.000000000000000000: invalid request",
		},
		{
			"too high upper price",
			func(msg *types.MsgAddLiquidity) {
				msg.UpperPrice = utils.ParseDec("100000000000000000000000000000000000000000")
			},
			"upper tick must not be higher than the maximum 2160000: invalid request",
		},
		{
			"invalid desired amount",
			func(msg *types.MsgAddLiquidity) {
				msg.DesiredAmount = sdk.Coins{sdk.NewInt64Coin("ucre", 0), sdk.NewInt64Coin("uusd", 0)}
			},
			"invalid desired amount: coin 0ucre amount is not positive: invalid coins",
		},
		{
			"invalid desired amount 1",
			func(msg *types.MsgAddLiquidity) {
				msg.DesiredAmount = sdk.NewCoins()
			},
			"invalid desired amount length: 0: invalid request",
		},
		{
			"invalid desired amount 2",
			func(msg *types.MsgAddLiquidity) {
				msg.DesiredAmount = utils.ParseCoins("100_000000ucre,500_000000uusd,100_000000uatom")
			},
			"invalid desired amount length: 3: invalid request",
		},
		{
			"single asset",
			func(msg *types.MsgAddLiquidity) {
				msg.DesiredAmount = utils.ParseCoins("100_000000ucre")
			},
			"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgAddLiquidity(
				senderAddr, 1,
				utils.ParseDec("4.5"), utils.ParseDec("5.5"),
				utils.ParseCoins("100_000000ucre,500_000000uusd"))
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgAddLiquidity, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgRemoveLiquidity_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgRemoveLiquidity)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgRemoveLiquidity) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgRemoveLiquidity) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid position id",
			func(msg *types.MsgRemoveLiquidity) {
				msg.PositionId = 0
			},
			"position is must not be 0: invalid request",
		},
		{
			"invalid liquidity",
			func(msg *types.MsgRemoveLiquidity) {
				msg.Liquidity = sdk.NewInt(0)
			},
			"liquidity must be positive: 0: invalid request",
		},
		{
			"invalid liquidity 2",
			func(msg *types.MsgRemoveLiquidity) {
				msg.Liquidity = sdk.NewInt(-1000_000000)
			},
			"liquidity must be positive: -1000000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgRemoveLiquidity(senderAddr, 1, sdk.NewInt(1000_000000))
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgRemoveLiquidity, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgCollect_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCollect)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgCollect) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgCollect) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid position id",
			func(msg *types.MsgCollect) {
				msg.PositionId = 0
			},
			"position is must not be 0: invalid request",
		},
		{
			"invalid amount",
			func(msg *types.MsgCollect) {
				msg.Amount = sdk.NewCoins()
			},
			"amount must be all positive: : invalid coins",
		},
		{
			"invalid amount 2",
			func(msg *types.MsgCollect) {
				msg.Amount = sdk.Coins{sdk.NewInt64Coin("ucre", 0)}
			},
			"invalid amount: coin 0ucre amount is not positive: invalid coins",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgCollect(senderAddr, 1, utils.ParseCoins("1_000000ucre,10_000000uatom"))
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgCollect, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgCreatePrivateFarmingPlan_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCreatePrivateFarmingPlan)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgCreatePrivateFarmingPlan) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgCreatePrivateFarmingPlan) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		// the rest of the checks done performed in TestFarmingPlan_Validate
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgCreatePrivateFarmingPlan(
				senderAddr, "Farming plan", utils.TestAddress(100),
				[]types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000ucre")),
				},
				utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2023-07-01T00:00:00Z"))
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgCreatePrivateFarmingPlan, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgTerminatePrivateFarmingPlan_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgTerminatePrivateFarmingPlan)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgTerminatePrivateFarmingPlan) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgTerminatePrivateFarmingPlan) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid farming plan id",
			func(msg *types.MsgTerminatePrivateFarmingPlan) {
				msg.FarmingPlanId = 0
			},
			"farming plan id must not be 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgTerminatePrivateFarmingPlan(senderAddr, 1)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgTerminatePrivateFarmingPlan, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
