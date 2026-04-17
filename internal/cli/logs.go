package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newLogsCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	var (
		address   string
		abiPath   string
		eventName string
		topics    []string
		fromBlock string
		toBlock   string
	)

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Query logs for an address and block range",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rawAddress := address
			address, err := normalizeStateAddress(rawAddress)
			if err != nil {
				return fmt.Errorf("invalid --address value %q", rawAddress)
			}

			endpoint, err := config.ResolveRPCURL(*flags, deps.lookupEnv)
			if err != nil {
				return err
			}
			from, err := rpc.ParseBlockRefOrLatest(fromBlock)
			if err != nil {
				return err
			}
			to, err := rpc.ParseBlockRefOrLatest(toBlock)
			if err != nil {
				return err
			}

			result, err := actions.GetLogs(cmd.Context(), rpc.NewClient(endpoint), actions.LogQuery{
				Address: address,
				ABIPath: abiPath,
				Event:   eventName,
				Topics:  topics,
				From:    from,
				To:      to,
			})
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, result)
			}

			lines := make([]string, 0, len(result))
			for _, entry := range result {
				lines = append(lines, fmt.Sprintf("block=%s tx=%s address=%s topics=%d", entry.BlockNumber, entry.TransactionHash, entry.Address, len(entry.Topics)))
			}
			if len(lines) == 0 {
				lines = append(lines, "no logs found")
			}
			return output.WriteText(deps.stdout, lines...)
		},
	}

	cmd.Flags().StringVar(&address, "address", "", "Contract address to filter")
	cmd.Flags().StringVar(&abiPath, "abi", "", "Path to ABI JSON file")
	cmd.Flags().StringVar(&eventName, "event", "", "Event name from the ABI")
	cmd.Flags().StringSliceVar(&topics, "topic", nil, "Additional topic filters")
	cmd.Flags().StringVar(&fromBlock, "from-block", "", "Start block number or tag")
	cmd.Flags().StringVar(&toBlock, "to-block", "", "End block number or tag")
	_ = cmd.MarkFlagRequired("address")

	return cmd
}
