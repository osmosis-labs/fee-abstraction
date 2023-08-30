package keeper

import (
	"fmt"
	"time"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
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
func (k Keeper) SendOsmosisQueryRequest(ctx sdk.Context, twapReqs []types.QueryArithmeticTwapToNowRequest, sourcePort, sourceChannel string) (uint64, error) {
	params := k.GetParams(ctx)
	icqReqs := make([]abci.RequestQuery, len(twapReqs))
	for i, req := range twapReqs {
		req := req
		icqReqs[i] = abci.RequestQuery{
			Path: params.OsmosisQueryTwapPath,
			Data: k.cdc.MustMarshal(&req),
		}
	}

	data, err := types.SerializeCosmosQuery(icqReqs)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "could not serialize reqs into cosmos query")
	}
	icqPacketData := types.NewInterchainQueryPacketData(data, "")

	sequence, err := k.SendInterchainQuery(ctx, icqPacketData.GetBytes(), sourcePort, sourceChannel)
	if err != nil {
		return 0, err
	}

	return sequence, nil
}

// Send request for query state over IBC
func (k Keeper) SendInterchainQuery(
	ctx sdk.Context,
	data []byte,
	sourcePort string,
	sourceChannel string,
) (uint64, error) {
	timeoutTimestamp := ctx.BlockTime().Add(time.Minute * 5).UnixNano()
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	sequence, err := k.channelKeeper.SendPacket(ctx, channelCap, sourcePort, sourceChannel, clienttypes.ZeroHeight(), uint64(timeoutTimestamp), data)
	if err != nil {
		return 0, err
	}

	return sequence, nil
}

func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, ack channeltypes.Acknowledgement) error {
	// Get hostzone of send request
	hostZoneConfig := k.GetAndRemoveSendingPacketInfo(ctx, packet.Sequence, packet.SourceChannel)

	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		var ackData types.InterchainQueryPacketAck
		if err := types.ModuleCdc.UnmarshalJSON(resp.Result, &ackData); err != nil {
			return sdkerrors.Wrap(err, "failed to unmarshal interchain query packet ack")
		}

		icqResponses, err := types.DeserializeCosmosResponse(ackData.Data)
		if err != nil {
			return sdkerrors.Wrap(err, "could not deserialize data to cosmos response")
		}

		// Current we just query 1 pool so should only have 1 response
		icqRes := icqResponses[0]

		if icqRes.Code != 0 {
			// Query error
			k.Logger(ctx).Error(fmt.Sprintf("Failed to send interchain query code %d", icqRes.Code))
			err := k.FrozenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to frozen host zone %s", err.Error()))
			}
			// resend query
			sequence, err := k.SendInterchainQuery(ctx, packet.GetData(), packet.SourcePort, packet.SourceChannel)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to resend ICQ on Ack result with error response code %s", err.Error()))
				return err
			}
			// set sequence for ack or timeout
			params := k.GetParams(ctx)
			k.SetSendingPacketInfo(ctx, sequence, params.IbcQueryIcqChannel, hostZoneConfig)
			return nil
		}

		twapRate, err := k.GetDecTWAPFromBytes(icqRes.Value)
		if err != nil {
			k.Logger(ctx).Error("Failed to get twap")
			return err
		}
		k.Logger(ctx).Info(fmt.Sprintf("%v TwapRate %v", hostZoneConfig.IbcDenom, twapRate))
		k.SetTwapRate(ctx, hostZoneConfig.IbcDenom, twapRate)

		// unfrozen hostzone
		err = k.UnFrozenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to unfrozen host zone %s", err.Error()))
			return err
		}

		k.Logger(ctx).Info("packet ICQ request successfully")

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		// frozen host zone
		k.Logger(ctx).Error(fmt.Sprintf("failed to send packet ICQ request %v", resp.Error))
		err := k.FrozenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to frozen host zone %s", err.Error()))
		}
		// resend query
		sequence, err := k.SendInterchainQuery(ctx, packet.GetData(), packet.SourcePort, packet.SourceChannel)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to resend ICQ on Ack Error %s", err.Error()))
			return err
		}
		// set sequence for ack or timeout
		params := k.GetParams(ctx)
		k.SetSendingPacketInfo(ctx, sequence, params.IbcQueryIcqChannel, hostZoneConfig)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
			),
		)
	}
	return nil
}

// OnTimeoutPacket resend packet when timeout
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet) error {
	// Get hostzone of send request
	hostZoneConfig := k.GetAndRemoveSendingPacketInfo(ctx, packet.Sequence, packet.SourceChannel)
	// resend query
	sequence, err := k.SendInterchainQuery(ctx, packet.GetData(), packet.SourcePort, packet.SourceChannel)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to resend ICQ on Timeout %s", err.Error()))
		return err
	}
	// set sequence for ack or timeout
	params := k.GetParams(ctx)
	k.SetSendingPacketInfo(ctx, sequence, params.IbcQueryIcqChannel, hostZoneConfig)
	return nil
}

func (k Keeper) GetChannelID(ctx sdk.Context) string {
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

func (k Keeper) transferOsmosisCrosschainSwap(ctx sdk.Context, hostChainConfig types.HostChainFeeAbsConfig) error {
	moduleAccountAddress := k.GetFeeAbsModuleAddress()
	token := k.bk.GetBalance(ctx, moduleAccountAddress, hostChainConfig.IbcDenom)
	params := k.GetParams(ctx)
	nativeDenomIBCedInOsmosis := params.NativeIbcedInOsmosis
	chainName := params.ChainName

	if !token.Amount.IsPositive() {
		return nil
	}

	memo, err := types.BuildCrossChainSwapMemo(nativeDenomIBCedInOsmosis, params.OsmosisCrosschainSwapAddress, moduleAccountAddress.String(), chainName)
	if err != nil {
		return err
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Minute * 5).UnixNano()

	transferMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    params.IbcTransferChannel,
		Token:            token,
		Sender:           moduleAccountAddress.String(),
		Receiver:         params.OsmosisCrosschainSwapAddress,
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

	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		// Will support pool route in the future so need array of reqs here
		var reqs []types.QueryArithmeticTwapToNowRequest
		req := types.NewQueryArithmeticTwapToNowRequest(
			hostZoneConfig.PoolId,
			params.NativeIbcedInOsmosis,
			hostZoneConfig.OsmosisPoolTokenDenomIn,
			startTime,
		)
		reqs = append(reqs, req)
		// send query
		sequence, err := k.SendOsmosisQueryRequest(ctx, reqs, types.IBCPortID, params.IbcQueryIcqChannel)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to query ICQ TWAP %s", err.Error()))
		}

		// set sequence for ack or timeout
		k.SetSendingPacketInfo(ctx, sequence, params.IbcQueryIcqChannel, hostZoneConfig)

		return false
	})

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

		err = k.transferOsmosisCrosschainSwap(ctx, hostZoneConfig)

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
