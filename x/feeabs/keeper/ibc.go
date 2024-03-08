package keeper

import (
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/osmosis-labs/fee-abstraction/v4/x/feeabs/types"
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

	sequence, err := k.channelKeeper.SendPacket(ctx, channelCap, sourcePort, sourceChannel, clienttypes.ZeroHeight(), uint64(timeoutTimestamp), icqPacketData.GetBytes())
	if err != nil {
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
<<<<<<< HEAD
				// if not found any request, end the iterator
				return true
=======
				k.Logger(ctx).Error(fmt.Sprintf("Error when get host zone by Osmosis denom %s %v not found", icqReqData.QuoteAsset, err))
				fmt.Println("error")
				continue
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
			}
			// Check if icq TWAP denom match with hostzone denom store
			if icqReqData.QuoteAsset != hostZoneConfig.OsmosisPoolTokenDenomIn {
				return false
			}
			// Get icq QueryArithmeticTwapToNowRequest response
			IcqRes := ICQResponses[index]
			index++

<<<<<<< HEAD
			if IcqRes.Code != 0 {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to send interchain query code %d", IcqRes.Code))
				err := k.FrozenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
				if err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("Failed to frozen host zone %s", err.Error()))
				}
				return false
=======
			icqRes := icqResponses[i]

			if icqRes.Code != 0 {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to send interchain query code %d", icqRes.Code))
				k.IncreaseBlockDelayToQuery(ctx, hostZoneConfig.IbcDenom)
				continue
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
			}

			twapRate, err := k.GetDecTWAPFromBytes(IcqRes.Value)
			if err != nil {
				k.Logger(ctx).Error("Failed to get twap")
				return false
			}
			k.Logger(ctx).Info(fmt.Sprintf("TwapRate %v", twapRate))
			k.SetTwapRate(ctx, hostZoneConfig.IbcDenom, twapRate)

<<<<<<< HEAD
			return false
		})
=======
			err = k.SetStateHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom, types.HostChainFeeAbsStatus_UPDATED)
			if err != nil {
				// should never happen
				k.Logger(ctx).Error(fmt.Sprintf("Failed to frozen host zone %s", err.Error()))
			}

			// reset block delay to query
			k.ResetBlockDelayToQuery(ctx, hostZoneConfig.IbcDenom)
		}
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))

		k.Logger(ctx).Info("packet ICQ request successfully")

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
<<<<<<< HEAD
			err := k.FrozenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to frozen host zone %s", err.Error()))
			}
=======
			// todo: should try to retry here instead of setting FROZEN
			k.IncreaseBlockDelayToQuery(ctx, hostZoneConfig.IbcDenom)
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))

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

