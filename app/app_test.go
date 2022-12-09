package app

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/tests/mocks"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	ica "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v3/modules/core"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/budget/x/budget"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/crescent-network/crescent/v3/x/claim"
	"github.com/crescent-network/crescent/v3/x/farming"
	"github.com/crescent-network/crescent/v3/x/liquidfarming"
	"github.com/crescent-network/crescent/v3/x/liquidity"
	"github.com/crescent-network/crescent/v3/x/liquidstaking"
	"github.com/crescent-network/crescent/v3/x/lpfarm"
	"github.com/crescent-network/crescent/v3/x/marketmaker"
	"github.com/crescent-network/crescent/v3/x/mint"
)

func TestAppExportAndBlockedAddrs(t *testing.T) {
	db := dbm.NewMemDB()
	encCfg := MakeTestEncodingConfig()
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	app := NewApp(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encCfg, EmptyAppOptions{}, wasm.EnableAllProposals, EmptyWasmOpts)

	for acc := range maccPerms {
		require.True(t, app.BankKeeper.BlockedAddr(app.AccountKeeper.GetModuleAddress(acc)), "ensure that blocked addresses are properly set in bank keeper")
	}

	genesisState := NewDefaultGenesisState(encCfg.Marshaler)
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewApp(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encCfg, EmptyAppOptions{}, wasm.EnableAllProposals, EmptyWasmOpts)
	res, err := app2.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")

	_, err = res.AppState.MarshalJSON()
	require.NoError(t, err)
}

func TestRunMigrations(t *testing.T) {
	db := dbm.NewMemDB()
	encCfg := MakeTestEncodingConfig()
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	app := NewApp(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encCfg, EmptyAppOptions{}, wasm.EnableAllProposals, EmptyWasmOpts)

	// Create a new baseapp and configurator for the purpose of this test.
	bApp := baseapp.NewBaseApp(AppName, logger, db, encCfg.TxConfig.TxDecoder())
	bApp.SetCommitMultiStoreTracer(nil)
	bApp.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	app.BaseApp = bApp
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())

	// We register all modules on the Configurator, except x/bank. x/bank will
	// serve as the test subject on which we run the migration tests.
	//
	// The loop below is the same as calling `RegisterServices` on
	// ModuleManager, except that we skip x/bank.
	for _, module := range app.mm.Modules {
		if module.Name() == banktypes.ModuleName {
			continue
		}

		module.RegisterServices(app.configurator)
	}

	// Initialize the chain
	app.InitChain(abci.RequestInitChain{})
	app.Commit()

	testCases := []struct {
		name         string
		moduleName   string
		forVersion   uint64
		expRegErr    bool // errors while registering migration
		expRegErrMsg string
		expRunErr    bool // errors while running migration
		expRunErrMsg string
		expCalled    int
	}{
		{
			"cannot register migration for version 0",
			"bank", 0,
			true, "module migration versions should start at 1: invalid version", false, "", 0,
		},
		{
			"throws error on RunMigrations if no migration registered for bank",
			"", 1,
			false, "", true, "no migrations found for module bank: not found", 0,
		},
		{
			"can register and run migration handler for x/bank",
			"bank", 1,
			false, "", false, "", 1,
		},
		{
			"cannot register migration handler for same module & forVersion",
			"bank", 1,
			true, "another migration for module bank and version 1 already exists: internal logic error", false, "", 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			// Since it's very hard to test actual in-place store migrations in
			// tests (due to the difficulty of maintaining multiple versions of a
			// module), we're just testing here that the migration logic is
			// called.
			called := 0

			if tc.moduleName != "" {
				// Register migration for module from version `forVersion` to `forVersion+1`.
				err = app.configurator.RegisterMigration(tc.moduleName, tc.forVersion, func(sdk.Context) error {
					called++

					return nil
				})

				if tc.expRegErr {
					require.EqualError(t, err, tc.expRegErrMsg)

					return
				}
			}
			require.NoError(t, err)

			// Run migrations only for bank. That's why we put the initial
			// version for bank as 1, and for all other modules, we put as
			// their latest ConsensusVersion.
			_, err = app.mm.RunMigrations(
				app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()}), app.configurator,
				module.VersionMap{
					"bank":          1,
					"auth":          auth.AppModule{}.ConsensusVersion(),
					"authz":         authzmodule.AppModule{}.ConsensusVersion(),
					"staking":       staking.AppModule{}.ConsensusVersion(),
					"mint":          mint.AppModule{}.ConsensusVersion(),
					"distribution":  distribution.AppModule{}.ConsensusVersion(),
					"slashing":      slashing.AppModule{}.ConsensusVersion(),
					"gov":           gov.AppModule{}.ConsensusVersion(),
					"params":        params.AppModule{}.ConsensusVersion(),
					"upgrade":       upgrade.AppModule{}.ConsensusVersion(),
					"vesting":       vesting.AppModule{}.ConsensusVersion(),
					"feegrant":      feegrantmodule.AppModule{}.ConsensusVersion(),
					"evidence":      evidence.AppModule{}.ConsensusVersion(),
					"crisis":        crisis.AppModule{}.ConsensusVersion(),
					"genutil":       genutil.AppModule{}.ConsensusVersion(),
					"capability":    capability.AppModule{}.ConsensusVersion(),
					"budget":        budget.AppModule{}.ConsensusVersion(),
					"farming":       farming.AppModule{}.ConsensusVersion(),
					"liquidity":     liquidity.AppModule{}.ConsensusVersion(),
					"liquidstaking": liquidstaking.AppModule{}.ConsensusVersion(),
					"liquidfarming": liquidfarming.AppModule{}.ConsensusVersion(),
					"claim":         claim.AppModule{}.ConsensusVersion(),
					"marketmaker":   marketmaker.AppModule{}.ConsensusVersion(),
					"lpfarm":        lpfarm.AppModule{}.ConsensusVersion(),
					"ibc":           ibc.AppModule{}.ConsensusVersion(),
					"ica":           ica.AppModule{}.ConsensusVersion(),
					"transfer":      transfer.AppModule{}.ConsensusVersion(),
					"wasm":          wasm.AppModule{}.ConsensusVersion(),
				},
			)
			if tc.expRunErr {
				require.EqualError(t, err, tc.expRunErrMsg)
			} else {
				require.NoError(t, err)
				// Make sure bank's migration is called.
				require.Equal(t, tc.expCalled, called)
			}
		})
	}
}

