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

const (
	timeoutDuration = 5 * time.Minute
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
	icqReqs := make([]abci.RequestQuery, len(twapReqs))
	for i, req := range twapReqs {
		req := req
		icqReqs[i] = abci.RequestQuery{
			Path: params.OsmosisQueryTwapPath,
			Data: k.cdc.MustMarshal(&req),
		}
	}
	k.Logger(ctx).Info("SendOsmosisQueryRequest", "num_requests", len(icqReqs), "sourcePort", sourcePort, "sourceChannel", sourceChannel)
	_, err := k.SendInterchainQuery(ctx, icqReqs, sourcePort, sourceChannel)
	if err != nil {
		k.Logger(ctx).Error("SendOsmosisQueryRequest: error when send interchain query", "err", err)
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
	timeoutTimestamp := ctx.BlockTime().Add(timeoutDuration).UnixNano()
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return 0, sdkerrors.Wrapf(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability: source_port: %s, source_channel: %s", sourcePort, sourceChannel)
	}

	data, err := types.SerializeCosmosQuery(reqs)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "could not serialize reqs into cosmos query")
	}
	icqPacketData := types.NewInterchainQueryPacketData(data, "")

	sequence, err := k.channelKeeper.SendPacket(ctx, channelCap, sourcePort, sourceChannel, clienttypes.ZeroHeight(), uint64(timeoutTimestamp), icqPacketData.GetBytes())
	if err != nil {
		k.Logger(ctx).Error("SendInterchainQuery: SendPacket failed", "err", err)
		return 0, err
	}

	k.Logger(ctx).Info("SendInterchainQuery: ", "sequence", sequence)

	return sequence, nil
}

func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, ack channeltypes.Acknowledgement, icqReqs []abci.RequestQuery) error {
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

		for i, icqReq := range icqReqs {
			var icqReqData types.QueryArithmeticTwapToNowRequest
			if err := k.cdc.Unmarshal(icqReq.GetData(), &icqReqData); err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to unmarshal icqReqData %s", err.Error()))
				continue
			}

			// get chain config
			hostZoneConfig, found := k.GetHostZoneConfigByOsmosisTokenDenom(ctx, icqReqData.QuoteAsset)
			if !found {
				k.Logger(ctx).Error(fmt.Sprintf("Error when get host zone by Osmosis denom %s %v not found", icqReqData.QuoteAsset, err))
				fmt.Println("error")
				continue
			}

			icqRes := icqResponses[i]

			if icqRes.Code != 0 {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to send interchain query code %d", icqRes.Code))
				k.IncreaseBlockDelayToQuery(ctx, hostZoneConfig.IbcDenom)
				continue
			}
			// k.Logger(ctx).Info(fmt.Sprintf("ICQ response %+v", icqRes))
			// Not sure why, but the value is unmarshalled to icqRes.Key instead of icqRes.Value
			// 10:36AM INF ICQ response {Code:0 Log: Info: Index:0 Key:[10 19 50 49 52 50 56 53 55 49 52 48 48 48 48 48 48 48 48 48 48] Value:[] ProofOps:<nil> Height:0 Codespace:}
			twapRate, err := k.GetDecTWAPFromBytes(icqRes.Key)
			if err != nil {
				k.Logger(ctx).Error("Failed to get twap")
				continue
			}
			k.Logger(ctx).Info(fmt.Sprintf("TwapRate %v", twapRate))
			k.SetTwapRate(ctx, hostZoneConfig.IbcDenom, twapRate)

			err = k.SetStateHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom, types.HostChainFeeAbsStatus_UPDATED)
			if err != nil {
				// should never happen
				k.Logger(ctx).Error(fmt.Sprintf("Failed to frozen host zone %s", err.Error()))
			}

			// reset block delay to query
			k.ResetBlockDelayToQuery(ctx, hostZoneConfig.IbcDenom)
		}

		k.Logger(ctx).Info("packet ICQ request successfully")

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
			// todo: should try to retry here instead of setting FROZEN
			k.IncreaseBlockDelayToQuery(ctx, hostZoneConfig.IbcDenom)

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

