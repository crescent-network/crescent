package app

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	// IBC modules
	ica "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts"
	icahost "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v3/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v3/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v3/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"

	// budget module
	"github.com/crescent-network/crescent/v5/x/budget"
	budgetkeeper "github.com/crescent-network/crescent/v5/x/budget/keeper"
	budgettypes "github.com/crescent-network/crescent/v5/x/budget/types"

	// core modules
	appparams "github.com/crescent-network/crescent/v5/app/params"
	v2_0_0 "github.com/crescent-network/crescent/v5/app/upgrades/mainnet/v2.0.0"
	v3 "github.com/crescent-network/crescent/v5/app/upgrades/mainnet/v3"
	v4 "github.com/crescent-network/crescent/v5/app/upgrades/mainnet/v4"
	v5 "github.com/crescent-network/crescent/v5/app/upgrades/mainnet/v5"
	"github.com/crescent-network/crescent/v5/app/upgrades/testnet/rc4"
	"github.com/crescent-network/crescent/v5/x/amm"
	ammclient "github.com/crescent-network/crescent/v5/x/amm/client"
	ammkeeper "github.com/crescent-network/crescent/v5/x/amm/keeper"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	"github.com/crescent-network/crescent/v5/x/claim"
	claimkeeper "github.com/crescent-network/crescent/v5/x/claim/keeper"
	claimtypes "github.com/crescent-network/crescent/v5/x/claim/types"
	"github.com/crescent-network/crescent/v5/x/exchange"
	exchangeclient "github.com/crescent-network/crescent/v5/x/exchange/client"
	exchangekeeper "github.com/crescent-network/crescent/v5/x/exchange/keeper"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
	"github.com/crescent-network/crescent/v5/x/farming"
	farmingkeeper "github.com/crescent-network/crescent/v5/x/farming/keeper"
	farmingtypes "github.com/crescent-network/crescent/v5/x/farming/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm"
	liquidammclient "github.com/crescent-network/crescent/v5/x/liquidamm/client"
	liquidammkeeper "github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	liquidammtypes "github.com/crescent-network/crescent/v5/x/liquidamm/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming"
	liquidfarmingkeeper "github.com/crescent-network/crescent/v5/x/liquidfarming/keeper"
	liquidfarmingtypes "github.com/crescent-network/crescent/v5/x/liquidfarming/types"
	"github.com/crescent-network/crescent/v5/x/liquidity"
	liquiditykeeper "github.com/crescent-network/crescent/v5/x/liquidity/keeper"
	liquiditytypes "github.com/crescent-network/crescent/v5/x/liquidity/types"
	"github.com/crescent-network/crescent/v5/x/liquidstaking"
	liquidstakingkeeper "github.com/crescent-network/crescent/v5/x/liquidstaking/keeper"
	liquidstakingtypes "github.com/crescent-network/crescent/v5/x/liquidstaking/types"
	"github.com/crescent-network/crescent/v5/x/lpfarm"
	lpfarmclient "github.com/crescent-network/crescent/v5/x/lpfarm/client"
	lpfarmkeeper "github.com/crescent-network/crescent/v5/x/lpfarm/keeper"
	lpfarmtypes "github.com/crescent-network/crescent/v5/x/lpfarm/types"
	"github.com/crescent-network/crescent/v5/x/marker"
	markerkeeper "github.com/crescent-network/crescent/v5/x/marker/keeper"
	markertypes "github.com/crescent-network/crescent/v5/x/marker/types"
	"github.com/crescent-network/crescent/v5/x/marketmaker"
	marketmakerclient "github.com/crescent-network/crescent/v5/x/marketmaker/client"
	marketmakerkeeper "github.com/crescent-network/crescent/v5/x/marketmaker/keeper"
	marketmakertypes "github.com/crescent-network/crescent/v5/x/marketmaker/types"
	"github.com/crescent-network/crescent/v5/x/mint"
	mintkeeper "github.com/crescent-network/crescent/v5/x/mint/keeper"
	minttypes "github.com/crescent-network/crescent/v5/x/mint/types"

	// unnamed import of statik for swagger UI support
	_ "github.com/crescent-network/crescent/v5/client/docs/statik"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler,
			distrclient.ProposalHandler,
			upgradeclient.ProposalHandler,
			upgradeclient.CancelProposalHandler,
			ibcclientclient.UpdateClientProposalHandler,
			ibcclientclient.UpgradeProposalHandler,
			marketmakerclient.ProposalHandler,
			lpfarmclient.ProposalHandler,
			exchangeclient.MarketParameterChangeProposalHandler,
			ammclient.PoolParameterChangeProposalHandler,
			ammclient.PublicFarmingPlanProposalHandler,
			liquidammclient.PublicPositionCreateProposalHandler,
			liquidammclient.PublicPositionParameterChangeProposalHandler,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		transfer.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		vesting.AppModuleBasic{},
		budget.AppModuleBasic{},
		farming.AppModuleBasic{},
		liquidity.AppModuleBasic{},
		liquidstaking.AppModuleBasic{},
		liquidfarming.AppModuleBasic{},
		liquidamm.AppModuleBasic{},
		claim.AppModuleBasic{},
		marketmaker.AppModuleBasic{},
		lpfarm.AppModuleBasic{},
		ica.AppModuleBasic{},
		marker.AppModuleBasic{},
		exchange.AppModuleBasic{},
		amm.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		budgettypes.ModuleName:         nil,
		farmingtypes.ModuleName:        nil,
		liquiditytypes.ModuleName:      {authtypes.Minter, authtypes.Burner},
		liquidstakingtypes.ModuleName:  {authtypes.Minter, authtypes.Burner},
		liquidfarmingtypes.ModuleName:  {authtypes.Minter, authtypes.Burner},
		liquidammtypes.ModuleName:      {authtypes.Minter, authtypes.Burner},
		claimtypes.ModuleName:          nil,
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		marketmakertypes.ModuleName:    nil,
		lpfarmtypes.ModuleName:         nil,
		exchangetypes.ModuleName:       nil,
		ammtypes.ModuleName:            {authtypes.Minter, authtypes.Burner},
		icatypes.ModuleName:            nil,
	}
)

