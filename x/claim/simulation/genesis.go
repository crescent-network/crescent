package simulation

import (
	"math/rand"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/claim/types"
)

// GenAirdrops randomly generates airdrops.
func GenAirdrops(r *rand.Rand) (airdrops []types.Airdrop) {
	numAirdrops := r.Intn(5)
	airdrops = make([]types.Airdrop, numAirdrops)
	for i := 0; i < numAirdrops; i++ {
		conditions := []types.ConditionType{
			types.ConditionTypeDeposit,
			types.ConditionTypeSwap,
			types.ConditionTypeLiquidStake,
			types.ConditionTypeVote,
		}
		rand.Shuffle(len(conditions), func(i, j int) {
			conditions[i], conditions[j] = conditions[j], conditions[i]
		})
		numConditions := r.Intn(4) + 1
		conditions = conditions[:numConditions]
		sort.Slice(conditions, func(i, j int) bool {
			return conditions[i] < conditions[j]
		})
		airdrops[i] = types.Airdrop{
			Id:            uint64(i + 1),
			SourceAddress: utils.TestAddress(i).String(),
			Conditions:    conditions,
			StartTime:     utils.ParseTime("0001-01-01T00:00:00Z"),
			EndTime:       utils.ParseTime("9999-12-31T00:00:00Z"),
		}
	}
	return
}

// GenClaimRecords randomly generates claim records.
func GenClaimRecords(r *rand.Rand, accs []simtypes.Account, airdrops []types.Airdrop) (claimRecords []types.ClaimRecord) {
	if len(airdrops) == 0 {
		return nil
	}
	accs = utils.ShuffleSimAccounts(r, accs)
	numClaimRecords := r.Intn(len(accs)) + 1
	claimRecords = make([]types.ClaimRecord, numClaimRecords)
	for i := 0; i < numClaimRecords; i++ {
		airdrop := airdrops[r.Intn(len(airdrops))]
		claimedConditions := make([]types.ConditionType, len(airdrop.Conditions))
		copy(claimedConditions, airdrop.Conditions)
		rand.Shuffle(len(claimedConditions), func(i, j int) {
			claimedConditions[i], claimedConditions[j] = claimedConditions[j], claimedConditions[i]
		})
		numClaimedConditions := r.Intn(len(claimedConditions) + 1)
		claimedConditions = claimedConditions[:numClaimedConditions]
		initialClaimableCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1+r.Int63n(1000000)))
		claimableCoins := simtypes.RandSubsetCoins(r, initialClaimableCoins)
		claimRecords[i] = types.ClaimRecord{
			AirdropId:             airdrop.Id,
			Recipient:             accs[i].Address.String(),
			InitialClaimableCoins: initialClaimableCoins,
			ClaimableCoins:        claimableCoins,
			ClaimedConditions:     claimedConditions,
		}
	}
	return
}

// RandomizedGenState generates a random genesis state for the module.
func RandomizedGenState(simState *module.SimulationState) {
	var airdrops []types.Airdrop
	simState.AppParams.GetOrGenerate(
		simState.Cdc, "airdrops", &airdrops, simState.Rand,
		func(r *rand.Rand) { airdrops = GenAirdrops(r) },
	)

	var claimRecords []types.ClaimRecord
	simState.AppParams.GetOrGenerate(
		simState.Cdc, "claim_records", &claimRecords, simState.Rand,
		func(r *rand.Rand) { claimRecords = GenClaimRecords(r, simState.Accounts, airdrops) },
	)

	airdropBalances := map[uint64]sdk.Coins{} // airdrop id => balances
	for _, claimRecord := range claimRecords {
		airdropBalances[claimRecord.AirdropId] = airdropBalances[claimRecord.AirdropId].Add(claimRecord.ClaimableCoins...)
	}

	genState := &types.GenesisState{
		Airdrops:     airdrops,
		ClaimRecords: claimRecords,
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genState)

	// Modify x/bank module's genesis state.
	var bankGenState banktypes.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[banktypes.ModuleName], &bankGenState)
	addedSupply := sdk.Coins{}
	for _, airdrop := range airdrops {
		if balances, ok := airdropBalances[airdrop.Id]; ok {
			bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
				Address: airdrop.SourceAddress,
				Coins:   balances,
			})
			addedSupply = addedSupply.Add(balances...)
		}
	}
	bankGenState.Supply = bankGenState.Supply.Add(addedSupply...)
	bz := simState.Cdc.MustMarshalJSON(&bankGenState)
	simState.GenState[banktypes.ModuleName] = bz
}
