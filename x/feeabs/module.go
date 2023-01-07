package feeabs

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/keeper"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface for the feeabs module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// NewAppModuleBasic instatiate an AppModuleBasic object
func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name return the feeabs module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec register module codec
// TODO: need to implement
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {

}

// RegisterInterfaces registers the module interface
// TODO: need to implement
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {

}

// DefaultGenesis returns feeabs module default genesis state.
// TODO: need to implement
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis validate genesis state for feeabs module
// TODO: need to implement
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterRESTRoutes registers REST service handlers for feeabs module
// TODO: need to implement
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
// TODO: need to implement
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
}

// GetTxCmd returns the feeabs module's root tx command.
// TODO: need to implement
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the feeabs module's root query command.
// TODO: need to implement
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implement AppModule interface for feeabs module
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

// NewAppModule instantiate AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

// Name return the feeabs module name
func (am AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the feeabs module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route return feeabs module message routing (not need anymore because using ADR 031)
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QueryRouter return feeabs module query routing key
// TODO: implement
func (AppModule) QuerierRoute() string {
	return ""
}

// LegacyQuerierHandler returns feeabs legacy querier handler
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
// TODO: implement
func (am AppModule) RegisterServices(cfg module.Configurator) {

}

// InitGenesis initial genesis state for feeabs module
// TODO: implement
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis export feeabs state as raw message for feeabs module
// TODO: implement
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return json.RawMessage{}
}

// BeginBlock returns the begin blocker for the feeabs module.
// TODO: implement if needed
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
}

// EndBlock returns the end blocker for the feeabs module. It returns no validator
// updates.
// TODO: implement if needed
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ConsensusVersion return module consensus version
func (AppModule) ConsensusVersion() uint64 { return 1 }
