package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestPool_DenomIn(t *testing.T) {
	pool := types.NewPool(1, 2, "ucre", "uusd", 10, sdk.NewDec(1), sdk.NewDec(1))
	require.Equal(t, "ucre", pool.DenomIn(true))
	require.Equal(t, "uusd", pool.DenomIn(false))
}

func TestPool_DenomOut(t *testing.T) {
	pool := types.NewPool(1, 2, "ucre", "uusd", 10, sdk.NewDec(1), sdk.NewDec(1))
	require.Equal(t, "uusd", pool.DenomOut(true))
	require.Equal(t, "ucre", pool.DenomOut(false))
}

func TestPool_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(pool *types.Pool)
		expectedErr string
	}{
		{
			"valid",
			func(pool *types.Pool) {},
			"",
		},
		{
			"invalid id",
			func(pool *types.Pool) {
				pool.Id = 0
			},
			"id must not be 0",
		},
		{
			"invalid market id",
			func(pool *types.Pool) {
				pool.MarketId = 0
			},
			"market id must not be 0",
		},
		{
			"invalid denom 0",
			func(pool *types.Pool) {
				pool.Denom0 = "invaliddenom!"
			},
			"invalid denom 0: invalid denom: invaliddenom!",
		},
		{
			"invalid denom 1",
			func(pool *types.Pool) {
				pool.Denom1 = "invaliddenom!"
			},
			"invalid denom 1: invalid denom: invaliddenom!",
		},
		{
			"same denom 0 and denom 1",
			func(pool *types.Pool) {
				pool.Denom0 = "ucre"
				pool.Denom1 = "ucre"
			},
			"denom 0 and denom 1 must not be same: ucre",
		},
		{
			"invalid reserve address",
			func(pool *types.Pool) {
				pool.ReserveAddress = "invalidaddr"
			},
			"invalid reserve address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid rewards pool",
			func(pool *types.Pool) {
				pool.RewardsPool = "invalidaddr"
			},
			"invalid rewards pool: decoding bech32 failed: invalid separator index -1",
		},
		{
			"not allowed tick spacing",
			func(pool *types.Pool) {
				pool.TickSpacing = 7
			},
			"tick spacing 7 is not allowed",
		},
		{
			"negative min order quantity",
			func(pool *types.Pool) {
				pool.MinOrderQuantity = sdk.NewDec(-1000)
			},
			"min order quantity must not be negative: -1000.000000000000000000",
		},
		{
			"negative min order quote",
			func(pool *types.Pool) {
				pool.MinOrderQuote = sdk.NewDec(-1000)
			},
			"min order quote must not be negative: -1000.000000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := types.NewPool(1, 2, "ucre", "uusd", 10, sdk.NewDec(1), sdk.NewDec(1))
			tc.malleate(&pool)
			err := pool.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestPoolState_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(poolState *types.PoolState)
		expectedErr string
	}{
		{
			"valid",
			func(poolState *types.PoolState) {},
			"",
		},
		{
			"invalid current sqrt price",
			func(poolState *types.PoolState) {
				poolState.CurrentSqrtPrice = utils.ParseDec("0")
			},
			"current sqrt price must be positive: 0.000000000000000000",
		},
		{
			"invalid current sqrt price 2",
			func(poolState *types.PoolState) {
				poolState.CurrentSqrtPrice = utils.ParseDec("-1.0")
			},
			"current sqrt price must be positive: -1.000000000000000000",
		},
		{
			"invalid current liquidity",
			func(poolState *types.PoolState) {
				poolState.CurrentLiquidity = sdk.NewInt(-1000_000000)
			},
			"current liquidity must not be negative: -1000000000",
		},
		{
			"invalid total liquidity",
			func(poolState *types.PoolState) {
				poolState.TotalLiquidity = sdk.NewInt(-1000)
			},
			"total liquidity must not be negative: -1000",
		},
		{
			"invalid fee growth global",
			func(poolState *types.PoolState) {
				poolState.FeeGrowthGlobal = sdk.DecCoins{sdk.NewInt64DecCoin("ucre", 0)}
			},
			"invalid fee growth global: coin 0.000000000000000000ucre amount is not positive",
		},
		{
			"wrong fee growth global coins number",
			func(poolState *types.PoolState) {
				poolState.FeeGrowthGlobal = utils.ParseDecCoins("0.00001ucre,0.00001uusd,0.0001uatom")
			},
			"number of coins in fee growth global must not be higher than 2: 3",
		},
		{
			"invalid farming rewards growth global",
			func(poolState *types.PoolState) {
				poolState.FarmingRewardsGrowthGlobal = sdk.DecCoins{sdk.NewInt64DecCoin("uatom", 0)}
			},
			"invalid farming rewards growth global: coin 0.000000000000000000uatom amount is not positive",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := types.PoolState{
				CurrentTick:                -10,
				CurrentSqrtPrice:           utils.ParseDec("0.99991100001"),
				CurrentLiquidity:           sdk.NewInt(1000_000000),
				TotalLiquidity:             sdk.NewInt(2000_000000),
				FeeGrowthGlobal:            utils.ParseDecCoins("0.0001ucre,0.0001uusd"),
				FarmingRewardsGrowthGlobal: utils.ParseDecCoins("0.0001uatom,0.0001stake"),
			}
			tc.malleate(&pool)
			err := pool.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
