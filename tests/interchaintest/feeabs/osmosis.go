package feeabs

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

func CreatePool(c *cosmos.CosmosChain, ctx context.Context, keyName string, params cosmos.OsmosisPoolParams) (string, error) {
	tn := getFullNode(c)
	poolbz, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	poolFile := "pool.json"

	err = tn.WriteFile(ctx, poolbz, poolFile)
	if err != nil {
		return "", fmt.Errorf("writing add host zone proposal: %w", err)
	}

	if _, err := tn.ExecTx(ctx, keyName,
		"gamm", "create-pool",
		"--pool-file", filepath.Join(tn.HomeDir(), poolFile), "--gas", "auto",
	); err != nil {
		return "", fmt.Errorf("failed to create pool: %w", err)
	}

	stdout, _, err := tn.ExecQuery(ctx, "gamm", "num-pools")
	if err != nil {
		return "", fmt.Errorf("failed to query num pools: %w", err)
	}
	var res map[string]string
	if err := json.Unmarshal(stdout, &res); err != nil {
		return "", fmt.Errorf("failed to unmarshal query response: %w", err)
	}

	numPools, ok := res["num_pools"]
	if !ok {
		return "", fmt.Errorf("could not find number of pools in query response: %w", err)
	}
	return numPools, nil
}

func SetupProposePFM(c *cosmos.CosmosChain, ctx context.Context, keyName string, contractAddress string, message string, ibcdenom string) (txHash string, err error) {
	oneCoin := strconv.FormatInt(1, 10)
	amount := oneCoin + ibcdenom
	tn := getFullNode(c)
	return tn.ExecTx(ctx, keyName,
		"wasm", "execute", contractAddress, message, "--amount", amount, "--gas", "1000000",
	)
}
