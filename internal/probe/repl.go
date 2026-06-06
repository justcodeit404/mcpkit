// Package probe implements an interactive REPL for exploring an MCP server.
package probe

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/justcodeit404/mcpkit/internal/mcp"
	"github.com/justcodeit404/mcpkit/internal/output"
)

// Options configures the REPL.
type Options struct {
	Client      *mcp.Client
	NoColor     bool
	ShowRaw     bool
	HistoryFile string
}

// REPL is the interactive shell.
type REPL struct {
	opts        Options
	reader      *bufio.Reader
	writer      io.Writer
	snapshot    *mcp.Snapshot
	toolNames   []string
	rsrcURIs    []string
	promptNames []string
	history     []string
}

// NewREPL constructs a REPL with the given options.
func NewREPL(opts Options) *REPL {
	return &REPL{
		opts:   opts,
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// Run starts the REPL loop.
func (r *REPL) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	r.reader = bufio.NewReader(in)
	r.writer = out

	// Fetch snapshot for tab completion.
	snap, err := r.opts.Client.Snapshot(ctx)
	if err != nil {
		// Non-fatal: REPL still works, just no completion.
		fmt.Fprintf(out, "warning: snapshot failed: %v\n", err)
	} else {
		r.snapshot = snap
		for _, t := range snap.Tools {
			r.toolNames = append(r.toolNames, t.Name)
		}
		for _, r2 := range snap.Resources {
			r.rsrcURIs = append(r.rsrcURIs, r2.URI)
		}
		for _, p := range snap.Prompts {
			r.promptNames = append(r.promptNames, p.Name)
		}
	}

	r.printBanner()
	r.loadHistory()

	for {
		r.printPrompt()
		line, err := r.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Fprintln(out, "")
				return r.persistHistory()
			}
			return err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		r.history = append(r.history, line)

		if err := r.dispatch(ctx, line); err != nil {
			if err == io.EOF {
				return r.persistHistory()
			}
			fmt.Fprintf(out, "error: %v\n", err)
		}
	}
}

func (r *REPL) printBanner() {
	bannerStyle := lipgloss.NewStyle().Foreground(output.Primary).Bold(true)
	subStyle := lipgloss.NewStyle().Foreground(output.Muted)
	fmt.Fprintln(r.writer, bannerStyle.Render("🛠  mcpkit probe — interactive MCP REPL"))
	if r.snapshot != nil {
		fmt.Fprintf(r.writer, "  %s  %s@%s\n",
			subStyle.Render("connected to"),
			r.snapshot.ServerInfo.Name,
			r.snapshot.ServerInfo.Version,
		)
		fmt.Fprintf(r.writer, "  %s  %s\n",
			subStyle.Render("protocol:"),
			r.snapshot.ProtocolVersion,
		)
		fmt.Fprintf(r.writer, "  %s  %d tools, %d resources, %d prompts\n",
			subStyle.Render("capabilities:"),
			len(r.snapshot.Tools),
			len(r.snapshot.Resources),
			len(r.snapshot.Prompts),
		)
	}
	fmt.Fprintln(r.writer, subStyle.Render("  type 'help' for commands, 'exit' to quit"))
	fmt.Fprintln(r.writer)
}

func (r *REPL) printPrompt() {
	p := lipgloss.NewStyle().Foreground(output.Primary).Bold(true).Render("mcp> ")
	fmt.Fprint(r.writer, p)
}

func (r *REPL) dispatch(ctx context.Context, line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}
	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "exit", "quit":
		return io.EOF
	case "help", "?":
		r.printHelp()
	case "clear":
		fmt.Fprint(r.writer, "\033[H\033[2J")
	case "list":
		return r.cmdList(args)
	case "tool":
		return r.cmdTool(ctx, args)
	case "read":
		return r.cmdRead(ctx, args)
	case "prompt":
		return r.cmdPrompt(ctx, args)
	case "info":
		return r.cmdInfo()
	case "ping":
		return r.cmdPing(ctx)
	case "raw":
		return r.cmdRaw(ctx, line)
	case "history":
		r.cmdHistory(args)
	case "stats":
		r.cmdStats()
	case "export":
		return r.cmdExport(args)
	default:
		fmt.Fprintf(r.writer, "unknown command: %s (try 'help')\n", cmd)
	}
	return nil
}

