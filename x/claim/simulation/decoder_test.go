package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/kv"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/claim/simulation"
	"github.com/crescent-network/crescent/v2/x/claim/types"
)

func TestDecodeClaimStore(t *testing.T) {
	cdc := chain.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	airdrop := types.Airdrop{
		Id:            1,
		SourceAddress: utils.TestAddress(0).String(),
		Conditions: []types.ConditionType{
			types.ConditionTypeDeposit,
			types.ConditionTypeSwap,
			types.ConditionTypeLiquidStake,
			types.ConditionTypeVote,
		},
		StartTime: utils.ParseTime("2022-01-01T00:00:00Z"),
		EndTime:   utils.ParseTime("2023-01-01T00:00:00Z"),
	}
	claimRecord := types.ClaimRecord{
		AirdropId:             1,
		Recipient:             utils.TestAddress(1).String(),
		InitialClaimableCoins: utils.ParseCoins("1000000stake"),
		ClaimableCoins:        utils.ParseCoins("500000stake"),
		ClaimedConditions: []types.ConditionType{
			types.ConditionTypeLiquidStake,
			types.ConditionTypeVote,
		},
	}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.AirdropKeyPrefix, Value: cdc.MustMarshal(&airdrop)},
			{Key: types.ClaimRecordKeyPrefix, Value: cdc.MustMarshal(&claimRecord)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Airdrop", fmt.Sprintf("%v\n%v", airdrop, airdrop)},
		{"ClaimRecord", fmt.Sprintf("%v\n%v", claimRecord, claimRecord)},
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
