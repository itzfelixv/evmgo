package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newDiffCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare EVM data across blocks",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	cmd.AddCommand(newDiffStateCmd(flags, deps))
	return cmd
}

func newDiffStateCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	var (
		fromBlock string
		toBlock   string
		slots     []string
		showAll   bool
	)

	cmd := &cobra.Command{
		Use:   "state <address>",
		Short: "Compare account state across two blocks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address, err := normalizeStateAddress(args[0])
			if err != nil {
				return err
			}
			if fromBlock == "" {
				return fmt.Errorf("missing --from-block")
			}
			if toBlock == "" {
				return fmt.Errorf("missing --to-block")
			}

			from, err := rpc.ParseBlockRef(fromBlock)
			if err != nil {
				return err
			}
			to, err := rpc.ParseBlockRef(toBlock)
			if err != nil {
				return err
			}

			seenSlots := make(map[string]struct{}, len(slots))
			for _, slot := range slots {
				if err := validateStorageSlot(slot); err != nil {
					return err
				}
				if _, ok := seenSlots[slot]; ok {
					return fmt.Errorf("duplicate slot %q", slot)
				}
				seenSlots[slot] = struct{}{}
			}

			endpoint, err := config.ResolveRPCURL(*flags, deps.lookupEnv)
			if err != nil {
				return err
			}

			result, err := actions.DiffState(cmd.Context(), rpc.NewClient(endpoint), actions.StateDiffQuery{
				Address:   address,
				FromBlock: from,
				ToBlock:   to,
				Slots:     slots,
			})
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, result)
			}

			return output.WriteText(deps.stdout, renderStateDiffText(result, showAll)...)
		},
	}

	cmd.Flags().StringVar(&fromBlock, "from-block", "", "Start block number or tag")
	cmd.Flags().StringVar(&toBlock, "to-block", "", "End block number or tag")
	cmd.Flags().StringArrayVar(&slots, "slot", nil, "Storage slot to compare")
	cmd.Flags().BoolVar(&showAll, "all", false, "Show unchanged fields")

	return cmd
}
