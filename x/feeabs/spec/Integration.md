## Example integration of the Fee Abstraction module

```
// app.go
import (
    ... 
    feeabsmodule "github.com/notional-labs/feeabstraction/v1/x/feeabs"
 feeabskeeper "github.com/notional-labs/feeabstraction/v1/x/feeabs/keeper"
 feeabstypes "github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
    ...

)
// Register the AppModule for the fee middleware module
ModuleBasics = module.NewBasicManager(
  ...
  feeabsmodule.AppModuleBasic{},
  ...
)

... 

// Add module account permissions for the fee abstractions
maccPerms = map[string][]string{
  ...
  feeabsmodule.ModuleName:            nil,
}

...

// Add fee abstractions Keeper
type App struct {
  ...

  FeeabsKeeper feeabskeeper.Keeper

  ...
}

...

// Create store keys 
keys := sdk.NewKVStoreKeys(
  ...
  feeabstypes.StoreKey,
  ...
)

... 

app.FeeabsKeeper = feeabskeeper.NewKeeper(
  appCodec,
  keys[feeabstypes.StoreKey],
  keys[feeabstypes.MemStoreKey],
  app.GetSubspace(feeabstypes.ModuleName),
  app.StakingKeeper,
  app.AccountKeeper,
  app.BankKeeper,
  app.TransferKeeper,
  app.IBCKeeper.ChannelKeeper,
  &app.IBCKeeper.PortKeeper,
  scopedFeeabsKeeper,
)

....
// IBC module to fee abstraction
  feeabsIBCModule := feeabsmodule.NewIBCModule(appCodec, app.FeeabsKeeper)
 // Create static IBC router, add app routes, then set and seal it
 ibcRouter := porttypes.NewRouter()

 ibcRouter.
  AddRoute(wasm.ModuleName, wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper)).
  AddRoute(ibctransfertypes.ModuleName, transferIBCModule).
  AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
  AddRoute(feeabstypes.ModuleName, feeabsIBCModule)

 app.IBCKeeper.SetRouter(ibcRouter)
...

// Add fee abstraction to begin blocker logic

app.moduleManager.SetOrderBeginBlockers(
  ...
  feeabstypes.ModuleName,
  ...
)

// Add fee abstraction to end blocker logic
app.moduleManager.SetOrderEndBlockers(
  ...
  feeabstypes.ModuleName,
  ...
)

// Add fee abstraction to init genesis logic
app.moduleManager.SetOrderInitGenesis(
  ...
  feeabstypes.ModuleName,
  ...
)
```

## Configuring with Fee Abtraction param and HostZoneConfig
In order to use Fee Abstraction, we need to add the HostZoneConfig as specified in the government proposals.
