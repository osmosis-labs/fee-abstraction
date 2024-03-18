package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"
	portkeeper "github.com/cosmos/ibc-go/v8/modules/core/05-port/keeper"

	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       storetypes.StoreKey
	sk             *stakingkeeper.Keeper
	ak             authkeeper.AccountKeeper
	bk             bankkeeper.BaseKeeper
	transferKeeper ibctransferkeeper.Keeper
	paramSpace     paramtypes.Subspace

	// ibc keeper
	portKeeper    *portkeeper.Keeper
	channelKeeper channelkeeper.Keeper
	scopedKeeper  capabilitykeeper.ScopedKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	sk *stakingkeeper.Keeper,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.BaseKeeper,
	transferKeeper ibctransferkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	portKeeper *portkeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		paramSpace:     ps,
		sk:             sk,
		ak:             ak,
		bk:             bk,
		transferKeeper: transferKeeper,
		channelKeeper:  channelKeeper,
		scopedKeeper:   scopedKeeper,
		portKeeper:     portKeeper,
	}
}

func (k Keeper) GetFeeAbsModuleAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.ak.GetModuleAccount(ctx, types.ModuleName)
}

func (k Keeper) GetFeeAbsModuleAddress() sdk.AccAddress {
	return k.ak.GetModuleAddress(types.ModuleName)
}

func (k Keeper) GetDefaultBondDenom(ctx sdk.Context) (string, error) {
	return k.sk.BondDenom(ctx)
}

// need to implement
func (k Keeper) CalculateNativeFromIBCCoins(ctx sdk.Context, ibcCoins sdk.Coins, chainConfig types.HostChainFeeAbsConfig) (coins sdk.Coins, err error) {
	err = k.verifyIBCCoins(ctx, ibcCoins)
	if err != nil {
		return sdk.Coins{}, err
	}

	twapRate, err := k.GetTwapRate(ctx, chainConfig.IbcDenom)
	if err != nil {
		return sdk.Coins{}, err
	}

	// mul
	coin := ibcCoins[0]
	nativeFeeAmount := twapRate.MulInt(coin.Amount).RoundInt()
	bondDenom, err := k.sk.BondDenom(ctx)
	if err != nil {
		return sdk.Coins{}, err
	}
	nativeFee := sdk.NewCoin(bondDenom, nativeFeeAmount)

	return sdk.NewCoins(nativeFee), nil
}

// SendAbstractionFeeToModuleAccount send IBC token to module account
func (k Keeper) SendAbstractionFeeToModuleAccount(ctx sdk.Context, ibcCoins sdk.Coins, nativeCoins sdk.Coins, feePayer sdk.AccAddress) error {
	err := k.bk.SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, ibcCoins)
	if err != nil {
		return err
	}
	return nil
}

// return err if IBC token isn't in allowed_list
func (k Keeper) verifyIBCCoins(ctx sdk.Context, ibcCoins sdk.Coins) error {
	if ibcCoins.Len() != 1 {
		return types.ErrInvalidIBCFees
	}

	ibcDenom := ibcCoins[0].Denom
	if k.HasHostZoneConfig(ctx, ibcDenom) {
		return nil
	}
	return types.ErrUnsupportedDenom.Wrapf("unsupported denom: %s", ibcDenom)
}

func (Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams gets the fee abstraction module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets all of the parameters in the abstraction module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) GetCapability(ctx sdk.Context, name string) *capabilitytypes.Capability {
	capability, ok := k.scopedKeeper.GetCapability(ctx, name)
	if !ok {
		k.Logger(ctx).Error(fmt.Sprintf("not found capability with given name: %s", name))
		return nil
	}
	return capability
}
