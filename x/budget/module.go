package budget

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/crescent-network/crescent/v5/x/budget/client/cli"
	"github.com/crescent-network/crescent/v5/x/budget/keeper"
	"github.com/crescent-network/crescent/v5/x/budget/simulation"
	"github.com/crescent-network/crescent/v5/x/budget/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the budget module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the budget module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the budget module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// DefaultGenesis returns default genesis state as raw bytes for the budget
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the budget module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return types.ValidateGenesis(data)
}

// RegisterRESTRoutes registers the REST routes for the budget module.
func (AppModuleBasic) RegisterRESTRoutes(_ sdkclient.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the budget module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the budget module.
// Modifying parameters of a budget can be done through governance process, not through transaction level.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the root query command for the budget module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaces implements InterfaceModule
func (b AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry) {}

// AppModule implements an application module for the budget module.
type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec, keeper keeper.Keeper, accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
	}
}

// Name returns the budget module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the budget module invariants.
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the budget module.
func (am AppModule) Route() sdk.Route {
	// Modification of Budgets of Params proceeds to governance proposition, not to Tx.
	return sdk.NewRoute(types.RouterKey, nil)
}

// QuerierRoute returns the budget module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// LegacyQuerierHandler returns the budget module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.Querier{Keeper: am.keeper})
}

// InitGenesis performs genesis initialization for the budget module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the budget
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock returns the begin blocker for the budget module.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock returns the end blocker for the budget module. It returns no validator
// updates.
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the budget module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the budget content functions used to
// simulate governance proposals.
func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return simulation.ProposalContents(am.keeper)
}

// RandomizedParams creates randomized budget param changes for the simulator.
func (AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return simulation.ParamChanges(r)
}

// RegisterStoreDecoder registers a decoder for budget module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
