package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryOsmosisSpotPrice(),
		GetCmdQueryFeeabsModuleBalances(),
	)

	return cmd
}

func GetCmdQueryOsmosisSpotPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "osmo-spot-price",
		Args:  cobra.NoArgs,
		Short: "Query the spot price of osmosis",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.OsmosisSpotPrice(cmd.Context(), &types.QueryOsmosisSpotPriceRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetCmdQueryFeeabsModuleBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-balances",
		Args:  cobra.NoArgs,
		Short: "Query feeabs module balances",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.FeeabsModuleBalances(cmd.Context(), &types.QueryFeeabsModuleBalacesRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
