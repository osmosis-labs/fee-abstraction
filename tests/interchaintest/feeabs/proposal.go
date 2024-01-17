package feeabs

import (
	"context"
	"fmt"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"os"
	"path/filepath"
)

func DeleteHostZoneProposal(c *cosmos.CosmosChain, ctx context.Context, keyName string, fileLocation string) (string, error) {
	tn := c.Validators[0]
	if len(c.FullNodes) > 0 {
		tn = c.FullNodes[0]
	}
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
	}
	return tn.ExecTx(ctx, keyName, command...)
}

func SetHostZoneProposal(c *cosmos.CosmosChain, ctx context.Context, keyName string, fileLocation string) (string, error) {
	tn := c.Validators[0]
	if len(c.FullNodes) > 0 {
		tn = c.FullNodes[0]
	}
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
	}
	return tn.ExecTx(ctx, keyName, command...)
}