// prettyPrint writes v as indented JSON to the REPL output.
func (r *REPL) prettyPrint(v any) error {
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func (r *REPL) printHelp() {
	help := []struct{ cmd, desc string }{
		{"list <kind>", "List tools | resources | prompts | all"},
		{"tool <name> [json]", "Call a tool with optional JSON arguments"},
		{"read <uri>", "Read a resource by URI"},
		{"prompt <name> [json]", "Get a prompt with optional JSON arguments"},
		{"info", "Show server info and capabilities"},
		{"ping", "Health check"},
		{"raw <jsonrpc>", "Send a raw JSON-RPC message"},
		{"history [N]", "Show last N raw JSON-RPC messages (default 20)"},
		{"stats", "Show session statistics"},
		{"export <file>", "Export session transcript to a file"},
		{"clear", "Clear the screen"},
		{"help", "Show this help"},
		{"exit", "Exit the REPL"},
	}
	for _, h := range help {
		fmt.Fprintf(r.writer, "  %s  %s\n",
			lipgloss.NewStyle().Foreground(output.Accent).Bold(true).Render(fmt.Sprintf("%-18s", h.cmd)),
			h.desc,
		)
	}
}

func (r *REPL) cmdList(args []string) error {
	kind := "all"
	if len(args) > 0 {
		kind = strings.ToLower(args[0])
	}
	if r.snapshot == nil {
		return fmt.Errorf("no snapshot available")
	}
	switch kind {
	case "all":
		r.printTools()
		fmt.Fprintln(r.writer)
		r.printResources()
		fmt.Fprintln(r.writer)
		r.printPrompts()
	case "tools", "tool":
		r.printTools()
	case "resources", "resource":
		r.printResources()
	case "prompts", "prompt":
		r.printPrompts()
	default:
		return fmt.Errorf("unknown list kind: %s (tools|resources|prompts|all)", kind)
	}
	return nil
}

func (r *REPL) printTools() {
	fmt.Fprintln(r.writer, lipgloss.NewStyle().Bold(true).Render("Tools:"))
	if len(r.snapshot.Tools) == 0 {
		fmt.Fprintln(r.writer, "  (none)")
		return
	}
	for _, t := range r.snapshot.Tools {
		desc := t.Description
		if desc == "" {
			desc = "(no description)"
		}
		fmt.Fprintf(r.writer, "  %s  %s\n",
			lipgloss.NewStyle().Foreground(output.Accent).Render(t.Name),
			lipgloss.NewStyle().Foreground(output.Muted).Render(desc),
		)
	}
}

func (r *REPL) printResources() {
	fmt.Fprintln(r.writer, lipgloss.NewStyle().Bold(true).Render("Resources:"))
	if len(r.snapshot.Resources) == 0 {
		fmt.Fprintln(r.writer, "  (none)")
		return
	}
	for _, r2 := range r.snapshot.Resources {
		fmt.Fprintf(r.writer, "  %s  %s\n",
			lipgloss.NewStyle().Foreground(output.Accent).Render(r2.URI),
			lipgloss.NewStyle().Foreground(output.Muted).Render(r2.Description),
		)
	}
}

func (r *REPL) printPrompts() {
	fmt.Fprintln(r.writer, lipgloss.NewStyle().Bold(true).Render("Prompts:"))
	if len(r.snapshot.Prompts) == 0 {
		fmt.Fprintln(r.writer, "  (none)")
		return
	}
	for _, p := range r.snapshot.Prompts {
		desc := p.Description
		if desc == "" {
			desc = "(no description)"
		}
		fmt.Fprintf(r.writer, "  %s  %s\n",
			lipgloss.NewStyle().Foreground(output.Accent).Render(p.Name),
			lipgloss.NewStyle().Foreground(output.Muted).Render(desc),
		)
	}
}

func (r *REPL) cmdTool(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tool <name> [json_args]")
	}
	name := args[0]
	var callArgs map[string]any
	if len(args) > 1 {
		raw := strings.Join(args[1:], " ")
		if err := json.Unmarshal([]byte(raw), &callArgs); err != nil {
			return fmt.Errorf("invalid JSON args: %w", err)
		}
	}
	res, err := r.opts.Client.CallTool(ctx, name, callArgs)
	if err != nil {
		return err
	}
	return r.prettyPrint(res)
}

