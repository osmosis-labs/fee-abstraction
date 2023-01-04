package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       sdk.StoreKey
	memKey         sdk.StoreKey
	paramstore     paramtypes.Subspace
	transferKeeper ibctransferkeeper.Keeper

	// ibc keeper
	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
	scopedKeeper  types.ScopedKeeper
}

// need to implement
func (k Keeper) GetModuleAddress() sdk.AccAddress {
	return sdk.AccAddress{}
}

// need to implement
func (k Keeper) CalculateNativeFromIBCCoin(ibcCoin sdk.Coins) (coins sdk.Coins, err error) {
	err = k.verifyIBCCoin(ibcCoin)
	if err != nil {
		return sdk.Coins{}, nil
	}
	return coins, nil
}

// TODO : need to implement
// return err if IBC token isn't in allowed_list
func (k Keeper) verifyIBCCoin(ibcCoin sdk.Coins) error {
	return nil
}