# Integrate the Fee-abstraction module

## Overview
A problem that regularly arises for new users of a Cosmos appchain is that they can't make their first transaction without having the native token. This can be a fairly serious barrier for new users, as it adds many steps to their experience with the application. Notional's fee abstraction module is an open-source Cosmos-SDK module that allows users to pay the fees on a host blockchain with IBC tokens from other blockchains, signifcantly reducing the number of steps required for a large class of new users.

## Prerequisites 
Projects that want to integrate the fee-abstraction module onto their Cosmos SDK chain must enable the following modules:
- [packet-forward-middleware](https://github.com/cosmos/ibc-apps/tree/main/middleware/packet-forward-middleware): Middleware for forwarding IBC packets. This middleware allows for the transfer ICS20 packet using Osmosis Path Unwinding
- [x/staking](https://github.com/cosmos/cosmos-sdk/tree/main/x/staking): The fee-abstraction module must know what token it will swap for, and it retrieves this information from the staking module under the token bond denom.
- [x/auth](https://github.com/cosmos/cosmos-sdk/tree/main/x/auth): the Fee-abstraction module will send the swapped tokens to its own module account to process the original transaction. In order to access its module account address, it needs access to the auth module.
- [x/bank](https://github.com/cosmos/cosmos-sdk/tree/main/x/bank): Allows Fee-abstraction to manage the balances on its own module account. 
- [ibc-transfer](https://github.com/cosmos/ibc-go): Allows the fee-abstraction module to transfer and receive ibc packets.

## Configuring and Adding Fee-abstraction
1. Add the Fee-abstraction package to the go.mod and install it.
    ```
    require (
    ...
    github.com/osmosis-labs/fee-abstraction v<VERSION>
    ...
    )
    ```
  **Note:** The version of the fee-abstraction module will depend on which version of the Cosmos SDK your chain is using. If in doubt about which version to use, please consult the documentation: https://github.com/osmosis-labs/fee-abstraction
  
2. Add the following modules to `app.go`
    ```
    import (
    ... 
        feeabsmodule "github.com/notional-labs/fee-abstraction/v2/x/feeabs"
        feeabskeeper "github.com/notional-labs/fee-abstraction/v2/x/feeabs/keeper"
        feeabstypes "github.com/notional-labs/fee-abstraction/v2/x/feeabs/types"
    ...
    )
    ```
3. In `app.go`: Register the AppModule for the fee middleware module.
    ```
    ModuleBasics = module.NewBasicManager(
      ...
      feeabsmodule.AppModuleBasic{},
      ...
    )
    ```
4. In `app.go`: Add module account permissions for the fee abstractions.
    ```
    maccPerms = map[string][]string{
      ...
      feeabsmodule.ModuleName:            nil,
    }
    // module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		feeabstypes.ModuleName: true,
	}
    ```
5. In `app.go`: Add fee abstraction keeper.
    ```
    type App struct {
      ...
      FeeabsKeeper feeabskeeper.Keeper
      ...
    }
    ```
6. In `app.go`: Add fee abstraction store key.
    ```
    keys := sdk.NewKVStoreKeys(
      ...
      feeabstypes.StoreKey,
      ...
    )
    ```
7. In `app.go`: Instantiate Fee abstraction keeper
    ```
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
    ```
8. In `app.go`: Add the IBC router.
    ```
    feeabsIBCModule := feeabsmodule.NewIBCModule(appCodec, app.FeeabsKeeper)
    
    ibcRouter := porttypes.NewRouter()
    ibcRouter.
    ...
    AddRoute(feeabstypes.ModuleName, feeabsIBCModule)
    ...
    ```
9. In `app.go`: Add the fee-abstraction module to the app manager and simulation manager instantiations.
    ```
    app.mm = module.NewManager(
        ...
        feeabsModule := feeabsmodule.NewAppModule(appCodec, app.FeeabsKeeper),
        ...
    )
    ```
    ```
    app.sm = module.NewSimulationManager(
        ...
        transferModule,
        feeabsModule := feeabsmodule.NewAppModule(appCodec, app.FeeabsKeeper),
        ...
    )
    ```
10. In `app.go`: Add the module as the final element to the following:
- SetOrderBeginBlockers
- SetOrderEndBlockers
- SetOrderInitGenesis
    ```
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
11. In `app.go`: Allow module account address.
    ```
    func (app *FeeAbs) ModuleAccountAddrs() map[string]bool {
	    blockedAddrs := make(map[string]bool)

	    accs := make([]string, 0, len(maccPerms))
	    for k := range maccPerms {
		    accs = append(accs, k)
	    }
	    sort.Strings(accs)

	    for _, acc := range accs {
		    blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	    }

	    return blockedAddrs
    }
    ```
12. In `app.go`: Add to Param keeper.
    ```
    func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	    paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)
        ...
	    paramsKeeper.Subspace(feeabstypes.ModuleName)
        ...
	    return paramsKeeper
    }
    ```
13. Modified Fee Antehandler

    To allow for this, we use modified versions of `MempoolFeeDecorator` and `DeductFeeDecorate`. In these ante handlers, IBC tokens are swapped to the native token before the next fee handler logic is executed.

    If a blockchain uses the Fee Abstraction module, it is necessary to replace the `MempoolFeeDecorator` and `DeductFeeDecorate` with the `FeeAbstrationMempoolFeeDecorator` and `FeeAbstractionDeductFeeDecorate`, respectively. These can be found in `app/ante.go`, and should be implemented as below:
    
    Example:
    ```
    anteDecorators := []sdk.AnteDecorator{
      ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
      ante.NewRejectExtensionOptionsDecorator(),
      feeabsante.NewFeeAbstrationMempoolFeeDecorator(options.FeeAbskeeper),
      ante.NewValidateBasicDecorator(),
      ante.NewTxTimeoutHeightDecorator(),
      ante.NewValidateMemoDecorator(options.AccountKeeper),
      ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
      feeabsante.NewFeeAbstractionDeductFeeDecorate(options.AccountKeeper, options.BankKeeper, options.FeeAbskeeper, options.FeegrantKeeper),
      // SetPubKeyDecorator must be called before all signature verification decorators
      ante.NewSetPubKeyDecorator(options.AccountKeeper),
      ante.NewValidateSigCountDecorator(options.AccountKeeper),
      ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
      ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
      ante.NewIncrementSequenceDecorator(options.AccountKeeper),
      ibcante.NewAnteDecorator(options.IBCKeeper),
     }
    ```
    
## Fee-abstraction Operation
### Create IBC interchain-query channel between Feeabs and Osmosis

To allow feeabs retrieve TWAP prices from Osmosis, we need to create a channel between feeabs and icq module on Osmosis.

    hermes create channel --a-chain osmosis --b-chain feeabs --a-port icqhost --b-port feeabs --new-client-connection --yes

### Register chain connection information on Crosschain Registry contract on Osmosis
In this step, we setup everything that is required for XCSv2 on Osmosis:
- IBC Channel links: All Ibc channel on feeabs chain. Osmosis will use this information for `path unwinding`
```
{
  "modify_chain_channel_links": {
		  "operations": [
			{"operation": "set","source_chain": "chainB","destination_chain": "osmosis","channel_id": "channel-0"},
			{"operation": "set","source_chain": "osmosis","destination_chain": "chainB","channel_id": "channel-0"},
			{"operation": "set","source_chain": "chainB","destination_chain": "chainC","channel_id": "channel-1"},
			{"operation": "set","source_chain": "chainC","destination_chain": "chainB","channel_id": "channel-0"},
			{"operation": "set","source_chain": "osmosis","destination_chain": "chainC","channel_id": "channel-1"},
			{"operation": "set","source_chain": "chainC","destination_chain": "osmosis","channel_id": "channel-1"},
			{"operation": "set","source_chain": "osmosis","destination_chain": "chainB-cw20","channel_id": "channel-2"},
			{"operation": "set","source_chain": "chainB-cw20","destination_chain": "osmosis","channel_id": "channel-2"}
		  ]
  }
}
```
- Chain - Address perfix pair: Osmosis will use this information to find where the receiver is
```
{
  "modify_bech32_prefixes": {
		  "operations": [
			{"operation": "set", "chain_name": "osmosis", "prefix": "osmo"},
			{"operation": "set", "chain_name": "chainB", "prefix": "feeabs"},
			{"operation": "set", "chain_name": "chainC", "prefix": "cosmos"}
		  ]
  }
}
```
- Propose PFM: Confirm that the propose chain has imported PFM this is necessary for `path unwinding`  
```
{
  "propose_pfm": {
    "chain": "chainB"
  }
}

{
  "propose_pfm": {
    "chain": "chainC"
  }
}
```
- Set swap router
```
{
  "set_route": {
    "input_denom":"ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E",
    "output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878",
    "pool_route":[
      {
        "pool_id":"1",
        "token_out_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
      }
    ]
  }
}
```

These setup should be supported by Osmosis team

### Update Feeabs Params via Param-change Gov
```
type Params struct {
	NativeIbcedInOsmosis string 
	OsmosisQueryTwapPath string
	ChainName string 
	IbcTransferChannel string 
	IbcQueryIcqChannel string 
	OsmosisCrosschainSwapAddress string 
}
```
- `NativeIbcedInOsmosis` is the denom of feeabs's native token on Osmosis. Which feeabs module will swap for
- `OsmosisQueryTwapPath` is the `ArithmeticTwapToNow` query path on Osmosis. Default `/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow`
- `ChainName` is the feeabs module chain name. It must be same with the chain name that declare on Osmosis Crosschain Registry Contract
- `IbcTransferChannel` Transfer channel with Osmosis using for swap ibc-token for native-token
- `IbcTransferChannel` Interchain query channel with Osmosis
- `OsmosisCrosschainSwapAddress` XCS contract on Osmosis
### Add HostZone proposal Gov
Add hostzone proposal will allow feeabs to paid fee in ibc denom which is defined

```
type HostChainFeeAbsConfig struct {
	IbcDenom string 
	OsmosisPoolTokenDenomIn string
	PoolId uint64 
	Frozen bool 
}
```
- `IbcDenom` denom of the ibc token that allow to pay fee
- `OsmosisPoolTokenDenomIn` denom of `IbcDenom` on Osmosis
- `PoolId` pool swap between `IbcDenom` and `params.NativeIbcedInOsmosis`
- `Frozen` this is lock flag for update TWAP price. When Add HostZone proposal, it must be `false`