package ibctesting

import (
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"
	"github.com/golang/protobuf/proto" //nolint
	feeabs "github.com/notional-labs/feeabstraction/v2/app"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
)

var wasmIdent = []byte("\x00\x61\x73\x6D")

// SeedNewContractInstance stores some wasm code and instantiates a new contract on this chain.
// This method can be called to prepare the store with some valid CodeInfo and ContractInfo. The returned
// Address is the contract address for this instance. Test should make use of this data and/or use NewIBCContractMockWasmer
// for using a contract mock in Go.
func (chain *TestChain) SeedNewContractInstance() sdk.AccAddress {
	pInstResp := chain.StoreCode(append(wasmIdent, rand.Bytes(10)...))
	codeID := pInstResp.CodeID

	anyAddressStr := chain.SenderAccount.GetAddress().String()
	initMsg := []byte(fmt.Sprintf(`{"verifier": %q, "beneficiary": %q}`, anyAddressStr, anyAddressStr))
	return chain.InstantiateContract(codeID, initMsg)
}

func (chain *TestChain) StoreCode(byteCode []byte) types.MsgStoreCodeResponse {
	storeMsg := &types.MsgStoreCode{
		Sender:       chain.SenderAccount.GetAddress().String(),
		WASMByteCode: byteCode,
	}
	r, err := chain.SendMsgs(storeMsg)
	require.NoError(chain.t, err)
	protoResult := chain.parseSDKResultData(r)
	require.Len(chain.t, protoResult.Data, 1)
	// unmarshal protobuf response from data
	var pInstResp types.MsgStoreCodeResponse
	require.NoError(chain.t, pInstResp.Unmarshal(protoResult.Data[0].Data))
	require.NotEmpty(chain.t, pInstResp.CodeID)
	return pInstResp
}

func (chain *TestChain) InstantiateContract(codeID uint64, initMsg []byte) sdk.AccAddress {
	instantiateMsg := &types.MsgInstantiateContract{
		Sender: chain.SenderAccount.GetAddress().String(),
		Admin:  chain.SenderAccount.GetAddress().String(),
		CodeID: codeID,
		Label:  "ibc-test",
		Msg:    initMsg,
		Funds:  sdk.Coins{ibctesting.TestCoin},
	}

	r, err := chain.SendMsgs(instantiateMsg)
	require.NoError(chain.t, err)
	protoResult := chain.parseSDKResultData(r)
	require.Len(chain.t, protoResult.Data, 1)

	var pExecResp types.MsgInstantiateContractResponse
	require.NoError(chain.t, pExecResp.Unmarshal(protoResult.Data[0].Data))
	a, err := sdk.AccAddressFromBech32(pExecResp.Address)
	require.NoError(chain.t, err)
	return a
}

func (chain *TestChain) parseSDKResultData(r *sdk.Result) sdk.TxMsgData {
	var protoResult sdk.TxMsgData
	require.NoError(chain.t, proto.Unmarshal(r.Data, &protoResult))
	return protoResult
}

// ContractInfo is a helper function to returns the ContractInfo for the given contract address
func (chain *TestChain) ContractInfo(contractAddr sdk.AccAddress) *types.ContractInfo {
	type testSupporter interface {
		TestSupport() *feeabs.TestSupport
	}
	return chain.App.(testSupporter).TestSupport().WasmKeeper().GetContractInfo(chain.GetContext(), contractAddr)
}
