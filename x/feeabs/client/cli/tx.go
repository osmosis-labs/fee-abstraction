package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
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
	txCmd.AddCommand(NewFundFeeAbsModuleAccount())

	return txCmd
}

func NewQueryOsmosisTWAPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "query-osmosis-twap",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSendQueryIbcDenomTWAP(clientCtx.GetFromAddress())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewSwapOverChainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "swap [ibc-denom]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgSwapCrossChain(clientCtx.GetFromAddress(), args[0])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewFundFeeAbsModuleAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "fund [amount]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgFundFeeAbsModuleAccount(clientCtx.GetFromAddress(), coins)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCmdSubmitAddHostZoneProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-hostzone-config [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit an add host zone proposal",
		Long: "Submit an add host zone proposal along with an initial deposit.\n" +
			"Please specify a IBC denom identifier you want to use as abstraction fee..\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			proposal, err := ParseAddHostZoneProposalJSON(clientCtx.LegacyAmino, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := types.NewAddHostZoneProposal(
				proposal.Title, proposal.Description, proposal.HostChainFeeAbsConfig,
			)

			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

func NewCmdSubmitDeleteHostZoneProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-hostzone-config [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit an delete host zone proposal",
		Long: "Submit an delete host zone proposal\n" +
			"Please specify a IBC denom identifier you want to use as abstraction fee..\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			proposal, err := ParseDeleteHostZoneProposalJSON(clientCtx.LegacyAmino, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := types.NewDeleteHostZoneProposal(
				proposal.Title, proposal.Description, proposal.IbcDenom,
			)

			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

func NewCmdSubmitSetHostZoneProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-hostzone-config [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit an change host zone proposal",
		Long: "Submit an change host zone proposal\n" +
			"Please specify a IBC denom identifier you want to use as abstraction fee..\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			proposal, err := ParseSetHostZoneProposalJSON(clientCtx.LegacyAmino, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := types.NewSetHostZoneProposal(
				proposal.Title, proposal.Description, proposal.HostChainFeeAbsConfig,
			)

			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
