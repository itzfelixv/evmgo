package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newBlockCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	return &cobra.Command{
		Use:     "block <number|0xnumber|latest|earliest|pending|safe|finalized>",
		Aliases: []string{"blk"},
		Short:   "Fetch a block by number or tag",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, err := config.ResolveRPCURL(*flags, deps.lookupEnv)
			if err != nil {
				return err
			}
			ref, err := rpc.ParseBlockRef(args[0])
			if err != nil {
				return err
			}

			result, err := actions.GetBlock(cmd.Context(), rpc.NewClient(endpoint), ref)
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, result)
			}

			return output.WriteText(
				deps.stdout,
				"number: "+result.Number,
				"hash: "+result.Hash,
				"parent: "+result.ParentHash,
				"timestamp: "+result.Timestamp,
				"transactions: "+fmt.Sprintf("%d", result.TransactionCount),
			)
		},
	}
}
