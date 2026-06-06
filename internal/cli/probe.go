package cli

import (
	"os"

	"github.com/justcodeit404/mcpkit/internal/probe"
	"github.com/spf13/cobra"
)

var probeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Interactive REPL for exploring an MCP server",
	Long: `Probe opens an interactive shell for browsing and testing an MCP server.

Inside the REPL, try:
  list tools                  # show all available tools
  tool <name> [json_args]     # call a tool
  list resources              # show resources
  read <uri>                  # read a resource
  list prompts                # show prompts
  prompt <name> [json_args]   # get a prompt
  info                        # show server info
  ping                        # health check
  raw {jsonrpc message}       # send raw JSON-RPC
  history [N]                 # show last N JSON-RPC messages (default 20)
  stats                       # show session statistics
  export <file>               # export session transcript to file
  clear                       # clear the screen
  help                        # show all commands
  exit                        # quit the REPL`,
	RunE: runProbe,
}

func init() {
	probeCmd.Flags().Bool("raw", false, "Show raw JSON-RPC messages")
	probeCmd.Flags().String("history-file", "", "Path to save command history (default ~/.mcpkit_history)")
}

func runProbe(cmd *cobra.Command, _ []string) error {
	flags := bindSharedFlags(cmd)
	raw := getBool(cmd.Flags(), "raw")
	historyFile := getString(cmd.Flags(), "history-file")

	client, ctx, cancel, err := connectClient(flags)
	if err != nil {
		return err
	}
	defer cancel()
	defer client.Disconnect()

	repl := probe.NewREPL(probe.Options{
		Client:      client,
		NoColor:     flags.NoColor,
		ShowRaw:     raw,
		HistoryFile: historyFile,
	})
	return repl.Run(ctx, os.Stdin, os.Stdout)
}
