package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v4/x/liquidfarming/simulation"
	"github.com/crescent-network/crescent/v4/x/liquidfarming/types"
)

func TestDecodeFarmingStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	liquidFarm := types.LiquidFarm{}
	compoundingRewards := types.CompoundingRewards{}
	rewardsAuction := types.RewardsAuction{}
	bid := types.Bid{}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.LiquidFarmKeyPrefix, Value: cdc.MustMarshal(&liquidFarm)},
			{Key: types.CompoundingRewardsKeyPrefix, Value: cdc.MustMarshal(&compoundingRewards)},
			{Key: types.RewardsAuctionKeyPrefix, Value: cdc.MustMarshal(&rewardsAuction)},
			{Key: types.BidKeyPrefix, Value: cdc.MustMarshal(&bid)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"LiquidFarm", fmt.Sprintf("%v\n%v", liquidFarm, liquidFarm)},
		{"CompoundingRewards", fmt.Sprintf("%v\n%v", compoundingRewards, compoundingRewards)},
		{"RewardsAuction", fmt.Sprintf("%v\n%v", rewardsAuction, rewardsAuction)},
		{"Bid", fmt.Sprintf("%v\n%v", bid, bid)},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
