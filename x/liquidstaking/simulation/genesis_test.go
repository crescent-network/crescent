package simulation_test

//// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
//// Abnormal scenarios are not tested here.
//func TestRandomizedGenState(t *testing.T) {
//	interfaceRegistry := codectypes.NewInterfaceRegistry()
//	cdc := codec.NewProtoCodec(interfaceRegistry)
//	s := rand.NewSource(1)
//	r := rand.New(s)
//
//	simState := module.SimulationState{
//		AppParams:    make(simtypes.AppParams),
//		Cdc:          cdc,
//		Rand:         r,
//		NumBonded:    3,
//		Accounts:     simtypes.RandomAccounts(r, 3),
//		InitialStake: 1000,
//		GenState:     make(map[string]json.RawMessage),
//	}
//
//	simulation.RandomizedGenState(&simState)
//
//	var genState types.GenesisState
//	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &genState)
//
//	require.Equal(t, sdk.MustNewDecFromStr("0.3"), genState.Params.BiquidStakings[0].Rate)
//	require.Equal(t, "cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta", genState.Params.BiquidStakings[0].SourceAddress)
//	require.Equal(t, "cosmos1ke7rn6vl3vmeasmcrxdm3pfrt37fsg5jfrex80pp3hvhwgu4h4usxgvk3e", genState.Params.BiquidStakings[0].DestinationAddress)
//	require.Equal(t, uint32(9), genState.Params.EpochBlocks)
//}
//
//// TestRandomizedGenState tests abnormal scenarios of applying RandomizedGenState.
//func TestRandomizedGenState1(t *testing.T) {
//	interfaceRegistry := codectypes.NewInterfaceRegistry()
//	cdc := codec.NewProtoCodec(interfaceRegistry)
//
//	s := rand.NewSource(1)
//	r := rand.New(s)
//
//	// all these tests will panic
//	tests := []struct {
//		simState module.SimulationState
//		panicMsg string
//	}{
//		{ // panic => reason: incomplete initialization of the simState
//			module.SimulationState{}, "invalid memory address or nil pointer dereference"},
//		{ // panic => reason: incomplete initialization of the simState
//			module.SimulationState{
//				AppParams: make(simtypes.AppParams),
//				Cdc:       cdc,
//				Rand:      r,
//			}, "assignment to entry in nil map"},
//	}
//
//	for _, tt := range tests {
//		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
//	}
//}
