package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

// GetPort returns the portID for the module. Used in ExportGenesis.
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.IBCPortKey))
}

// DONTCOVER
// No need to cover this simple methods

// IsBound checks if the module is already bound to the desired port.
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

// BindPort defines a wrapper function for the port Keeper's function in
// order to expose it to module's InitGenesis function.
func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	capability := k.portKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, capability, host.PortPath(portID))
}

// SetPort sets the portID for the module. Used in InitGenesis.
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.IBCPortKey, []byte(portID))
}

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function.
func (k Keeper) AuthenticateCapability(ctx sdk.Context, capability *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, capability, name)
}

// ClaimCapability wraps the scopedKeeper's ClaimCapability method.
func (k Keeper) ClaimCapability(ctx sdk.Context, capability *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, capability, name)
}

// Send request for query EstimateSwapExactAmountIn over IBC. Move to use TWAP.
func (k Keeper) SendOsmosisQueryRequest(ctx sdk.Context, poolId uint64, baseDenom string, quoteDenom string, startTime time.Time, sourcePort, sourceChannel string) error {
	path := "/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow" // hard code for now should add to params
	packetData := types.NewQueryArithmeticTwapToNowRequest(poolId, baseDenom, quoteDenom, startTime)
	fmt.Println("========packetData=============")
	fmt.Println(packetData)
	fmt.Println("=========packetData============")

	// 	========packetData=============
	// {1 ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878 uosmo 2023-02-19 22:48:46.847436117 +0000 UTC}
	// =========packetData============
	_, err := k.SendInterchainQuery(ctx, path, packetData.GetBytes(), sourcePort, sourceChannel)
	if err != nil {
		return err
	}

	return nil
}

// Send request for query state over IBC
func (k Keeper) SendInterchainQuery(
	ctx sdk.Context,
	path string,
	data []byte,
	sourcePort string,
	sourceChannel string,
) (uint64, error) {
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return 0, sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", sourcePort, sourceChannel,
		)
	}
	sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return 0, sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
	}

	destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
	destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

	timeoutHeight := clienttypes.NewHeight(0, 100000000)
	timeoutTimestamp := uint64(0)

	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	reqs := []types.InterchainQueryRequest{
		{
			Data: data,
			Path: path,
		},
	}

	packetData := types.NewInterchainQueryRequestPacket(reqs)

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)

	if err := k.channelKeeper.SendPacket(ctx, channelCap, packet); err != nil {
		return 0, err
	}

	return sequence, nil
}

func (k Keeper) GetChannelId(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.KeyChannelID))
}

// TODO: need to test this function
func (k Keeper) UnmarshalPacketBytesToICQtResponses(bz []byte) (types.IcqRespones, error) {
	var res types.IcqRespones
	err := json.Unmarshal(bz, &res)
	if err != nil {
		return types.IcqRespones{}, sdkerrors.New("ibc ack data umarshal", 1, err.Error())
	}

	return res, nil
}

// TODO: add testing
func (k Keeper) GetDecTWAPFromBytes(bz []byte) (sdk.Dec, error) {
	var ibcTokenTwap types.ArithmeticTWAP
	err := json.Unmarshal(bz, &ibcTokenTwap)
	if err != nil {
		return sdk.Dec{}, sdkerrors.New("arithmeticTwap data umarshal", 1, err.Error())
	}

	ibcTokenTwapDec, err := sdk.NewDecFromStr(ibcTokenTwap.ArithmeticTWAP)
	if err != nil {
		return sdk.Dec{}, sdkerrors.New("ibc ack data umarshal", 1, "error when NewDecFromStr")
	}
	return ibcTokenTwapDec, nil
}

func (k Keeper) transferIBCTokenToHostChain(ctx sdk.Context, hostChainConfig types.HostChainFeeAbsConfig) error {
	moduleAccountAddress := k.GetFeeAbsModuleAddress()
	token := k.bk.GetBalance(ctx, moduleAccountAddress, hostChainConfig.IbcDenom)
	nativeDenomIBCedInOsmosis := k.GetParams(ctx).NativeIbcedInOsmosis

	// TODO: don't use it in product version.
	if sdk.NewInt(1).GTE(token.Amount) {
		return nil
	}

	inputToken := sdk.NewCoin(hostChainConfig.HostChainNativeDenomIbcedOnOsmosis, token.Amount)
	memo, err := types.BuildPacketMiddlewareMemo(inputToken, nativeDenomIBCedInOsmosis, moduleAccountAddress.String(), hostChainConfig)
	if err != nil {
		return err
	}

	transferMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostChainConfig.IbcTransferChannel,
		Token:            token,
		Sender:           moduleAccountAddress.String(),
		Receiver:         hostChainConfig.MiddlewareAddress,
		TimeoutHeight:    clienttypes.NewHeight(0, 100000000),
		TimeoutTimestamp: uint64(0),
		Memo:             memo,
	}

	_, err = k.executeTransferMsg(ctx, &transferMsg)
	if err != nil {
		return err
	}

	return nil
}

// TODO: don't use if/else logic.
func (k Keeper) transferIBCTokenToOsmosisChain(ctx sdk.Context, hostChainConfig types.HostChainFeeAbsConfig) error {
	moduleAccountAddress := k.GetFeeAbsModuleAddress()
	token := k.bk.GetBalance(ctx, moduleAccountAddress, hostChainConfig.IbcDenom)
	nativeDenomIBCedInOsmosis := k.GetParams(ctx).NativeIbcedInOsmosis

	// TODO: don't use it in product version.
	if sdk.NewInt(1).GTE(token.Amount) {
		return nil
	}

	inputToken := sdk.NewCoin(hostChainConfig.HostChainNativeDenomIbcedOnOsmosis, token.Amount)
	memo, err := types.BuildCrossChainSwapMemo(inputToken, nativeDenomIBCedInOsmosis, hostChainConfig.CrosschainSwapAddress, moduleAccountAddress.String())
	if err != nil {
		return err
	}

	transferMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostChainConfig.IbcTransferChannel,
		Token:            token,
		Sender:           moduleAccountAddress.String(),
		Receiver:         hostChainConfig.MiddlewareAddress,
		TimeoutHeight:    clienttypes.NewHeight(0, 100000000),
		TimeoutTimestamp: uint64(0),
		Memo:             memo,
	}

	_, err = k.executeTransferMsg(ctx, &transferMsg)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) executeTransferMsg(ctx sdk.Context, transferMsg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error) {
	if err := transferMsg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("bad msg %v", err.Error())
	}
	return k.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), transferMsg)

}

// TODO: use TWAP instead of spotprice
func (k Keeper) handleOsmosisIbcQuery(ctx sdk.Context) {
	// TODO: it should be a chain param
	startTime := ctx.BlockTime().Add(-time.Minute * 5)

	params := k.GetParams(ctx)
	channelID := params.OsmosisQueryChannel

	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		err := k.SendOsmosisQueryRequest(ctx, hostZoneConfig.PoolId, params.NativeIbcedInOsmosis, "uosmo", startTime, types.IBCPortID, channelID)
		if err != nil {
			k.FronzenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
		}
		return false
	})
}
