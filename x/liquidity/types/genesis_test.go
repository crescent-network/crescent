package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func TestGenesisState_Validate(t *testing.T) {
	// Valid structs.
	pair := types.NewPair(1, "denom1", "denom2")
	pool := types.NewBasicPool(1, 1, testAddr)
	depositReq := types.DepositRequest{
		Id:             1,
		PoolId:         1,
		MsgHeight:      1,
		Depositor:      sdk.AccAddress(crypto.AddressHash([]byte("depositor"))).String(),
		DepositCoins:   utils.ParseCoins("1000000denom1,1000000denom2"),
		MintedPoolCoin: sdk.NewInt64Coin("pool1", 0),
		Status:         types.RequestStatusNotExecuted,
	}
	withdrawReq := types.WithdrawRequest{
		Id:         1,
		PoolId:     1,
		MsgHeight:  1,
		Withdrawer: sdk.AccAddress(crypto.AddressHash([]byte("withdrawer"))).String(),
		PoolCoin:   sdk.NewInt64Coin("pool1", 1000000),
		Status:     types.RequestStatusNotExecuted,
	}
	order := types.Order{
		Id:                 1,
		PairId:             1,
		MsgHeight:          1,
		Orderer:            sdk.AccAddress(crypto.AddressHash([]byte("orderer"))).String(),
		Direction:          types.OrderDirectionSell,
		OfferCoin:          sdk.NewInt64Coin("denom1", 1000000),
		RemainingOfferCoin: sdk.NewInt64Coin("denom1", 500000),
		ReceivedCoin:       sdk.NewInt64Coin("denom2", 500000),
		Price:              utils.ParseDec("1.0"),
		Amount:             sdk.NewInt(1000000),
		OpenAmount:         sdk.NewInt(500000),
		BatchId:            1,
		ExpireAt:           utils.ParseTime("2022-02-01T00:00:00Z"),
		Status:             types.OrderStatusPartiallyMatched,
	}

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
			"invalid params",
			func(genState *types.GenesisState) {
				genState.Params = types.Params{}
			},
			"invalid params: batch size must be positive: 0",
		},
		{
			"invalid pair",
			func(genState *types.GenesisState) {
				genState.Pairs[0] = types.Pair{}
			},
			"invalid pair at index 0: pair id must not be 0",
		},
		{
			"wrong pair id",
			func(genState *types.GenesisState) {
				genState.Pairs[0].Id = 2
			},
			"pair at index 0 has an id greater than last pair id: 2",
		},
		{
			"duplicate pair",
			func(genState *types.GenesisState) {
				genState.Pairs = []types.Pair{pair, pair}
			},
			"pair at index 1 has a duplicate id: 1",
		},
		{
			"invalid pool",
			func(genState *types.GenesisState) {
				genState.Pools[0] = types.Pool{}
			},
			"invalid pool at index 0: pool id must not be 0",
		},
		{
			"wrong pool id",
			func(genState *types.GenesisState) {
				genState.Pools[0].Id = 2
			},
			"pool at index 0 has an id greater than last pool id: 2",
		},
		{
			"duplicate pool",
			func(genState *types.GenesisState) {
				genState.Pools = []types.Pool{pool, pool}
			},
			"pool at index 1 has a duplicate pool id: 1",
		},
		{
			"unknown pair id",
			func(genState *types.GenesisState) {
				genState.Pools[0].PairId = 2
			},
			"pool at index 0 has unknown pair id: 2",
		},
		{
			"invalid deposit request",
			func(genState *types.GenesisState) {
				genState.DepositRequests[0] = types.DepositRequest{}
			},
			"invalid deposit request at index 0: id must not be 0",
		},
		{
			"unknown pool",
			func(genState *types.GenesisState) {
				genState.DepositRequests[0].PoolId = 2
			},
			"deposit request at index 0 has unknown pool id: 2",
		},
		{
			"wrong minted pool coin",
			func(genState *types.GenesisState) {
				genState.DepositRequests[0].MintedPoolCoin.Denom = "pool2"
			},
			"deposit request at index 0 has wrong minted pool coin: 0pool2",
		},
		{
			"wrong deposit coins",
			func(genState *types.GenesisState) {
				genState.DepositRequests[0].DepositCoins = utils.ParseCoins("1000000denom1,1000000denom3")
			},
			"deposit request at index 0 has wrong deposit coins: 1000000denom1,1000000denom3",
		},
		{
			"duplicate deposit request",
			func(genState *types.GenesisState) {
				genState.DepositRequests = []types.DepositRequest{depositReq, depositReq}
			},
			"deposit request at index 1 has a duplicate id: 1",
		},
		{
			"invalid withdraw request",
			func(genState *types.GenesisState) {
				genState.WithdrawRequests[0] = types.WithdrawRequest{}
			},
			"invalid withdraw request at index 0: id must not be 0",
		},
		{
			"unknown pool",
			func(genState *types.GenesisState) {
				genState.WithdrawRequests[0].PoolId = 2
			},
			"withdraw request at index 0 has unknown pool id: 2",
		},
		{
			"wrong pool coin denom",
			func(genState *types.GenesisState) {
				genState.WithdrawRequests[0].PoolCoin.Denom = "pool2"
			},
			"withdraw request at index 0 has wrong pool coin: 1000000pool2",
		},
		{
			"duplicate withdraw request",
			func(genState *types.GenesisState) {
				genState.WithdrawRequests = []types.WithdrawRequest{withdrawReq, withdrawReq}
			},
			"withdraw request at index 1 has a duplicate id: 1",
		},
		{
			"invalid order",
			func(genState *types.GenesisState) {
				genState.Orders[0] = types.Order{}
			},
			"invalid order at index 0: id must not be 0",
		},
		{
			"unknown pair",
			func(genState *types.GenesisState) {
				genState.Orders[0].PairId = 2
			},
			"order at index 0 has unknown pair id: 2",
		},
		{
			"wrong batch id",
			func(genState *types.GenesisState) {
				genState.Orders[0].BatchId = 2
			},
			"order at index 0 has a batch id greater than its pair's current batch id: 2",
		},
		{
			"wrong offer coin denom",
			func(genState *types.GenesisState) {
				genState.Orders[0].OfferCoin.Denom = "denom3"
				genState.Orders[0].RemainingOfferCoin.Denom = "denom3"
			},
			"order at index 0 has wrong offer coin denom: denom3 != denom1",
		},
		{
			"wrong demand coin denom",
			func(genState *types.GenesisState) {
				genState.Orders[0].ReceivedCoin.Denom = "denom4"
			},
			"order at index 0 has wrong demand coin denom: denom1 != denom2",
		},
		{
			"duplicate order",
			func(genState *types.GenesisState) {
				genState.Orders = []types.Order{order, order}
			},
			"order at index 1 has a duplicate id: 1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesis()
			genState.Pairs = []types.Pair{pair}
			genState.LastPairId = 1
			genState.Pools = []types.Pool{pool}
			genState.LastPoolId = 1
			genState.DepositRequests = []types.DepositRequest{depositReq}
			genState.WithdrawRequests = []types.WithdrawRequest{withdrawReq}
			genState.Orders = []types.Order{order}
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
