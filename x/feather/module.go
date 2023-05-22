package feather

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"github.com/terra-money/feather-core/x/feather/keeper"
	"github.com/terra-money/feather-core/x/feather/simulation"
	"github.com/terra-money/feather-core/x/feather/types"
)

type AppModuleBasic struct{}

func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

func (a AppModuleBasic) DefaultGenesis(jsonCodec codec.JSONCodec) json.RawMessage {
	return jsonCodec.MustMarshalJSON(DefaultGenesisState())
}

func (a AppModuleBasic) ValidateGenesis(jsonCodec codec.JSONCodec, _ client.TxEncodingConfig, message json.RawMessage) error {
	var genesis types.GenesisState
	if err := jsonCodec.UnmarshalJSON(message, &genesis); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return ValidateGenesis(&genesis)
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	/* #nosec */
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)) //nolint:errcheck
}

// AppModule represents the AppModule for this module
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		keeper: k,
	}
}

func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(a.keeper))
}

func (a AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genesis)
	return a.keeper.InitGenesis(ctx, genesis)
}

func (a AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := a.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genesis)
}

func (a AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlock(ctx, a.keeper)
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the content functions used to
// simulate governance proposals.
func (a AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}

// RegisterStoreDecoder registers a decoder the module
func (a AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (a AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return []simtypes.WeightedOperation{}
}
