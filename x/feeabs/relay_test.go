package feeabs_test

import (
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	wasmibctesting "github.com/notional-labs/feeabstraction/v1/x/feeabs/ibctesting"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"

	wasmvm "github.com/CosmWasm/wasmvm"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeeAbsIBCToContract(t *testing.T) {
	specs := map[string]struct {
		contract      wasmtesting.IBCContractCallbacks
		setupContract func(t *testing.T, contract wasmtesting.IBCContractCallbacks, chain *wasmibctesting.TestChain)
	}{
		"query": {
			contract: &queryFeeabsContract{},
			setupContract: func(t *testing.T, contract wasmtesting.IBCContractCallbacks, chain *wasmibctesting.TestChain) {
				c := contract.(*queryFeeabsContract)
				c.t = t
				c.chain = chain
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var (
				chainAOpts = []wasmkeeper.Option{wasmkeeper.WithWasmEngine(
					wasmtesting.NewIBCContractMockWasmer(spec.contract),
				)}
				coordinator = wasmibctesting.NewCoordinator(t, 2, []wasmkeeper.Option{}, chainAOpts)
				chainA      = coordinator.GetChain(wasmibctesting.GetChainID(0))
				chainB      = coordinator.GetChain(wasmibctesting.GetChainID(1))
			)

			coordinator.CommitBlock(chainA, chainB)
			myContractAddr := chainB.SeedNewContractInstance()
			contractBPortID := chainB.ContractInfo(myContractAddr).IBCPortID

			spec.setupContract(t, spec.contract, chainB)

			path := wasmibctesting.NewPath(chainA, chainB)
			path.EndpointA.ChannelConfig = &ibctesting.ChannelConfig{
				PortID:  "feeabs",
				Version: "",
				Order:   channeltypes.UNORDERED,
			}
			path.EndpointB.ChannelConfig = &ibctesting.ChannelConfig{
				PortID:  contractBPortID,
				Version: "",
				Order:   channeltypes.UNORDERED,
			}

			coordinator.SetupConnections(path)
			coordinator.CreateChannels(path)

			// set params
			params := chainA.GetTestSupport().FeeAbsKeeper().GetParams(chainA.GetContext())
			params.NativeIbcedInOsmosis = "denom"
			chainA.GetTestSupport().FeeAbsKeeper().SetParams(chainA.GetContext(), params)

			// set hostzone config
			hostZoneConfig := types.HostChainFeeAbsConfig{
				IbcDenom:                "ibc/denom",
				OsmosisPoolTokenDenomIn: "denom",
				OsmosisQueryChannel:     path.EndpointA.ChannelID,
			}
			err := chainA.GetTestSupport().FeeAbsKeeper().SetHostZoneConfig(chainA.GetContext(), "ibc/denom", hostZoneConfig)
			require.NoError(t, err)

			msg := types.NewMsgSendQueryIbcDenomTWAP(
				chainA.SenderAccount.GetAddress(),
				"ibc/denom",
				time.Now().UTC(),
			)
			_, err = chainA.SendMsgs(msg)
			require.NoError(t, err)
			require.NoError(t, path.EndpointB.UpdateClient())

			// then
			require.Equal(t, 1, len(chainA.PendingSendPackets))
			require.Equal(t, 0, len(chainB.PendingSendPackets))

			// and when relay to chain B and handle Ack on chain A
			err = coordinator.RelayAndAckPendingPackets(path)
			require.NoError(t, err)

			// then
			require.Equal(t, 0, len(chainA.PendingSendPackets))
			require.Equal(t, 0, len(chainB.PendingSendPackets))

			expectedTwapPrice, err := sdk.NewDecFromStr("2.0")
			require.NoError(t, err)
			twapPrice, err := chainA.GetTestSupport().FeeAbsKeeper().GetTwapRate(chainA.GetContext(), "ibc/denom")
			require.NoError(t, err)
			require.Equal(t, expectedTwapPrice, twapPrice)
		})
	}
}

func TestFromIBCTransferToContract(t *testing.T) {
	// scenario: given two chains,
	//           with a contract on chain B
	//           then the contract can handle the receiving side of an ics20 transfer
	//           that was started on chain A via ibc transfer module

	transferAmount := sdk.NewInt(1)
	specs := map[string]struct {
		contract             wasmtesting.IBCContractCallbacks
		setupContract        func(t *testing.T, contract wasmtesting.IBCContractCallbacks, chain *wasmibctesting.TestChain)
		expChainABalanceDiff sdk.Int
		expChainBBalanceDiff sdk.Int
	}{
		"ack": {
			contract: &ackReceiverContract{},
			setupContract: func(t *testing.T, contract wasmtesting.IBCContractCallbacks, chain *wasmibctesting.TestChain) {
				c := contract.(*ackReceiverContract)
				c.t = t
				c.chain = chain
			},
			expChainABalanceDiff: transferAmount.Neg(),
			expChainBBalanceDiff: transferAmount,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var (
				chainAOpts = []wasmkeeper.Option{wasmkeeper.WithWasmEngine(
					wasmtesting.NewIBCContractMockWasmer(spec.contract),
				)}
				coordinator = wasmibctesting.NewCoordinator(t, 2, []wasmkeeper.Option{}, chainAOpts)
				chainA      = coordinator.GetChain(wasmibctesting.GetChainID(0))
				chainB      = coordinator.GetChain(wasmibctesting.GetChainID(1))
			)
			coordinator.CommitBlock(chainA, chainB)
			myContractAddr := chainB.SeedNewContractInstance()
			contractBPortID := chainB.ContractInfo(myContractAddr).IBCPortID

			spec.setupContract(t, spec.contract, chainB)

			path := wasmibctesting.NewPath(chainA, chainB)
			path.EndpointA.ChannelConfig = &ibctesting.ChannelConfig{
				PortID:  "transfer",
				Version: ibctransfertypes.Version,
				Order:   channeltypes.UNORDERED,
			}
			path.EndpointB.ChannelConfig = &ibctesting.ChannelConfig{
				PortID:  contractBPortID,
				Version: ibctransfertypes.Version,
				Order:   channeltypes.UNORDERED,
			}

			coordinator.SetupConnections(path)
			coordinator.CreateChannels(path)

			originalChainABalance := chainA.Balance(chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
			// when transfer via sdk transfer from A (module) -> B (contract)
			coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
			timeoutHeight := clienttypes.NewHeight(1, 110)
			msg := ibctransfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, chainA.SenderAccount.GetAddress().String(), chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0)
			_, err := chainA.SendMsgs(msg)
			require.NoError(t, err)
			require.NoError(t, path.EndpointB.UpdateClient())

			// then
			require.Equal(t, 1, len(chainA.PendingSendPackets))
			require.Equal(t, 0, len(chainB.PendingSendPackets))

			// and when relay to chain B and handle Ack on chain A
			err = coordinator.RelayAndAckPendingPackets(path)
			require.NoError(t, err)

			// then
			require.Equal(t, 0, len(chainA.PendingSendPackets))
			require.Equal(t, 0, len(chainB.PendingSendPackets))

			// and source chain balance was decreased
			newChainABalance := chainA.Balance(chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
			assert.Equal(t, originalChainABalance.Amount.Add(spec.expChainABalanceDiff), newChainABalance.Amount)

			// and dest chain balance contains voucher
			expBalance := ibctransfertypes.GetTransferCoin(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, coinToSendToB.Denom, spec.expChainBBalanceDiff)
			gotBalance := chainB.Balance(chainB.SenderAccount.GetAddress(), expBalance.Denom)
			assert.Equal(t, expBalance, gotBalance, "got total balance: %s", chainB.AllBalances(chainB.SenderAccount.GetAddress()))
		})
	}
}

var _ wasmtesting.IBCContractCallbacks = &queryFeeabsContract{}

// contract that acts as the receiving side for a query feeabs
type queryFeeabsContract struct {
	contractStub
	t     *testing.T
	chain *wasmibctesting.TestChain
}

func (c *queryFeeabsContract) IBCPacketReceive(
	codeID wasmvm.Checksum,
	env wasmvmtypes.Env,
	msg wasmvmtypes.IBCPacketReceiveMsg,
	store wasmvm.KVStore,
	goapi wasmvm.GoAPI,
	querier wasmvm.Querier,
	gasMeter wasmvm.GasMeter,
	gasLimit uint64,
	deserCost wasmvmtypes.UFraction,
) (*wasmvmtypes.IBCReceiveResult, uint64, error) {
	result := `{"responses":[{"success":true,"data":"eyJhcml0aG1ldGljX3R3YXAiOiIyLjAwMDAwMDAwMDAwMDAwMDAwMCJ9"}]}`
	ack := channeltypes.NewResultAcknowledgement([]byte(result)).Acknowledgement()
	var log []wasmvmtypes.EventAttribute
	return &wasmvmtypes.IBCReceiveResult{Ok: &wasmvmtypes.IBCReceiveResponse{Acknowledgement: ack, Attributes: log}}, 0, nil
}

func (c *queryFeeabsContract) IBCPacketAck(
	codeID wasmvm.Checksum,
	env wasmvmtypes.Env,
	msg wasmvmtypes.IBCPacketAckMsg,
	store wasmvm.KVStore,
	goapi wasmvm.GoAPI,
	querier wasmvm.Querier,
	gasMeter wasmvm.GasMeter,
	gasLimit uint64,
	deserCost wasmvmtypes.UFraction,
) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	return &wasmvmtypes.IBCBasicResponse{}, 0, nil
}

