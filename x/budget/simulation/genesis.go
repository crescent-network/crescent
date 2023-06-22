package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

// DONTCOVER

// Simulation parameter constants
const (
	EpochBlocks = "epoch_blocks"
	Budgets     = "budgets"
)

// GenEpochBlocks returns randomized epoch blocks.
func GenEpochBlocks(r *rand.Rand) uint32 {
	return uint32(simtypes.RandIntBetween(r, int(types.DefaultEpochBlocks), 10))
}

// GenBudgets returns randomized budgets.
func GenBudgets(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) []types.Budget {
	ranBudgets := make([]types.Budget, 0)

	for i := 0; i < simtypes.RandIntBetween(r, 1, 20); i++ {
		var sourceAddr, destAddr sdk.AccAddress
		if i%2 == 1 {
			sourceAddr = types.DeriveAddress(types.AddressType20Bytes, "", "fee_collector")
		} else {
			simAccount, _ := simtypes.RandomAcc(r, accs)
			sourceAddr = simAccount.Address
		}
		for {
			simAccount, _ := simtypes.RandomAcc(r, accs)
			if !simAccount.Address.Equals(sourceAddr) {
				destAddr = simAccount.Address
				break
			}
		}

		budget := types.Budget{
			Name:               "simulation-test-" + simtypes.RandStringOfLength(r, 5),
			Rate:               sdk.NewDecFromIntWithPrec(sdk.NewInt(int64(simtypes.RandIntBetween(r, 1, 5))), 2), // 1~5%
			SourceAddress:      sourceAddr.String(),                                                               // Cosmos Hub's FeeCollector module account
			DestinationAddress: destAddr.String(),
			StartTime:          ctx.BlockTime(),
			EndTime:            ctx.BlockTime().AddDate(0, 0, simtypes.RandIntBetween(r, 1, 28)),
		}
		ranBudgets = append(ranBudgets, budget)
	}

	return ranBudgets
}

func InitBudgets(r *rand.Rand) []types.Budget {
	return []types.Budget{{
		Name:               "simulation-test-" + simtypes.RandStringOfLength(r, 5),
		Rate:               sdk.NewDecFromIntWithPrec(sdk.NewInt(int64(simtypes.RandIntBetween(r, 1, 4))), 1), // 10~30%
		SourceAddress:      types.DeriveAddress(types.AddressType20Bytes, "", "fee_collector").String(),       // Cosmos Hub's FeeCollector module account
		DestinationAddress: sdk.AccAddress(address.Module(types.ModuleName, []byte("GravityDEXFarmingBudget"))).String(),
		StartTime:          types.MustParseRFC3339("2000-01-01T00:00:00Z"),
		EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
	}}
}

// RandomizedGenState generates a random GenesisState for budget.
func RandomizedGenState(simState *module.SimulationState) {
	var epochBlocks uint32
	var budgets []types.Budget
	simState.AppParams.GetOrGenerate(
		simState.Cdc, EpochBlocks, &epochBlocks, simState.Rand,
		func(r *rand.Rand) { epochBlocks = GenEpochBlocks(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, Budgets, &budgets, simState.Rand,
		func(r *rand.Rand) { budgets = InitBudgets(r) },
	)

	budgetGenesis := types.GenesisState{
		Params: types.Params{
			EpochBlocks: epochBlocks,
			Budgets:     budgets,
		},
	}

	bz, _ := json.MarshalIndent(&budgetGenesis, "", " ")
	fmt.Printf("Selected randomly generated budget parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&budgetGenesis)
}
