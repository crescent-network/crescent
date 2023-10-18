package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestPosition_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(position *types.Position)
		expectedErr string
	}{
		{
			"valid",
			func(position *types.Position) {},
			"",
		},
		{
			"invalid id",
			func(position *types.Position) {
				position.Id = 0
			},
			"id must not be 0",
		},
		{
			"invalid pool id",
			func(position *types.Position) {
				position.PoolId = 0
			},
			"pool id must not be 0",
		},
		{
			"invalid owner address",
			func(position *types.Position) {
				position.Owner = "invalidaddr"
			},
			"invalid owner address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid liquidity",
			func(position *types.Position) {
				position.Liquidity = sdk.NewInt(-1000_000000)
			},
			"liquidity must not be negative: -1000000000",
		},
		{
			"invalid last fee growth inside",
			func(position *types.Position) {
				position.LastFeeGrowthInside = sdk.DecCoins{sdk.NewInt64DecCoin("ucre", 0)}
			},
			"invalid last fee growth inside: coin 0.000000000000000000ucre amount is not positive",
		},
		{
			"wrong last fee growth inside coins number",
			func(position *types.Position) {
				position.LastFeeGrowthInside = utils.ParseDecCoins("0.0001ucre,0.0001uusd,0.0001uatom")
			},
			"number of coins in last fee growth inside must not be higher than 2: 3",
		},
		{
			"invalid owed fee",
			func(position *types.Position) {
				position.OwedFee = sdk.Coins{sdk.NewInt64Coin("ucre", 0)}
			},
			"invalid owed fee: coin 0ucre amount is not positive",
		},
		{
			"invalid owed fee coins number",
			func(position *types.Position) {
				position.OwedFee = utils.ParseCoins("10_000000ucre,50_000000uusd,10_000000uatom")
			},
			"number of coins in owed fee must not be higher than 2: 3",
		},
		{
			"invalid last farming rewards growth inside",
			func(position *types.Position) {
				position.LastFarmingRewardsGrowthInside = sdk.DecCoins{sdk.NewInt64DecCoin("ucre", 0)}
			},
			"invalid last farming rewards growth inside: coin 0.000000000000000000ucre amount is not positive",
		},
		{
			"invalid owed farming rewards",
			func(position *types.Position) {
				position.OwedFarmingRewards = sdk.Coins{sdk.NewInt64Coin("ucre", 0)}
			},
			"invalid owed farming rewards: coin 0ucre amount is not positive",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			position := types.NewPosition(1, 2, utils.TestAddress(1), -500, 500)
			tc.malleate(&position)
			err := position.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestPosition_MustGetAddress(t *testing.T) {
	position := types.NewPosition(1, 2, utils.TestAddress(1), -500, 500)
	require.Equal(t, position.Owner, position.MustGetOwnerAddress().String())
}