// Verify app interface at compile time
var (
	_ simapp.App              = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

const (
	FlagDisableUpgradeEvents = "disable-upgrade-events"
)

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey

	// keepers
	AccountKeeper       authkeeper.AccountKeeper
	BankKeeper          bankkeeper.Keeper
	CapabilityKeeper    *capabilitykeeper.Keeper
	StakingKeeper       *stakingkeeper.Keeper
	SlashingKeeper      slashingkeeper.Keeper
	MintKeeper          mintkeeper.Keeper
	DistrKeeper         distrkeeper.Keeper
	GovKeeper           govkeeper.Keeper
	CrisisKeeper        crisiskeeper.Keeper
	UpgradeKeeper       upgradekeeper.Keeper
	ParamsKeeper        paramskeeper.Keeper
	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper      evidencekeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper
	FeeGrantKeeper      feegrantkeeper.Keeper
	AuthzKeeper         authzkeeper.Keeper
	BudgetKeeper        budgetkeeper.Keeper
	FarmingKeeper       farmingkeeper.Keeper
	LiquidityKeeper     liquiditykeeper.Keeper
	LiquidStakingKeeper liquidstakingkeeper.Keeper
	LiquidFarmingKeeper liquidfarmingkeeper.Keeper
	LiquidAMMKeeper     liquidammkeeper.Keeper
	ClaimKeeper         claimkeeper.Keeper
	MarketMakerKeeper   marketmakerkeeper.Keeper
	LPFarmKeeper        lpfarmkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	MarkerKeeper        markerkeeper.Keeper
	ExchangeKeeper      exchangekeeper.Keeper
	AMMKeeper           ammkeeper.Keeper

	// scoped keepers
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper

	// IBC app modules
	transferModule transfer.AppModule
	icaModule      ica.AppModule

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, AppUserHomeDir)
}