func TestInitGenesisOnMigration(t *testing.T) {
	db := dbm.NewMemDB()
	encCfg := MakeTestEncodingConfig()
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	app := NewApp(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encCfg, EmptyAppOptions{}, wasm.EnableAllProposals, EmptyWasmOpts)
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})

	// Create a mock module. This module will serve as the new module we're
	// adding during a migration.
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockModule := mocks.NewMockAppModule(mockCtrl)
	mockDefaultGenesis := json.RawMessage(`{"key": "value"}`)
	mockModule.EXPECT().DefaultGenesis(gomock.Eq(app.appCodec)).Times(1).Return(mockDefaultGenesis)
	mockModule.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(app.appCodec), gomock.Eq(mockDefaultGenesis)).Times(1).Return(nil)
	mockModule.EXPECT().ConsensusVersion().Times(1).Return(uint64(0))

	app.mm.Modules["mock"] = mockModule

	// Run migrations only for "mock" module. We exclude it from
	// the VersionMap to simulate upgrading with a new module.
	_, err := app.mm.RunMigrations(ctx, app.configurator,
		module.VersionMap{
			"bank":          bank.AppModule{}.ConsensusVersion(),
			"auth":          auth.AppModule{}.ConsensusVersion(),
			"authz":         authzmodule.AppModule{}.ConsensusVersion(),
			"staking":       staking.AppModule{}.ConsensusVersion(),
			"mint":          mint.AppModule{}.ConsensusVersion(),
			"distribution":  distribution.AppModule{}.ConsensusVersion(),
			"slashing":      slashing.AppModule{}.ConsensusVersion(),
			"gov":           gov.AppModule{}.ConsensusVersion(),
			"params":        params.AppModule{}.ConsensusVersion(),
			"upgrade":       upgrade.AppModule{}.ConsensusVersion(),
			"vesting":       vesting.AppModule{}.ConsensusVersion(),
			"feegrant":      feegrantmodule.AppModule{}.ConsensusVersion(),
			"evidence":      evidence.AppModule{}.ConsensusVersion(),
			"crisis":        crisis.AppModule{}.ConsensusVersion(),
			"genutil":       genutil.AppModule{}.ConsensusVersion(),
			"capability":    capability.AppModule{}.ConsensusVersion(),
			"budget":        budget.AppModule{}.ConsensusVersion(),
			"farming":       farming.AppModule{}.ConsensusVersion(),
			"liquidity":     liquidity.AppModule{}.ConsensusVersion(),
			"liquidstaking": liquidstaking.AppModule{}.ConsensusVersion(),
			"liquidfarming": liquidfarming.AppModule{}.ConsensusVersion(),
			"claim":         claim.AppModule{}.ConsensusVersion(),
			"marketmaker":   marketmaker.AppModule{}.ConsensusVersion(),
			"lpfarm":        lpfarm.AppModule{}.ConsensusVersion(),
			"ibc":           ibc.AppModule{}.ConsensusVersion(),
			"transfer":      transfer.AppModule{}.ConsensusVersion(),
			"ica":           ica.AppModule{}.ConsensusVersion(),
			"wasm":          wasm.AppModule{}.ConsensusVersion(),
		},
	)
	require.NoError(t, err)
}

func TestUpgradeStateOnGenesis(t *testing.T) {
	encCfg := MakeTestEncodingConfig()
	db := dbm.NewMemDB()
	app := NewApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encCfg, EmptyAppOptions{}, wasm.EnableAllProposals, EmptyWasmOpts)
	genesisState := NewDefaultGenesisState(encCfg.Marshaler)
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)

	// make sure the upgrade keeper has version map in state
	ctx := app.NewContext(false, tmproto.Header{})
	vm := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	for v, i := range app.mm.Modules {
		require.Equal(t, vm[v], i.ConsensusVersion())
	}
}

func TestGetEnabledProposals(t *testing.T) {
	cases := map[string]struct {
		proposalsEnabled string
		specificEnabled  string
		expected         []wasm.ProposalType
	}{
		"all disabled": {
			proposalsEnabled: "false",
			expected:         wasm.DisableAllProposals,
		},
		"all enabled": {
			proposalsEnabled: "true",
			expected:         wasm.EnableAllProposals,
		},
		"some enabled": {
			proposalsEnabled: "okay",
			specificEnabled:  "StoreCode,InstantiateContract",
			expected:         []wasm.ProposalType{wasm.ProposalTypeStoreCode, wasm.ProposalTypeInstantiateContract},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			WasmProposalsEnabled = tc.proposalsEnabled
			EnableSpecificProposals = tc.specificEnabled
			proposals := GetEnabledProposals()
			require.Equal(t, tc.expected, proposals)
		})
	}
}