var _ wasmtesting.IBCContractCallbacks = &ackReceiverContract{}

// contract that acts as the receiving side for an ics-20 transfer.
type ackReceiverContract struct {
	contractStub
	t     *testing.T
	chain *wasmibctesting.TestChain
}

func (c *ackReceiverContract) IBCPacketReceive(codeID wasmvm.Checksum, env wasmvmtypes.Env, msg wasmvmtypes.IBCPacketReceiveMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCReceiveResult, uint64, error) {
	packet := msg.Packet

	var src ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.Data, &src); err != nil {
		return nil, 0, err
	}
	require.NoError(c.t, src.ValidateBasic())

	// call original ibctransfer keeper to not copy all code into this
	ibcPacket := toIBCPacket(packet)
	ctx := c.chain.GetContext() // HACK: please note that this is not reverted after checkTX
	err := c.chain.GetTestSupport().TransferKeeper().OnRecvPacket(ctx, ibcPacket, src)
	if err != nil {
		return nil, 0, sdkerrors.Wrap(err, "within our smart contract")
	}

	var log []wasmvmtypes.EventAttribute // note: all events are under `wasm` event type
	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)}).Acknowledgement()
	return &wasmvmtypes.IBCReceiveResult{Ok: &wasmvmtypes.IBCReceiveResponse{Acknowledgement: ack, Attributes: log}}, 0, nil
}