<<<<<<< HEAD
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
=======
// OnTimeoutPacket resend packet when timeout
func (k Keeper) OnTimeoutPacket(ctx sdk.Context) error {
	ctx.Logger().Info("IBC Timeout packet")
	_, err := k.HandleOsmosisIbcQuery(ctx)
	return err
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
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

<<<<<<< HEAD
func (k Keeper) handleOsmosisIbcQuery(ctx sdk.Context) error {
	hasQueryEpochInfo := k.HasEpochInfo(ctx, types.DefaultQueryEpochIdentifier)
	if !hasQueryEpochInfo {
=======
func (k Keeper) HandleOsmosisIbcQuery(ctx sdk.Context) (int, error) {
	// set startTime for query twap
	queryTwapEpochInfo, found := k.GetEpochInfo(ctx, types.DefaultQueryEpochIdentifier)
	if !found {
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
		k.Logger(ctx).Error(fmt.Sprintf("Don't have query epoch information: %s", types.DefaultQueryEpochIdentifier))
		return 0, nil
	}

	// set startTime for query twap
	queryTwapEpochInfo := k.GetEpochInfo(ctx, types.DefaultQueryEpochIdentifier)
	startTime := ctx.BlockTime().Add(-queryTwapEpochInfo.Duration)
	k.Logger(ctx).Info(fmt.Sprintf("Start time: %v", startTime.Unix()))

	params := k.GetParams(ctx)

	var reqs []types.QueryArithmeticTwapToNowRequest
<<<<<<< HEAD
	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
=======
	batchCounter := 0
	var errorFound error
	// fee abstraction will not send query to a frozen host zone
	// however, it will continue to send query to other host zone if UPDATED, or OUTDATED
	// this logic iterate through registered host zones and collect requests before sending it
	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		if k.IbcQueryHostZoneFilter(ctx, hostZoneConfig, queryTwapEpochInfo) {
			return false
		}

>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
		req := types.NewQueryArithmeticTwapToNowRequest(
			hostZoneConfig.PoolId,
			params.NativeIbcedInOsmosis,
			hostZoneConfig.OsmosisPoolTokenDenomIn,
			startTime,
		)
		reqs = append(reqs, req)
		return false
	})
<<<<<<< HEAD
	err := k.SendOsmosisQueryRequest(ctx, reqs, types.IBCPortID, params.IbcQueryIcqChannel)
	if err != nil {
		return err
	}

	return nil
=======

	if errorFound != nil {
		return 0, errorFound
	}

	if len(reqs) > 0 {
		k.Logger(ctx).Info("handleOsmosisIbcQuery", "requests", len(reqs))
		err := k.SendOsmosisQueryRequest(ctx, reqs, types.IBCPortID, params.IbcQueryIcqChannel)
		if err != nil {
			k.Logger(ctx).Error("handleOsmosisIbcQuery: SendOsmosisQueryRequest failed", "err", err)
			return 0, err
		}
	} else {
		k.Logger(ctx).Info("handleOsmosisIbcQuery: no requests")
	}
	return len(reqs), nil
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
}

// executeAllHostChainTWAPQuery will iterate all hostZone and send the IBC Query Packet to Osmosis chain.
func (k Keeper) ExecuteAllHostChainTWAPQuery(ctx sdk.Context) {
	_, err := k.HandleOsmosisIbcQuery(ctx)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error executeAllHostChainTWAPQuery %s", err.Error()))
	}
}

// executeAllHostChainTWAPSwap will iterate all hostZone and execute swap over chain.
// If the hostZone is frozen, it will not execute the swap.
func (k Keeper) ExecuteAllHostChainSwap(ctx sdk.Context) {
	// should only execute swap when the host zone is not frozen
	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		var err error

		if hostZoneConfig.Status == types.HostChainFeeAbsStatus_FROZEN {
			return false
		}

		// if the host zone is outdated, it should not execute swap
		if hostZoneConfig.Status == types.HostChainFeeAbsStatus_OUTDATED {
			return false
		}

		err = k.transferOsmosisCrosschainSwap(ctx, hostZoneConfig)

		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to transfer IBC token %s", err.Error()))
<<<<<<< HEAD
			err = k.FrozenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
=======
			// should be set to OUTDATED if failed to transfer to preserve funds
			// if the newest twap query successes, it will be set to UPDATED
			err = k.SetStateHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom, types.HostChainFeeAbsStatus_OUTDATED)
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to frozem host zone %s", err.Error()))
			}
		}

		return false
	})
}

func (k Keeper) IbcQueryHostZoneFilter(ctx sdk.Context, hostZoneConfig types.HostChainFeeAbsConfig, queryTwapEpochInfo types.EpochInfo) bool {
	if hostZoneConfig.Status == types.HostChainFeeAbsStatus_FROZEN {
		return true
	}

	// determine what host zone gets to query
	exponential := k.GetBlockDelayToQuery(ctx, hostZoneConfig.IbcDenom)
	if exponential.Jump == types.ExponentialOutdatedJump {
		err := k.SetStateHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom, types.HostChainFeeAbsStatus_OUTDATED)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to set host zone status %s", err.Error()))
		}
	}

	if queryTwapEpochInfo.CurrentEpoch < exponential.FutureEpoch {
		return true
	}

	return false
}
