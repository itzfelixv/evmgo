package cli

import (
	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newStorageCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	var block string

	cmd := &cobra.Command{
		Use:   "storage <address> <slot>",
		Short: "Read a storage slot for a contract",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			address, err := normalizeStateAddress(args[0])
			if err != nil {
				return err
			}
			if err := validateStorageSlot(args[1]); err != nil {
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

			result, err := actions.GetStorage(cmd.Context(), rpc.NewClient(endpoint), address, args[1], ref)
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, result)
			}

			return output.WriteText(
				deps.stdout,
				"address: "+result.Address,
				"slot: "+result.Slot,
				"block: "+result.Block,
				"value: "+result.Value,
			)
		},
	}

	cmd.Flags().StringVar(&block, "block", "", "Block number or tag")
	return cmd
}
