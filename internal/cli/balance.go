package cli

import (
	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newBalanceCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	return &cobra.Command{
		Use:     "balance <address>",
		Aliases: []string{"bal"},
		Short:   "Read the native balance for an address",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address, err := normalizeStateAddress(args[0])
			if err != nil {
				return err
			}

			endpoint, err := config.ResolveRPCURL(*flags, deps.lookupEnv)
			if err != nil {
				return err
			}

			result, err := actions.GetBalance(cmd.Context(), rpc.NewClient(endpoint), address)
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, result)
			}

			return output.WriteText(deps.stdout, "address: "+result.Address, "balance: "+result.Balance)
		},
	}
}
