package feather_connect

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"github.com/terra-money/feather-core/x/feather-connect/keeper"
	"github.com/terra-money/feather-core/x/feather-connect/types"
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

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the ibc-transfer module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}

func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// AppModule represents the AppModule for this module
type AppModule struct {
	AppModuleBasic
	cdc    codec.Codec
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		keeper: k,
	}
}

func (a AppModule) InitGenesis(ctx sdk.Context, jsonCodec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	var genesis types.GenesisState
	jsonCodec.MustUnmarshalJSON(message, &genesis)
	return a.keeper.InitGenesis(ctx, &genesis)
}

func (a AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	genesis := a.keeper.ExportGenesis(ctx)
	return a.cdc.MustMarshalJSON(genesis)
}

func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlock(ctx, am.keeper)
}
