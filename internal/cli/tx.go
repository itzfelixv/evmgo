package cli

import (
	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newTxCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	var (
		abiPath   string
		showInput bool
	)

	cmd := &cobra.Command{
		Use:   "tx <hash>",
		Short: "Fetch a transaction by hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateTransactionHash(args[0]); err != nil {
				return err
			}

			endpoint, err := config.ResolveRPCURL(*flags, deps.lookupEnv)
			if err != nil {
				return err
			}

			result, err := actions.GetTransactionView(cmd.Context(), rpc.NewClient(endpoint), args[0], abiPath, showInput)
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, result)
			}

			return output.WriteText(deps.stdout, renderTxText(result)...)
		},
	}

	cmd.Flags().StringVar(&abiPath, "abi", "", "Path to ABI JSON file")
	cmd.Flags().BoolVar(&showInput, "input", false, "Show raw transaction input bytes")
	return cmd
}
