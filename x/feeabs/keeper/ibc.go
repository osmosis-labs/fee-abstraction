package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	"github.com/notional-labs/feeabstraction/v2/x/feeabs/types"
	abci "github.com/tendermint/tendermint/abci/types"
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
func (k Keeper) SendOsmosisQueryRequest(ctx sdk.Context, twapReqs []types.QueryArithmeticTwapToNowRequest, sourcePort, sourceChannel string) error {
	params := k.GetParams(ctx)
	IcqReqs := make([]abci.RequestQuery, len(twapReqs))
	for i, req := range twapReqs {
		IcqReqs[i] = abci.RequestQuery{
			Path: params.OsmosisQueryTwapPath,
			Data: k.cdc.MustMarshal(&req),
		}
	}

	_, err := k.SendInterchainQuery(ctx, IcqReqs, sourcePort, sourceChannel)
	if err != nil {
		return err
	}

	return nil
}

// Send request for query state over IBC
func (k Keeper) SendInterchainQuery(
	ctx sdk.Context,
	reqs []abci.RequestQuery,
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

	timeoutTimestamp := ctx.BlockTime().Add(time.Minute * 5).UnixNano()
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	data, err := types.SerializeCosmosQuery(reqs)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "could not serialize reqs into cosmos query")
	}
	icqPacketData := types.NewInterchainQueryPacketData(data, "")

	packet := channeltypes.NewPacket(
		icqPacketData.GetBytes(),
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		clienttypes.ZeroHeight(),
		uint64(timeoutTimestamp),
	)

	if err := k.channelKeeper.SendPacket(ctx, channelCap, packet); err != nil {
		return 0, err
	}

	return sequence, nil
}

func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, ack channeltypes.Acknowledgement, icqReqs []abci.RequestQuery) error {
	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		var ackData types.InterchainQueryPacketAck
		if err := types.ModuleCdc.UnmarshalJSON(resp.Result, &ackData); err != nil {
			return sdkerrors.Wrap(err, "failed to unmarshal interchain query packet ack")
		}

		ICQResponses, err := types.DeserializeCosmosResponse(ackData.Data)
		if err != nil {
			return sdkerrors.Wrap(err, "could not deserialize data to cosmos response")
		}

		index := 0
		k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
			// Get icq data
			icqReqData, reqPosition, found := k.getQueryArithmeticTwapToNowRequest(ctx, icqReqs, index)
			// update the index
			index = reqPosition
			if !found {
				// if not found any request, end the iterator
				return true
			}
			// Check if icq TWAP denom match with hostzone denom store
			if icqReqData.QuoteAsset != hostZoneConfig.OsmosisPoolTokenDenomIn {
				return false
			}
			// Get icq QueryArithmeticTwapToNowRequest response
			IcqRes := ICQResponses[index]
			index++

			if IcqRes.Code != 0 {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to send interchain query code %d", IcqRes.Code))
				err := k.FrozenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
				if err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("Failed to frozen host zone %s", err.Error()))
				}
				return false
			}

			twapRate, err := k.GetDecTWAPFromBytes(IcqRes.Value)
			if err != nil {
				k.Logger(ctx).Error("Failed to get twap")
				return false
			}
			k.Logger(ctx).Info(fmt.Sprintf("TwapRate %v", twapRate))
			k.SetTwapRate(ctx, hostZoneConfig.IbcDenom, twapRate)

			return false
		})

		k.Logger(ctx).Info("packet ICQ request successfully")

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
			err := k.FrozenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to frozen host zone %s", err.Error()))
			}

			return false
		})
		k.Logger(ctx).Error(fmt.Sprintf("failed to send packet ICQ request %v", resp.Error))

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
			),
		)
	}
	return nil
}

func (k Keeper) getQueryArithmeticTwapToNowRequest(
	ctx sdk.Context,
	icqReqs []abci.RequestQuery,
	index int,
) (types.QueryArithmeticTwapToNowRequest, int, bool) {
	packetLen := len(icqReqs)
	found := false
	var icqReqData types.QueryArithmeticTwapToNowRequest
	for (index < packetLen) && (!found) {
		icqReq := icqReqs[index]
		if err := k.cdc.Unmarshal(icqReq.GetData(), &icqReqData); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to unmarshal icqReqData %s", err.Error()))
			index++
		} else {
			found = true
		}
	}

	return icqReqData, index, found
}

func (k Keeper) GetChannelId(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.KeyChannelID))
}

// TODO: add testing
func (k Keeper) GetDecTWAPFromBytes(bz []byte) (sdk.Dec, error) {
	var ibcTokenTwap types.QueryArithmeticTwapToNowResponse
	err := k.cdc.Unmarshal(bz, &ibcTokenTwap)
	if err != nil {
		return sdk.Dec{}, sdkerrors.New("arithmeticTwap data umarshal", 1, err.Error())
	}

	return ibcTokenTwap.ArithmeticTwap, nil
}

