package feeabs

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types"
	authTx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

func getFullNode(c *cosmos.CosmosChain) *cosmos.ChainNode {
	if len(c.FullNodes) > 0 {
		return c.FullNodes[0]
	}
	return c.Validators[0]
}

func getTransaction(ctx client.Context, txHash string) (*types.TxResponse, error) {
	var txResp *types.TxResponse
	err := retry.Do(func() error {
		var err error
		txResp, err = authTx.QueryTx(ctx, txHash)
		return err
	},
		retry.Attempts(15),
		retry.Delay(200*time.Millisecond),
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
	)
	return txResp, err
}

func CrossChainSwap(c *cosmos.CosmosChain, ctx context.Context, keyName string, ibcDenom string) (tx ibc.Tx, _ error) {
	tn := c.GetNode()

	txHash, err := tn.ExecTx(ctx, keyName,
		"feeabs", "swap", ibcDenom,
		"--gas", "auto",
	)
	if err != nil {
		return tx, fmt.Errorf("executing transaction failed: %w", err)
	}

	if err := testutil.WaitForBlocks(ctx, 5, tn); err != nil {
		return tx, err
	}

	txResp, err := getTransaction(tn.CliContext(), txHash)
	if err != nil {
		return tx, fmt.Errorf("failed to get transaction %s: %w", txHash, err)
	}
	tx.Height = txResp.Height
	tx.TxHash = txHash

	tx.GasSpent = txResp.GasWanted

	const evType = "send_packet"

	var (
		seq, _           = AttributeValue(txResp, evType, "packet_sequence")
		srcPort, _       = AttributeValue(txResp, evType, "packet_src_port")
		srcChan, _       = AttributeValue(txResp, evType, "packet_src_channel")
		dstPort, _       = AttributeValue(txResp, evType, "packet_dst_port")
		dstChan, _       = AttributeValue(txResp, evType, "packet_dst_channel")
		timeoutHeight, _ = AttributeValue(txResp, evType, "packet_timeout_height")
		timeoutTs, _     = AttributeValue(txResp, evType, "packet_timeout_timestamp")
		data, _          = AttributeValue(txResp, evType, "packet_data")
	)

	tx.Packet.SourcePort = srcPort
	tx.Packet.SourceChannel = srcChan
	tx.Packet.DestPort = dstPort
	tx.Packet.DestChannel = dstChan
	tx.Packet.TimeoutHeight = timeoutHeight
	tx.Packet.Data = []byte(data)

	seqNum, err := strconv.Atoi(seq)
	if err != nil {
		return tx, fmt.Errorf("invalid packet sequence from txResp %s: %w", seq, err)
	}
	tx.Packet.Sequence = uint64(seqNum)

	timeoutNano, err := strconv.ParseUint(timeoutTs, 10, 64)
	if err != nil {
		return tx, fmt.Errorf("invalid packet timestamp timeout %s: %w", timeoutTs, err)
	}
	tx.Packet.TimeoutTimestamp = ibc.Nanoseconds(timeoutNano)

	return tx, err
}

func AddHostZoneProposal(c *cosmos.CosmosChain, ctx context.Context, keyName string, fileLocation string) (string, error) {
	tn := c.GetNode()
	dat, err := os.ReadFile(fileLocation)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	fileName := "add-hostzone.json"

	err = tn.WriteFile(ctx, dat, fileName)
	if err != nil {
		return "", fmt.Errorf("writing add host zone proposal: %w", err)
	}

	filePath := filepath.Join(tn.HomeDir(), fileName)

	command := []string{
		"gov", "submit-legacy-proposal",
		"add-hostzone-config", filePath,
		"--gas", "auto", "--gas-adjustment", "1.5",
	}
	return tn.ExecTx(ctx, keyName, command...)
}

func DeleteHostZoneProposal(c *cosmos.CosmosChain, ctx context.Context, keyName string, fileLocation string) (string, error) {
	tn := c.GetNode()
	dat, err := os.ReadFile(fileLocation)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	fileName := "delete-hostzone.json"

	err = tn.WriteFile(ctx, dat, fileName)
	if err != nil {
		return "", fmt.Errorf("writing delete host zone proposal: %w", err)
	}

	filePath := filepath.Join(tn.HomeDir(), fileName)

	command := []string{
		"gov", "submit-legacy-proposal",
		"delete-hostzone-config", filePath,
		"--gas", "auto", "--gas-adjustment", "1.5",
	}
	return tn.ExecTx(ctx, keyName, command...)
}

func SetHostZoneProposal(c *cosmos.CosmosChain, ctx context.Context, keyName string, fileLocation string) (string, error) {
	tn := c.GetNode()
	dat, err := os.ReadFile(fileLocation)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	fileName := "set-hostzone.json"

	err = tn.WriteFile(ctx, dat, fileName)
	if err != nil {
		return "", fmt.Errorf("writing set host zone proposal: %w", err)
	}

	filePath := filepath.Join(tn.HomeDir(), fileName)

	command := []string{
		"gov", "submit-legacy-proposal",
		"set-hostzone-config", filePath,
		"--gas", "auto", "--gas-adjustment", "1.5",
	}
	return tn.ExecTx(ctx, keyName, command...)
}

func ParamChangeProposal(c *cosmos.CosmosChain, ctx context.Context, keyName string, prop *paramsutils.ParamChangeProposalJSON) (tx cosmos.TxProposal, _ error) {
	tn := c.GetNode()
	content, err := json.Marshal(prop)
	if err != nil {
		return tx, err
	}

	hash := sha256.Sum256(content)
	proposalFilename := fmt.Sprintf("%x.json", hash)
	err = tn.WriteFile(ctx, content, proposalFilename)
	if err != nil {
		return tx, fmt.Errorf("writing param change proposal: %w", err)
	}

	proposalPath := filepath.Join(tn.HomeDir(), proposalFilename)

	command := []string{
		"gov", "submit-legacy-proposal",
		"param-change",
		proposalPath,
		"--gas", "auto",
		"--gas-adjustment", "1.5", "--type", "param-change",
	}

	txHash, err := tn.ExecTx(ctx, keyName, command...)
	if err != nil {
		return tx, fmt.Errorf("executing transaction failed: %w", err)
	}
	return txProposal(c, txHash)
}

func txProposal(c *cosmos.CosmosChain, txHash string) (tx cosmos.TxProposal, _ error) {
	fn := c.GetNode()

	txResp, err := getTransaction(fn.CliContext(), txHash)
	if err != nil {
		return tx, fmt.Errorf("failed to get transaction %s: %w", txHash, err)
	}
	tx.Height = txResp.Height
	tx.TxHash = txHash
	// In cosmos, user is charged for entire gas requested, not the actual gas used.
	tx.GasSpent = txResp.GasWanted

	tx.DepositAmount, _ = AttributeValue(txResp, "proposal_deposit", "amount")

	evtSubmitProp := "submit_proposal"
	tx.ProposalID, _ = AttributeValue(txResp, evtSubmitProp, "proposal_id")
	tx.ProposalType, _ = AttributeValue(txResp, evtSubmitProp, "proposal_type")

	return tx, nil
}