// NewApp returns a reference to an initialized App.
func NewApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	msgFilterFlag bool,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(AppName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		ibchost.StoreKey,
		upgradetypes.StoreKey,
		evidencetypes.StoreKey,
		ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey,
		feegrant.StoreKey,
		authzkeeper.StoreKey,
		budgettypes.StoreKey,
		farmingtypes.StoreKey,
		liquiditytypes.StoreKey,
		liquidstakingtypes.StoreKey,
		liquidfarmingtypes.StoreKey,
		liquidammtypes.StoreKey,
		claimtypes.StoreKey,
		marketmakertypes.StoreKey,
		lpfarmtypes.StoreKey,
		icahosttypes.StoreKey,
		markertypes.StoreKey,
		exchangetypes.StoreKey,
		ammtypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(
		paramstypes.TStoreKey,
	)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	bApp.SetParamStore(
		app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()),
	)

	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.CapabilityKeeper.Seal()

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		app.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.GetSubspace(banktypes.ModuleName),
		app.ModuleAccountAddrs(),
	)
	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		appCodec,
		app.BaseApp.MsgServiceRouter(),
	)
	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		keys[feegrant.StoreKey],
		app.AccountKeeper,
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(stakingtypes.ModuleName),
	)
	app.StakingKeeper = &stakingKeeper

	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		app.GetSubspace(minttypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		app.GetSubspace(distrtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		keys[slashingtypes.StoreKey],
		app.StakingKeeper,
		app.GetSubspace(slashingtypes.ModuleName),
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		app.GetSubspace(crisistypes.ModuleName),
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
	)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper = app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		app.GetSubspace(ibchost.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
	)
	app.BudgetKeeper = budgetkeeper.NewKeeper(
		appCodec,
		keys[budgettypes.StoreKey],
		app.GetSubspace(budgettypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.ModuleAccountAddrs(),
	)
	app.FarmingKeeper = farmingkeeper.NewKeeper(
		appCodec,
		keys[farmingtypes.StoreKey],
		app.GetSubspace(farmingtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
	)
	app.LiquidityKeeper = liquiditykeeper.NewKeeper(
		appCodec,
		keys[liquiditytypes.StoreKey],
		app.GetSubspace(liquiditytypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
	)
	app.MarketMakerKeeper = marketmakerkeeper.NewKeeper(
		appCodec,
		keys[marketmakertypes.StoreKey],
		app.GetSubspace(marketmakertypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
	)
	app.LPFarmKeeper = lpfarmkeeper.NewKeeper(
		appCodec,
		keys[lpfarmtypes.StoreKey],
		app.GetSubspace(lpfarmtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.LiquidityKeeper,
	)
	app.LiquidStakingKeeper = liquidstakingkeeper.NewKeeper(
		appCodec,
		keys[liquidstakingtypes.StoreKey],
		app.GetSubspace(liquidstakingtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.LiquidityKeeper,
		app.LPFarmKeeper,
		app.SlashingKeeper,
	)
	app.ExchangeKeeper = exchangekeeper.NewKeeper(
		appCodec,
		keys[exchangetypes.StoreKey],
		app.GetSubspace(exchangetypes.ModuleName),
		app.BankKeeper,
	)
	app.MarkerKeeper = markerkeeper.NewKeeper(
		appCodec,
		keys[markertypes.StoreKey],
		app.GetSubspace(markertypes.ModuleName),
	)
	app.AMMKeeper = ammkeeper.NewKeeper(
		appCodec,
		keys[ammtypes.StoreKey],
		app.GetSubspace(ammtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.ExchangeKeeper,
		app.MarkerKeeper,
	)
	app.ExchangeKeeper.SetOrderSources(
		ammkeeper.NewOrderSource(app.AMMKeeper),
	)
	app.LiquidFarmingKeeper = liquidfarmingkeeper.NewKeeper(
		appCodec,
		keys[liquidfarmingtypes.StoreKey],
		app.GetSubspace(liquidfarmingtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.LPFarmKeeper,
		app.LiquidityKeeper,
	)
	app.LiquidAMMKeeper = liquidammkeeper.NewKeeper(
		appCodec,
		keys[liquidammtypes.StoreKey],
		app.GetSubspace(liquidammtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.AMMKeeper,
	)

	// register the proposal types
	govRouter := govtypes.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(farmingtypes.RouterKey, farming.NewPublicPlanProposalHandler(app.FarmingKeeper)).
		AddRoute(marketmakertypes.RouterKey, marketmaker.NewMarketMakerProposalHandler(app.MarketMakerKeeper)).
		AddRoute(lpfarmtypes.RouterKey, lpfarm.NewFarmingPlanProposalHandler(app.LPFarmKeeper)).
		AddRoute(exchangetypes.RouterKey, exchange.NewProposalHandler(app.ExchangeKeeper)).
		AddRoute(ammtypes.RouterKey, amm.NewProposalHandler(app.AMMKeeper)).
		AddRoute(liquidammtypes.RouterKey, liquidamm.NewProposalHandler(app.LiquidAMMKeeper))

	app.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		keys[govtypes.StoreKey],
		app.GetSubspace(govtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		govRouter,
	)
	app.GovKeeper = *app.GovKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
			app.LiquidStakingKeeper.Hooks(),
		),
	)
	app.ClaimKeeper = claimkeeper.NewKeeper(
		appCodec,
		keys[claimtypes.StoreKey],
		app.BankKeeper,
		app.DistrKeeper,
		app.GovKeeper,
		app.LiquidityKeeper,
		app.LiquidStakingKeeper,
	)
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
	)
	app.transferModule = transfer.NewAppModule(app.TransferKeeper)
	transferIBCModule := transfer.NewIBCModule(app.TransferKeeper)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
	)
	app.icaModule = ica.NewAppModule(nil, &app.ICAHostKeeper)
	icaHostIBCModule := icahost.NewIBCModule(app.ICAHostKeeper)

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
	)
	app.EvidenceKeeper = *evidenceKeeper

	// create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()

	ibcRouter.
		AddRoute(ibctransfertypes.ModuleName, transferIBCModule).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule)
	app.IBCKeeper.SetRouter(ibcRouter)

	/****  Module Options ****/
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		budget.NewAppModule(appCodec, app.BudgetKeeper, app.AccountKeeper, app.BankKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		staking.NewAppModule(appCodec, *app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		liquidity.NewAppModule(appCodec, app.LiquidityKeeper, app.AccountKeeper, app.BankKeeper),
		farming.NewAppModule(appCodec, app.FarmingKeeper, app.AccountKeeper, app.BankKeeper),
		liquidstaking.NewAppModule(appCodec, app.LiquidStakingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GovKeeper),
		liquidfarming.NewAppModule(appCodec, app.LiquidFarmingKeeper, app.AccountKeeper, app.BankKeeper),
		liquidamm.NewAppModule(appCodec, app.LiquidAMMKeeper, app.AccountKeeper, app.BankKeeper, app.AMMKeeper),
		claim.NewAppModule(appCodec, app.ClaimKeeper, app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.GovKeeper, app.LiquidityKeeper, app.LiquidStakingKeeper),
		marketmaker.NewAppModule(appCodec, app.MarketMakerKeeper, app.AccountKeeper, app.BankKeeper),
		lpfarm.NewAppModule(appCodec, app.LPFarmKeeper, app.AccountKeeper, app.BankKeeper, app.LiquidityKeeper),
		marker.NewAppModule(appCodec, app.MarkerKeeper),
		exchange.NewAppModule(appCodec, app.ExchangeKeeper, app.AccountKeeper, app.BankKeeper),
		amm.NewAppModule(appCodec, app.AMMKeeper, app.AccountKeeper, app.BankKeeper, app.ExchangeKeeper),
		app.transferModule,
		app.icaModule,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		budgettypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		liquidstakingtypes.ModuleName,
		liquiditytypes.ModuleName,
		liquidfarmingtypes.ModuleName,
		liquidammtypes.ModuleName, // must be prior to amm
		ibchost.ModuleName,
		ammtypes.ModuleName,
		lpfarmtypes.ModuleName,

		// empty logic modules
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		farmingtypes.ModuleName,
		claimtypes.ModuleName,
		marketmakertypes.ModuleName,
		icatypes.ModuleName,
		markertypes.ModuleName,
		exchangetypes.ModuleName,
	)

	app.mm.SetOrderMidBlockers(
		exchangetypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		// EndBlocker of crisis module called AssertInvariants
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		liquiditytypes.ModuleName,
		farmingtypes.ModuleName,
		liquidstakingtypes.ModuleName,

		// empty logic modules
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		claimtypes.ModuleName,
		budgettypes.ModuleName,
		marketmakertypes.ModuleName,
		lpfarmtypes.ModuleName,
		icatypes.ModuleName,
		exchangetypes.ModuleName,
		ammtypes.ModuleName,
		liquidammtypes.ModuleName,
		liquidfarmingtypes.ModuleName,

		markertypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		ibchost.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		ibctransfertypes.ModuleName,
		feegrant.ModuleName,
		authz.ModuleName,
		budgettypes.ModuleName,
		farmingtypes.ModuleName,
		liquiditytypes.ModuleName,
		liquidstakingtypes.ModuleName,
		liquidfarmingtypes.ModuleName,
		liquidammtypes.ModuleName,
		claimtypes.ModuleName,
		marketmakertypes.ModuleName,
		lpfarmtypes.ModuleName,
		markertypes.ModuleName,
		exchangetypes.ModuleName,
		ammtypes.ModuleName,

		// empty logic modules
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		icatypes.ModuleName,

		// InitGenesis of crisis module called AssertInvariants
		crisistypes.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		// Temporarily disable x/authz simulation
		// authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
		budget.NewAppModule(appCodec, app.BudgetKeeper, app.AccountKeeper, app.BankKeeper),
		farming.NewAppModule(appCodec, app.FarmingKeeper, app.AccountKeeper, app.BankKeeper),
		staking.NewAppModule(appCodec, *app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		params.NewAppModule(app.ParamsKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		liquidity.NewAppModule(appCodec, app.LiquidityKeeper, app.AccountKeeper, app.BankKeeper),
		liquidstaking.NewAppModule(appCodec, app.LiquidStakingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GovKeeper),
		liquidfarming.NewAppModule(appCodec, app.LiquidFarmingKeeper, app.AccountKeeper, app.BankKeeper),
		claim.NewAppModule(appCodec, app.ClaimKeeper, app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.GovKeeper, app.LiquidityKeeper, app.LiquidStakingKeeper),
		liquidamm.NewAppModule(appCodec, app.LiquidAMMKeeper, app.AccountKeeper, app.BankKeeper, app.AMMKeeper),
		marketmaker.NewAppModule(appCodec, app.MarketMakerKeeper, app.AccountKeeper, app.BankKeeper),
		lpfarm.NewAppModule(appCodec, app.LPFarmKeeper, app.AccountKeeper, app.BankKeeper, app.LiquidityKeeper),
		marker.NewAppModule(appCodec, app.MarkerKeeper),
		exchange.NewAppModule(appCodec, app.ExchangeKeeper, app.AccountKeeper, app.BankKeeper),
		amm.NewAppModule(appCodec, app.AMMKeeper, app.AccountKeeper, app.BankKeeper, app.ExchangeKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		app.transferModule,
	)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				FeegrantKeeper:  app.FeeGrantKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			Codec:         appCodec,
			GovKeeper:     &app.GovKeeper,
			IBCKeeper:     app.IBCKeeper,
			MsgFilterFlag: msgFilterFlag,
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}

	app.SetAnteHandler(anteHandler)
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetMidBlocker(app.MidBlocker)

	app.SetUpgradeStoreLoaders()
	disableUpgradeEvents := cast.ToBool(appOpts.Get(FlagDisableUpgradeEvents))
	app.SetUpgradeHandlers(app.mm, app.configurator, disableUpgradeEvents)

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(fmt.Sprintf("failed to load latest version: %s", err))
		}
	}

	return app
}

// Name returns the name of the App.
func (app *App) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block.
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block.
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization.
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height.
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// add farming rewards reserve account
	modAccAddrs[farmingtypes.RewardsReserveAcc.String()] = true

	return modAccAddrs
}

// LegacyAmino returns App's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns App's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns App's InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *sdk.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *sdk.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *sdk.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	rpc.RegisterRoutes(clientCtx, apiSvr.Router)
	// Register legacy tx routes.
	authrest.RegisterTxRoutes(clientCtx, apiSvr.Router)
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(budgettypes.ModuleName)
	paramsKeeper.Subspace(farmingtypes.ModuleName)
	paramsKeeper.Subspace(liquiditytypes.ModuleName)
	paramsKeeper.Subspace(liquidstakingtypes.ModuleName)
	paramsKeeper.Subspace(liquidfarmingtypes.ModuleName)
	paramsKeeper.Subspace(liquidammtypes.ModuleName)
	paramsKeeper.Subspace(marketmakertypes.ModuleName)
	paramsKeeper.Subspace(lpfarmtypes.ModuleName)
	paramsKeeper.Subspace(markertypes.ModuleName)
	paramsKeeper.Subspace(exchangetypes.ModuleName)
	paramsKeeper.Subspace(ammtypes.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)

	return paramsKeeper
}

func (app *App) SetUpgradeStoreLoaders() {
	// common logics for set upgrades
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	// testnet upgrade state loaders
	if upgradeInfo.Name == rc4.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &rc4.StoreUpgrades))
	}

	// mainnet upgrade state loaders
	if upgradeInfo.Name == v2_0_0.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &v2_0_0.StoreUpgrades))
	}
	if upgradeInfo.Name == v3.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &v3.StoreUpgrades))
	}
	if upgradeInfo.Name == v4.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &v4.StoreUpgrades))
	}
	if upgradeInfo.Name == v5.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &v5.StoreUpgrades))
	}
}

