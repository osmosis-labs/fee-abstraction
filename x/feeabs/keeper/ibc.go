package keeper

import (
	"encoding/json"
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

	// Create the IBC packet
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

	// Send the IBC packet
	return k.channelKeeper.SendPacket(ctx, channelCap, packet)
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

// TODO: clean up
func (k Keeper) transferIBCTokenToOsmosisContract(ctx sdk.Context, token sdk.Coin) error {
	moduleAddress, _ := sdk.AccAddressFromBech32("cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr")
	sourcePort := transfertypes.PortID
	// TODO: channel for IBC transfer should set in params
	sourceChannel := "channel-0"

	// TODO: contractAddress should be store in paramspace
	crossChainSwapContractAddress := "osmo1j08452mqwadp8xu25kn9rleyl2gufgfjnv0sn8dvynynakkjukcqq33hql"
	timeoutHeight := clienttypes.NewHeight(0, 100000000)
	timeoutTimestamp := uint64(0)
	token = sdk.NewCoin("ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518", sdk.NewInt(1000000))

	tokenInOsmosisSwap := sdk.NewCoin("uosmo", sdk.NewInt(1000000))
	memo, _ := buildMemo(tokenInOsmosisSwap, "ibc/4D74FBE09BED153381B75FF0D0B030A839E68AE17761F3945A8AF5671B915928", crossChainSwapContractAddress, moduleAddress.String())

	fmt.Println("============memo===============")
	fmt.Println(memo)
	fmt.Println("============memo===============")

	transferMsg := transfertypes.MsgTransfer{
		SourcePort:       sourcePort,
		SourceChannel:    sourceChannel,
		Token:            token,
		Sender:           moduleAddress.String(),
		Receiver:         crossChainSwapContractAddress,
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             memo,
	}

	response, err := k.executeTransferMsg(ctx, &transferMsg)
	if err != nil {
		return err
	}

	fmt.Println(response)
	return err
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
		return nil, fmt.Errorf("bad msg", err.Error())
	}
	return k.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), transferMsg)

}
