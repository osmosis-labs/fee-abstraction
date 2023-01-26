package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v4/modules/apps/transfer/keeper"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
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
	//TODO: need to use expected keeper
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

func (k Keeper) GetModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(types.ModuleName)
}

// need to refactor
func (k Keeper) CalculateNativeFromIBCCoins(ctx sdk.Context, ibcCoins sdk.Coins) (coins sdk.Coins, err error) {
	err = k.verifyIBCCoins(ctx, ibcCoins)
	if err != nil {
		return sdk.Coins{}, err
	}
	// add logic here
	spotPrice, err := k.GetOsmosisExchangeRate(ctx)
	if err != nil {
		return sdk.Coins{}, err
	}

	nativeAmount := spotPrice.MulInt(ibcCoins[0].Amount).TruncateInt()

	return sdk.Coins{sdk.Coin{Amount: nativeAmount, Denom: "stake"}}, nil
}

// return err if IBC token isn't in allowed_list
func (k Keeper) verifyIBCCoins(ctx sdk.Context, ibcCoin sdk.Coins) error {
	osmosisDenom := k.GetOsmosisIBCDenomParams(ctx)

	// khanh ngu vcl
	if ibcCoin[0].Denom != osmosisDenom {
		return fmt.Errorf("unallowed denom for tx fee")
	}

	return nil
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

// GetParams gets the fee abstraction module's parameters.
func (k Keeper) GetOsmosisIBCDenomParams(ctx sdk.Context) (denom string) {
	params := k.GetParams(ctx)
	return params.OsmosisIbcDenom
}

// need to refactor
func (k Keeper) SendFeeFromFeePayerToModuleAccount(ctx sdk.Context, feePayer sdk.AccAddress, token sdk.Coins) {
	k.bk.SendCoinsFromAccountToModule(ctx, feePayer, "fee_collector", token)
	k.bk.SendCoinsFromModuleToAccount(ctx, "fee_collector", k.GetModuleAddress(), token)
}
