package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newCallCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	var (
		to      string
		abiPath string
		method  string
		args    []string
		block   string
	)

	cmd := &cobra.Command{
		Use:   "call",
		Short: "Execute an ABI-backed eth_call",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rawTo := to
			to, err := normalizeStateAddress(rawTo)
			if err != nil {
				return fmt.Errorf("invalid --to address %q", rawTo)
			}

			endpoint, err := config.ResolveRPCURL(*flags, deps.lookupEnv)
			if err != nil {
				return err
			}

			ref, err := rpc.ParseBlockRefOrLatest(block)
			if err != nil {
				return err
			}

			result, err := actions.CallContract(
				cmd.Context(),
				rpc.NewClient(endpoint),
				to,
				abiPath,
				method,
				args,
				ref,
			)
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, result)
			}

			lines := []string{
				"to: " + result.To,
				"method: " + result.Method,
				"data: " + result.Data,
				"block: " + result.Block,
				"raw: " + result.Raw,
			}
			for _, value := range result.Decoded {
				lines = append(lines, "decoded: "+value)
			}
			return output.WriteText(deps.stdout, lines...)
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Target contract address")
	cmd.Flags().StringVar(&abiPath, "abi", "", "Path to ABI JSON file")
	cmd.Flags().StringVar(&method, "method", "", "ABI method name")
	cmd.Flags().StringSliceVar(&args, "args", nil, "Method arguments")
	cmd.Flags().StringVar(&block, "block", "", "Block number or tag")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("abi")
	_ = cmd.MarkFlagRequired("method")

	return cmd
}