// OnTimeoutPacket resend packet when timeout
func (k Keeper) OnTimeoutPacket(ctx sdk.Context) error {
	ctx.Logger().Info("IBC Timeout packet")
	_, err := k.HandleOsmosisIbcQuery(ctx)
	return err
}

func (k Keeper) GetChannelID(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.KeyChannelID))
}

func (k Keeper) GetDecTWAPFromBytes(bz []byte) (sdk.Dec, error) {
	if bz == nil {
		return sdk.Dec{}, sdkerrors.New("GetDecTWAPFromBytes: err ", 1, "nil bytes")
	}
	var ibcTokenTwap types.QueryArithmeticTwapToNowResponse
	err := k.cdc.Unmarshal(bz, &ibcTokenTwap)
	if err != nil || ibcTokenTwap.ArithmeticTwap.IsNil() {
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

	if !token.Amount.IsPositive() || token.Amount.LT(sdk.NewIntFromUint64(hostChainConfig.MinSwapAmount)) {
		return fmt.Errorf("invalid amount to transfer, expect minimum %v, got %v", hostChainConfig.MinSwapAmount, token.Amount)
	}

	memo, err := types.BuildCrossChainSwapMemo(nativeDenomIBCedInOsmosis, params.OsmosisCrosschainSwapAddress, moduleAccountAddress.String(), chainName)
	if err != nil {
		return err
	}

	timeoutTimestamp := ctx.BlockTime().Add(timeoutDuration).UnixNano()

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

func (k Keeper) HandleOsmosisIbcQuery(ctx sdk.Context) (int, error) {
	// set startTime for query twap
	queryTwapEpochInfo, found := k.GetEpochInfo(ctx, types.DefaultQueryEpochIdentifier)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Don't have query epoch information: %s", types.DefaultQueryEpochIdentifier))
		return 0, nil
	}
	startTime := ctx.BlockTime().Add(-queryTwapEpochInfo.Duration)
	k.Logger(ctx).Info(fmt.Sprintf("Start time: %v", startTime.Unix()))

	params := k.GetParams(ctx)

	batchSize := 10
	var reqs []types.QueryArithmeticTwapToNowRequest
	batchCounter := 0
	var errorFound error
	// fee abstraction will not send query to a frozen host zone
	// however, it will continue to send query to other host zone if UPDATED, or OUTDATED
	// this logic iterate through registered host zones and collect requests before sending it
	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		if k.IbcQueryHostZoneFilter(ctx, hostZoneConfig, queryTwapEpochInfo) {
			return false
		}

		req := types.NewQueryArithmeticTwapToNowRequest(
			hostZoneConfig.PoolId,
			params.NativeIbcedInOsmosis,
			hostZoneConfig.OsmosisPoolTokenDenomIn,
			startTime,
		)
		k.Logger(ctx).Info("handleOsmosisIbcQuery: NewQueryArithmeticTwapToNowRequest", "req", fmt.Sprintf("%+v", req))
		reqs = append(reqs, req)
		batchCounter++
		if batchCounter == batchSize {
			err := k.SendOsmosisQueryRequest(ctx, reqs, types.IBCPortID, params.IbcQueryIcqChannel)
			if err != nil {
				errorFound = err
				return true
			}
			reqs = []types.QueryArithmeticTwapToNowRequest{}
			batchCounter = 0
		}
		return false
	})

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

		if err := sdk.ValidateDenom(hostZoneConfig.IbcDenom); err != nil {
			k.Logger(ctx).Error("executeAllHostChainSwap: invalid ibc denom", "denom", hostZoneConfig.IbcDenom, "err", err)
			return false
		}

		err = k.transferOsmosisCrosschainSwap(ctx, hostZoneConfig)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to transfer IBC token %s", err.Error()))
			// should be set to OUTDATED if failed to transfer to preserve funds
			// if the newest twap query successes, it will be set to UPDATED
			err = k.SetStateHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom, types.HostChainFeeAbsStatus_OUTDATED)
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

// for testing
func (k Keeper) TransferOsmosisCrosschainSwap(ctx sdk.Context, hostChainConfig types.HostChainFeeAbsConfig) error {
	return k.transferOsmosisCrosschainSwap(ctx, hostChainConfig)
}
