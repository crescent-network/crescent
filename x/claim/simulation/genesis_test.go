package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banksimulation "github.com/cosmos/cosmos-sdk/x/bank/simulation"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/claim/simulation"
	"github.com/crescent-network/crescent/v2/x/claim/types"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abnormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: 1000,
		GenState:     make(map[string]json.RawMessage),
	}

	banksimulation.RandomizedGenState(&simState)
	simulation.RandomizedGenState(&simState)

	var genState types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &genState)

	require.Len(t, genState.Airdrops, 4)
	require.Equal(t, uint64(1), genState.Airdrops[0].Id)
	require.Equal(t, "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a", genState.Airdrops[0].SourceAddress)
	require.Equal(t, []types.ConditionType{
		types.ConditionTypeDeposit,
		types.ConditionTypeSwap,
		types.ConditionTypeLiquidStake,
		types.ConditionTypeVote,
	}, genState.Airdrops[0].Conditions)
	require.Equal(t, utils.ParseTime("0001-01-01T00:00:00Z"), genState.Airdrops[0].StartTime)
	require.Equal(t, utils.ParseTime("9999-12-31T00:00:00Z"), genState.Airdrops[0].EndTime)
	require.Len(t, genState.ClaimRecords, 3)
	require.Equal(t, uint64(2), genState.ClaimRecords[0].AirdropId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", genState.ClaimRecords[0].Recipient)
	require.Equal(t, "778274stake", genState.ClaimRecords[0].InitialClaimableCoins.String())
	require.Equal(t, "500946stake", genState.ClaimRecords[0].ClaimableCoins.String())
	require.Equal(t, []types.ConditionType{types.ConditionTypeLiquidStake, types.ConditionTypeVote}, genState.ClaimRecords[0].ClaimedConditions)
}

// TestRandomizedGenState tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenState1(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)

	// all these tests will panic
	tests := []struct {
		simState module.SimulationState
		panicMsg string
	}{
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{}, "invalid memory address or nil pointer dereference"},
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{
				AppParams: make(simtypes.AppParams),
				Cdc:       cdc,
				Rand:      r,
			}, "assignment to entry in nil map"},
	}

	for _, tt := range tests {
		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
	}
}
