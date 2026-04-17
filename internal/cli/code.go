package cli

import (
	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newCodeCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	var block string

	cmd := &cobra.Command{
		Use:   "code <address>",
		Short: "Read bytecode for a contract address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address, err := normalizeStateAddress(args[0])
			if err != nil {
				return err
			}

			endpoint, err := config.ResolveRPCURL(*flags, deps.lookupEnv)
			if err != nil {
				return err
			}

			ref, err := rpc.ParseBlockRefOrLatest(block)
			if err != nil {
				return err
			}

			result, err := actions.GetCode(cmd.Context(), rpc.NewClient(endpoint), address, ref)
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, result)
			}

			return output.WriteText(deps.stdout, "address: "+result.Address, "block: "+result.Block, "code: "+result.Code)
		},
	}

	cmd.Flags().StringVar(&block, "block", "", "Block number or tag")
	return cmd
}