func (c *ackReceiverContract) IBCPacketAck(codeID wasmvm.Checksum, env wasmvmtypes.Env, msg wasmvmtypes.IBCPacketAckMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(msg.OriginalPacket.Data, &data); err != nil {
		return nil, 0, err
	}
	// call original ibctransfer keeper to not copy all code into this

	var ack channeltypes.Acknowledgement
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(msg.Acknowledgement.Data, &ack); err != nil {
		return nil, 0, err
	}

	// call original ibctransfer keeper to not copy all code into this
	ctx := c.chain.GetContext() // HACK: please note that this is not reverted after checkTX
	ibcPacket := toIBCPacket(msg.OriginalPacket)
	err := c.chain.GetTestSupport().TransferKeeper().OnAcknowledgementPacket(ctx, ibcPacket, data, ack)
	if err != nil {
		return nil, 0, sdkerrors.Wrap(err, "within our smart contract")
	}

	return &wasmvmtypes.IBCBasicResponse{}, 0, nil
}

// simple helper struct that implements connection setup methods.
type contractStub struct{}

func (s *contractStub) IBCChannelOpen(codeID wasmvm.Checksum, env wasmvmtypes.Env, msg wasmvmtypes.IBCChannelOpenMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBC3ChannelOpenResponse, uint64, error) {
	return &wasmvmtypes.IBC3ChannelOpenResponse{}, 0, nil
}

func (s *contractStub) IBCChannelConnect(codeID wasmvm.Checksum, env wasmvmtypes.Env, msg wasmvmtypes.IBCChannelConnectMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	return &wasmvmtypes.IBCBasicResponse{}, 0, nil
}

func (s *contractStub) IBCChannelClose(codeID wasmvm.Checksum, env wasmvmtypes.Env, msg wasmvmtypes.IBCChannelCloseMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	panic("implement me")
}

func (s *contractStub) IBCPacketReceive(codeID wasmvm.Checksum, env wasmvmtypes.Env, msg wasmvmtypes.IBCPacketReceiveMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCReceiveResult, uint64, error) {
	panic("implement me")
}

func (s *contractStub) IBCPacketAck(codeID wasmvm.Checksum, env wasmvmtypes.Env, msg wasmvmtypes.IBCPacketAckMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	return &wasmvmtypes.IBCBasicResponse{}, 0, nil
}

func (s *contractStub) IBCPacketTimeout(codeID wasmvm.Checksum, env wasmvmtypes.Env, msg wasmvmtypes.IBCPacketTimeoutMsg, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.IBCBasicResponse, uint64, error) {
	panic("implement me")
}

func toIBCPacket(p wasmvmtypes.IBCPacket) channeltypes.Packet {
	var height clienttypes.Height
	if p.Timeout.Block != nil {
		height = clienttypes.NewHeight(p.Timeout.Block.Revision, p.Timeout.Block.Height)
	}
	return channeltypes.Packet{
		Sequence:           p.Sequence,
		SourcePort:         p.Src.PortID,
		SourceChannel:      p.Src.ChannelID,
		DestinationPort:    p.Dest.PortID,
		DestinationChannel: p.Dest.ChannelID,
		Data:               p.Data,
		TimeoutHeight:      height,
		TimeoutTimestamp:   p.Timeout.Timestamp,
	}
}
