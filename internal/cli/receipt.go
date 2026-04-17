package cli

import (
	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newReceiptCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	return &cobra.Command{
		Use:     "receipt <hash>",
		Aliases: []string{"rcpt"},
		Short:   "Fetch a transaction receipt by hash",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateTransactionHash(args[0]); err != nil {
				return err
			}

			endpoint, err := config.ResolveRPCURL(*flags, deps.lookupEnv)
			if err != nil {
				return err
			}

			result, err := actions.GetReceipt(cmd.Context(), rpc.NewClient(endpoint), args[0])
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, result)
			}

			return output.WriteText(
				deps.stdout,
				"tx: "+result.TransactionHash,
				"block: "+result.BlockNumber,
				"status: "+result.Status,
				"gasUsed: "+result.GasUsed,
				"contractAddress: "+result.ContractAddress,
			)
		},
	}
}
