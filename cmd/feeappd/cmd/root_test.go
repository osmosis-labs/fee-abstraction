package cmd_test

import (
	"testing"

	app "github.com/osmosis-labs/fee-abstraction/v4/app"
	"github.com/osmosis-labs/fee-abstraction/v4/cmd/feeappd/cmd"
	"github.com/stretchr/testify/require"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func TestRootCmdConfig(t *testing.T) {
	rootCmd, _ := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{
		"config",          // Test the config cmd
		"keyring-backend", // key
		"test",            // value
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", app.DefaultNodeHome))
}
