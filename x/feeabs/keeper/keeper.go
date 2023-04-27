package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v4/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/notional-labs/fee-abstraction/v2/x/feeabs/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       sdk.StoreKey
	sk             types.StakingKeeper
	ak             types.AccountKeeper
	bk             types.BankKeeper
	transferKeeper ibctransferkeeper.Keeper
	paramSpace     paramtypes.Subspace

	// ibc keeper
	portKeeper    types.PortKeeper
	channelKeeper types.ChannelKeeper
	scopedKeeper  types.ScopedKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	sk types.StakingKeeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	transferKeeper ibctransferkeeper.Keeper,

	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper types.ScopedKeeper,

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

// need to implement
func (k Keeper) CalculateNativeFromIBCCoins(ctx sdk.Context, ibcCoins sdk.Coins, chainConfig types.HostChainFeeAbsConfig) (coins sdk.Coins, err error) {
	err = k.verifyIBCCoins(ctx, ibcCoins)
	if err != nil {
		return sdk.Coins{}, nil
	}

	twapRate, err := k.GetTwapRate(ctx, chainConfig.IbcDenom)
	if err != nil {
		return sdk.Coins{}, nil
	}

	// mul
	coin := ibcCoins[0]
	nativeFeeAmount := twapRate.MulInt(coin.Amount).RoundInt()
	nativeFee := sdk.NewCoin(k.sk.BondDenom(ctx), nativeFeeAmount)

	return sdk.NewCoins(nativeFee), nil
}

func (k Keeper) SendAbstractionFeeToModuleAccount(ctx sdk.Context, IBCcoins sdk.Coins, nativeCoins sdk.Coins, feePayer sdk.AccAddress) error {
	err := k.bk.SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, IBCcoins)
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

	if k.HasHostZoneConfig(ctx, ibcCoins[0].Denom) {
		return nil
	}
	// TODO: we should register error for this
	return fmt.Errorf("unallowed %s for tx fee", ibcCoins[0].Denom)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
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

// OnTimeoutPacket resend packet when timeout
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	return k.channelKeeper.SendPacket(ctx, chanCap, packet)
}

func (k Keeper) GetCapability(ctx sdk.Context, name string) *capabilitytypes.Capability {
	cap, ok := k.scopedKeeper.GetCapability(ctx, name)
	if !ok {
		k.Logger(ctx).Error("Error ErrChannelCapabilityNotFound ")
		return nil
	}
	return cap
}
