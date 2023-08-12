package ibctesting

import (
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	feeabs "github.com/osmosis-labs/fee-abstraction/v4/app"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/libs/rand"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
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
	require.Len(chain.t, r.MsgResponses, 1)
	require.NotEmpty(chain.t, r.MsgResponses[0].GetCachedValue())
	// unmarshal protobuf response from data
	pInstResp := r.MsgResponses[0].GetCachedValue().(*types.MsgStoreCodeResponse)
	require.NotEmpty(chain.t, pInstResp.CodeID)
	require.NotEmpty(chain.t, pInstResp.Checksum)
	return *pInstResp
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
	require.Len(chain.t, r.MsgResponses, 1)
	require.NotEmpty(chain.t, r.MsgResponses[0].GetCachedValue())
	pExecResp := r.MsgResponses[0].GetCachedValue().(*types.MsgInstantiateContractResponse)
	a, err := sdk.AccAddressFromBech32(pExecResp.Address)
	require.NoError(chain.t, err)

	return a
}

// ContractInfo is a helper function to returns the ContractInfo for the given contract address
func (chain *TestChain) ContractInfo(contractAddr sdk.AccAddress) *types.ContractInfo {
	type testSupporter interface {
		TestSupport() *feeabs.TestSupport
	}
	return chain.App.(testSupporter).TestSupport().WasmKeeper().GetContractInfo(chain.GetContext(), contractAddr)
}
