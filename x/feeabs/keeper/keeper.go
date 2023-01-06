package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       sdk.StoreKey
	paramstore     paramtypes.Subspace
	transferKeeper ibctransferkeeper.Keeper

	// ibc keeper
	channelKeeper types.ChannelKeeper
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

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