func (r *REPL) cmdRead(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: read <uri>")
	}
	res, err := r.opts.Client.ReadResource(ctx, args[0])
	if err != nil {
		return err
	}
	return r.prettyPrint(res)
}

func (r *REPL) cmdPrompt(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: prompt <name> [json_args]")
	}
	name := args[0]
	var promptArgs map[string]string
	if len(args) > 1 {
		raw := strings.Join(args[1:], " ")
		if err := json.Unmarshal([]byte(raw), &promptArgs); err != nil {
			return fmt.Errorf("invalid JSON args: %w", err)
		}
	}
	res, err := r.opts.Client.GetPrompt(ctx, name, promptArgs)
	if err != nil {
		return err
	}
	return r.prettyPrint(res)
}

func (r *REPL) cmdInfo() error {
	if r.snapshot == nil {
		return fmt.Errorf("no snapshot available")
	}
	return r.prettyPrint(r.snapshot.ServerInfo)
}

func (r *REPL) cmdPing(ctx context.Context) error {
	if err := r.opts.Client.Ping(ctx); err != nil {
		return err
	}
	fmt.Fprintln(r.writer, lipgloss.NewStyle().Foreground(output.Secondary).Render("✓ pong"))
	return nil
}

func (r *REPL) cmdRaw(ctx context.Context, line string) error {
	idx := strings.Index(line, "raw")
	if idx == -1 {
		return fmt.Errorf("usage: raw <jsonrpc>")
	}
	raw := strings.TrimSpace(line[idx+3:])
	var msg json.RawMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	// Send raw via the client (delegates to session).
	_, err := r.opts.Client.SendRaw(ctx, msg)
	return err
}

func (r *REPL) cmdHistory(args []string) {
	limit := 20
	if len(args) > 0 {
		var n int
		_, _ = fmt.Sscanf(args[0], "%d", &n)
		if n > 0 {
			limit = n
		}
	}
	start := 0
	if len(r.opts.Client.Messages()) > limit {
		start = len(r.opts.Client.Messages()) - limit
	}
	for i, m := range r.opts.Client.Messages()[start:] {
		marker := "→"
		if m.Direction == "recv" {
			marker = "←"
		}
		fmt.Fprintf(r.writer, "  %s [%d] %s\n", marker, i+start, string(m.Payload))
	}
}

func (r *REPL) cmdStats() {
	msgs := r.opts.Client.Messages()
	sent, recv := 0, 0
	for _, m := range msgs {
		if m.Direction == "send" {
			sent++
		} else {
			recv++
		}
	}
	fmt.Fprintf(r.writer, "  Messages sent:  %d\n", sent)
	fmt.Fprintf(r.writer, "  Messages recv:  %d\n", recv)
	fmt.Fprintf(r.writer, "  Commands run:   %d\n", len(r.history))
}

func (r *REPL) cmdExport(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: export <file>")
	}
	f, err := os.Create(args[0])
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(r.opts.Client.Messages()); err != nil {
		return err
	}
	fmt.Fprintf(r.writer, "  exported %d messages to %s\n",
		len(r.opts.Client.Messages()), args[0])
	return nil
}

func (r *REPL) historyFile() string {
	if r.opts.HistoryFile != "" {
		return r.opts.HistoryFile
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".mcpkit_history")
}

func (r *REPL) loadHistory() {
	path := r.historyFile()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line != "" {
			r.history = append(r.history, line)
		}
	}
}

func (r *REPL) persistHistory() error {
	path := r.historyFile()
	if path == "" {
		return nil
	}
	return os.WriteFile(path, []byte(strings.Join(r.history, "\n")), 0o600)
}
