package cli

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/itzfelixv/evmgo/internal/config"
	"github.com/itzfelixv/evmgo/internal/output"
)

type commandDeps struct {
	stdout    io.Writer
	stderr    io.Writer
	lookupEnv func(string) string
}

func Execute(stdout, stderr io.Writer) error {
	return executeWithEnv(stdout, stderr, os.Getenv, os.Args[1:])
}

func executeWithEnv(stdout, stderr io.Writer, lookupEnv func(string) string, args []string) error {
	cmd := NewRootCmd(stdout, stderr, lookupEnv)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err != nil {
		jsonOut, _ := cmd.PersistentFlags().GetBool("json")
		_ = output.WriteError(stderr, jsonOut, err)
	}
	return err
}

func NewRootCmd(stdout, stderr io.Writer, lookupEnv func(string) string) *cobra.Command {
	flags := &config.GlobalFlags{}
	deps := commandDeps{
		stdout:    stdout,
		stderr:    stderr,
		lookupEnv: lookupEnv,
	}

	root := &cobra.Command{
		Use:          "evmgo",
		Short:        "CLI for read-only EVM interactions",
		Args:         cobra.NoArgs,
		RunE:         func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
		SilenceUsage: true,
	}

	root.SetOut(stdout)
	root.SetErr(stderr)
	root.PersistentFlags().StringVar(&flags.RPC, "rpc", "", "JSON-RPC endpoint URL")
	root.PersistentFlags().BoolVar(&flags.JSON, "json", false, "Emit machine-readable JSON")

	registerCommands(root, flags, deps)
	return root
}

func registerCommands(root *cobra.Command, flags *config.GlobalFlags, deps commandDeps) {
	root.AddCommand(
		newRPCCmd(flags, deps),
		newBlockCmd(flags, deps),
		newTxCmd(flags, deps),
		newReceiptCmd(flags, deps),
		newBalanceCmd(flags, deps),
		newCodeCmd(flags, deps),
		newStorageCmd(flags, deps),
		newCallCmd(flags, deps),
		newLogsCmd(flags, deps),
	)
}
