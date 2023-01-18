package keeper

import (
	"encoding/json"
	"errors"
	"fmt"

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

// Send request for query EstimateSwapExactAmountIn over IBC
func (k Keeper) SendOsmosisQueryRequest(ctx sdk.Context, poolId uint64, baseDenom string, quoteDenom string, sourcePort, sourceChannel string) error {
	packetData := types.NewOsmosisQueryRequestPacketData(poolId, baseDenom, quoteDenom)

	// Get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", sourcePort, sourceChannel,
		)
	}
	sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
	}

	destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
	destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

	timeoutHeight := clienttypes.NewHeight(0, 100000000)
	timeoutTimestamp := uint64(0)

	// Begin createOutgoingPacket logic
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	packetBytes := packetData.GetBytes()
	// Create the IBC packet
	packet := channeltypes.NewPacket(
		packetBytes,
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)

	// Send the IBC packet
	return k.channelKeeper.SendPacket(ctx, channelCap, packet)
}

// OnAcknowledgementIbcSwapAmountInRoute handle Acknowledgement for SwapAmountInRoute packet
func (k Keeper) OnAcknowledgementIbcOsmosisQueryRequest(ctx sdk.Context, ack channeltypes.Acknowledgement) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		_ = dispatchedAck.Error
		return nil
	case *channeltypes.Acknowledgement_Result:
		// Unmarshal dispatchedAck result
		spotPrice, err := k.UnmarshalPacketBytesToPrice(dispatchedAck.Result)
		if err != nil {
			return err
		}
		k.SetOsmosisExchangeRate(ctx, spotPrice)
		return nil
	default:
		// The counter-party module doesn't implement the correct acknowledgment format
		return errors.New("invalid acknowledgment format")
	}
}

// Send request for swap SwapAmountInRoute over IBC
func (k Keeper) SendIbcSwapAmountInRoute(
	ctx sdk.Context,
	poolId uint64,
	tokenOutDenom string,
	sourcePort string,
	sourceChannel string,
) error {
	packetData := types.NewSwapAmountInRoutePacketData(poolId, tokenOutDenom)

	// Get source channel endpoint
	sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
	}

	// Get counter-party chain endpoint infor
	destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
	destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

	// Get next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", sourcePort, sourceChannel,
		)
	}

	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	packetBytes := packetData.GetBytes()
	timeoutHeight := clienttypes.NewHeight(0, 100000000)
	timeoutTimestamp := uint64(0)

	// Create the IBC packet
	packet := channeltypes.NewPacket(
		packetBytes,
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)

	// Send the IBC packet
	return k.channelKeeper.SendPacket(ctx, channelCap, packet)
}

// OnAcknowledgementIbcSwapAmountInRoute handle Acknowledgement for SwapAmountInRoute packet
func (k Keeper) OnAcknowledgementIbcSwapAmountInRoute(ctx sdk.Context, ack channeltypes.Acknowledgement) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		_ = dispatchedAck.Error
		return nil
	case *channeltypes.Acknowledgement_Result:
		// Unmarshal dispatchedAck result

		// TODO: implement logic swap success
		return nil
	default:
		// The counter-party module doesn't implement the correct acknowledgment format
		return errors.New("invalid acknowledgment format")
	}
}

func (k Keeper) GetChannelId(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.KeyChannelID))
}

// TODO: need to test this function
func (k Keeper) UnmarshalPacketBytesToPrice(bz []byte) (sdk.Dec, error) {
	var spotPrice types.SpotPrice
	fmt.Println(string(bz))
	err := json.Unmarshal(bz, &spotPrice)
	if err != nil {
		return sdk.Dec{}, sdkerrors.New("ibc ack data umarshal", 1, "error when json.Unmarshal")
	}

	spotPriceDec, err := sdk.NewDecFromStr(spotPrice.SpotPrice)
	if err != nil {
		return sdk.Dec{}, sdkerrors.New("ibc ack data umarshal", 1, "error when NewDecFromStr")
	}
	return spotPriceDec, nil
}

// ParseMsgToMemo build a memo from msg, contractAddr, compatible with ValidateAndParseMemo in https://github.com/osmosis-labs/osmosis/blob/nicolas/crosschain-swaps-new/x/ibc-hooks/wasm_hook.go
func ParseMsgToMemo(msg types.OsmosisSwapMsg, contractAddr string, receiver string) (string, error) {
	// TODO: need to validate the msg && contract address
	memo := types.OsmosisSpecialMemo{
		Wasm: make(map[string]interface{}),
	}

	memo.Wasm["contract"] = contractAddr
	memo.Wasm["msg"] = msg
	memo.Wasm["receiver"] = receiver

	memo_marshalled, err := json.Marshal(&memo)
	if err != nil {
		return "", nil
	}
	return string(memo_marshalled), nil
}

func (k Keeper) transferIBCTokenToOsmosisContract(ctx sdk.Context) error {
	params := k.GetParams(ctx)

	moduleAccountAddress := k.GetModuleAddress()
	token := k.bk.GetBalance(ctx, moduleAccountAddress, params.OsmosisIbcDenom)

	// if token
	if sdk.NewInt(1).GTE(token.Amount) {
		return nil
	}

	memo, err := buildMemo(sdk.NewCoin("uosmo", token.Amount), params.NativeIbcDenom, params.OsmosisSwapContract, moduleAccountAddress.String())
	if err != nil {
		return err
	}

	transferMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    params.OsmosisTransferChannel,
		Token:            token,
		Sender:           moduleAccountAddress.String(),
		Receiver:         params.OsmosisSwapContract,
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

func buildMemo(inputToken sdk.Coin, outputDenom string, contractAddress, receiver string) (string, error) {
	swap := types.Swap{
		InputCoin:   inputToken,
		OutPutDenom: outputDenom,
		Slippage: types.Twap{
			Twap: types.TwapRouter{
				SlippagePercentage: "20",
				WindowSeconds:      10,
			},
		},
		Receiver: receiver,
	}

	msgSwap := types.OsmosisSwapMsg{
		OsmosisSwap: swap,
	}
	return ParseMsgToMemo(msgSwap, contractAddress, receiver)
}

func (k Keeper) executeTransferMsg(ctx sdk.Context, transferMsg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error) {
	if err := transferMsg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("bad msg %v", err.Error())
	}
	return k.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), transferMsg)

}

// TODO: use TWAP instead of spotprice
func (k Keeper) handleOsmosisIbcQuery(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	channelID := params.OsmosisQueryChannel
	poolId := params.PoolId // for testing
	baseDenom := params.NativeIbcDenom

	return k.SendOsmosisQueryRequest(ctx, poolId, baseDenom, "uosmo", types.IBCPortID, channelID)
}
