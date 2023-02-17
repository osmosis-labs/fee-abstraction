package cli

import (
	"strconv"
	"time"

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

	txCmd.AddCommand(NewQueryOsmosisTWAPCmd())
	txCmd.AddCommand(NewSwapOverChainCmd())

	return txCmd
}

func NewQueryOsmosisTWAPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "query-osmosis-twap [ibc-denom] [start-time]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			if err != nil {
				return err
			}

			timeUnix, _ := strconv.ParseInt(args[1], 10, 64)
			startTime := time.Unix(timeUnix, 0)

			msg := types.NewMsgSendQueryIbcDenomTWAP(clientCtx.GetFromAddress(), args[0], startTime)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)

		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewSwapOverChainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "swap",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgSwapCrossChain(clientCtx.GetFromAddress())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)

		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
