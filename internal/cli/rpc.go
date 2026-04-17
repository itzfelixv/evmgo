package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/actions"
	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
	"github.com/itzfelixv/evmgo/internal/rpc"
)

func newRPCCmd(flags *config.GlobalFlags, deps commandDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "rpc <method> [params...]",
		Short: "Send a raw JSON-RPC request",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, err := config.ResolveRPCURL(*flags, deps.lookupEnv)
			if err != nil {
				return err
			}

			client := rpc.NewClient(endpoint)
			result, err := actions.CallRaw(cmd.Context(), client, args[0], args[1:])
			if err != nil {
				return err
			}

			if flags.JSON {
				return output.WriteJSON(deps.stdout, map[string]any{"result": result})
			}

			pretty, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			return output.WriteText(deps.stdout, string(pretty))
		},
	}
}
