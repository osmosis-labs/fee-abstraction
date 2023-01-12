package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
	"github.com/spf13/cobra"
)

// NewTxCmd returns a root CLI command handler for all x/exp transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Exp transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(NewQueryOsmosisSpotPriceCmd())
	return txCmd
}
func NewQueryOsmosisSpotPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "queryomosis",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgSendQuerySpotPrice(clientCtx.GetFromAddress())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)

		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