func (k Keeper) transferIBCTokenToHostChainWithMiddlewareMemo(ctx sdk.Context, hostChainConfig types.HostChainFeeAbsConfig) error {
	moduleAccountAddress := k.GetFeeAbsModuleAddress()
	token := k.bk.GetBalance(ctx, moduleAccountAddress, hostChainConfig.IbcDenom)
	nativeDenomIBCedInOsmosis := k.GetParams(ctx).NativeIbcedInOsmosis

	// TODO: don't use it in product version. Use params instead of.
	if sdk.NewInt(1).GTE(token.Amount) {
		return nil
	}

	inputToken := sdk.NewCoin(hostChainConfig.OsmosisPoolTokenDenomIn, token.Amount)
	memo, err := types.BuildPacketMiddlewareMemo(inputToken, nativeDenomIBCedInOsmosis, moduleAccountAddress.String(), hostChainConfig)
	if err != nil {
		return err
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Minute * 5).UnixNano()

	transferMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostChainConfig.IbcTransferChannel,
		Token:            token,
		Sender:           moduleAccountAddress.String(),
		Receiver:         hostChainConfig.MiddlewareAddress,
		TimeoutHeight:    clienttypes.ZeroHeight(),
		TimeoutTimestamp: uint64(timeoutTimestamp),
		Memo:             memo,
	}

	_, err = k.executeTransferMsg(ctx, &transferMsg)
	if err != nil {
		return err
	}

	return nil
}

// TODO: don't use if/else logic.
func (k Keeper) transferIBCTokenToOsmosisChainWithIBCHookMemo(ctx sdk.Context, hostChainConfig types.HostChainFeeAbsConfig) error {
	moduleAccountAddress := k.GetFeeAbsModuleAddress()
	token := k.bk.GetBalance(ctx, moduleAccountAddress, hostChainConfig.IbcDenom)
	nativeDenomIBCedInOsmosis := k.GetParams(ctx).NativeIbcedInOsmosis

	// TODO: don't use it in product version.
	if sdk.NewInt(1).GTE(token.Amount) {
		return nil
	}

	inputToken := sdk.NewCoin(hostChainConfig.OsmosisPoolTokenDenomIn, token.Amount)
	memo, err := types.BuildCrossChainSwapMemo(inputToken, nativeDenomIBCedInOsmosis, hostChainConfig.CrosschainSwapAddress, moduleAccountAddress.String())
	if err != nil {
		return err
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Minute * 5).UnixNano()

	transferMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostChainConfig.IbcTransferChannel,
		Token:            token,
		Sender:           moduleAccountAddress.String(),
		Receiver:         hostChainConfig.CrosschainSwapAddress,
		TimeoutHeight:    clienttypes.ZeroHeight(),
		TimeoutTimestamp: uint64(timeoutTimestamp),
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

func (k Keeper) handleOsmosisIbcQuery(ctx sdk.Context) error {
	hasQueryEpochInfo := k.HasEpochInfo(ctx, types.DefaultQueryEpochIdentifier)
	if !hasQueryEpochInfo {
		k.Logger(ctx).Error(fmt.Sprintf("Don't have query epoch information: %s", types.DefaultQueryEpochIdentifier))
		return nil
	}

	// set startTime for query twap
	queryTwapEpochInfo := k.GetEpochInfo(ctx, types.DefaultQueryEpochIdentifier)
	startTime := ctx.BlockTime().Add(-queryTwapEpochInfo.Duration)
	k.Logger(ctx).Info(fmt.Sprintf("Start time: %v", startTime.Unix()))

	params := k.GetParams(ctx)

	var reqs []types.QueryArithmeticTwapToNowRequest
	var queryChannel string
	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		req := types.NewQueryArithmeticTwapToNowRequest(
			hostZoneConfig.PoolId,
			params.NativeIbcedInOsmosis,
			hostZoneConfig.OsmosisPoolTokenDenomIn,
			startTime,
		)
		reqs = append(reqs, req)
		queryChannel = hostZoneConfig.OsmosisQueryChannel
		return false
	})
	err := k.SendOsmosisQueryRequest(ctx, reqs, types.IBCPortID, queryChannel)
	if err != nil {
		return err
	}

	return nil
}

// executeAllHostChainTWAPQuery will iterate all hostZone and send the IBC Query Packet to Osmosis chain.
func (k Keeper) executeAllHostChainTWAPQuery(ctx sdk.Context) {
	err := k.handleOsmosisIbcQuery(ctx)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error executeAllHostChainTWAPQuery %s", err.Error()))
	}
}

// executeAllHostChainTWAPSwap will iterate all hostZone and execute swap over chain.
func (k Keeper) executeAllHostChainSwap(ctx sdk.Context) {
	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		var err error

		if hostZoneConfig.Frozen {
			return false
		}

		if hostZoneConfig.IsOsmosis {
			err = k.transferIBCTokenToOsmosisChainWithIBCHookMemo(ctx, hostZoneConfig)
		} else {
			err = k.transferIBCTokenToHostChainWithMiddlewareMemo(ctx, hostZoneConfig)
		}

		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to transfer IBC token %s", err.Error()))
			err = k.FrozenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to frozem host zone %s", err.Error()))
			}
		}

		return false
	})
}