func (app *App) SetUpgradeHandlers(mm *module.Manager, configurator module.Configurator, enableMigrationEventEmit bool) {
	// testnet upgrade handlers
	app.UpgradeKeeper.SetUpgradeHandler(
		rc4.UpgradeName, rc4.UpgradeHandler)

	// mainnet upgrade handlers
	app.UpgradeKeeper.SetUpgradeHandler(
		v2_0_0.UpgradeName, v2_0_0.UpgradeHandler(mm, configurator, app.MintKeeper, app.BudgetKeeper, app.LiquidityKeeper))

	app.UpgradeKeeper.SetUpgradeHandler(
		v3.UpgradeName, v3.UpgradeHandler(
			mm, configurator, app.MarketMakerKeeper, app.LiquidityKeeper, app.LPFarmKeeper, app.FarmingKeeper, app.BankKeeper))

	app.UpgradeKeeper.SetUpgradeHandler(
		v4.UpgradeName, v4.UpgradeHandler(
			mm, configurator, app.icaModule))

	app.UpgradeKeeper.SetUpgradeHandler(
		v5.UpgradeName, v5.UpgradeHandler(
			mm, configurator, app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.LiquidityKeeper,
			app.LPFarmKeeper, app.ExchangeKeeper, app.AMMKeeper, app.MarkerKeeper, app.FarmingKeeper,
			app.ClaimKeeper, enableMigrationEventEmit))
}
