// Package mcp wraps the modelcontextprotocol/go-sdk client with a clean,
// ergonomic API. It also captures raw JSON-RPC traffic for replay and audit.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ClientVersion is the version reported to MCP servers. Set via ldflags at build time.
var ClientVersion = "dev"

// Client is a thin wrapper around the go-sdk client. It tracks a single
// session and the raw JSON-RPC traffic that has flowed through it.
type Client struct {
	cfg       Config
	session   *mcpsdk.ClientSession
	messages  []Message
	mu        sync.Mutex
	connected bool
}

// Config defines how to connect to an MCP server.
type Config struct {
	// Transport: "stdio" (default), "sse", "streamable-http".
	Transport string
	// Command + Args: for stdio transport (e.g. "npx -y some-server").
	Command string
	Args    []string
	// URL: for HTTP transports.
	URL string
	// Headers: optional HTTP headers (for auth, custom routing).
	Headers map[string]string
	// Env: extra environment variables for stdio subprocess.
	Env []string
	// ProtocolVersion: MCP protocol version to negotiate (e.g. "2025-11-25").
	ProtocolVersion string
}

// Message is a captured JSON-RPC message for the raw message log.
type Message struct {
	Direction string          // "send" or "recv"
	Method    string          `json:",omitempty"`
	ID        json.RawMessage `json:",omitempty"`
	Payload   json.RawMessage
}

// Snapshot captures the state of a server after a full discovery handshake.
type Snapshot struct {
	ServerInfo         *mcpsdk.Implementation
	ServerCapabilities *mcpsdk.ServerCapabilities
	ProtocolVersion    string
	Instructions       string
	Tools              []*mcpsdk.Tool
	Resources          []*mcpsdk.Resource
	ResourceTemplates  []*mcpsdk.ResourceTemplate
	Prompts            []*mcpsdk.Prompt
}

// NewClient creates a new client with the given configuration.
func NewClient(cfg Config) *Client {
	if cfg.Transport == "" {
		cfg.Transport = "stdio"
	}
	return &Client{cfg: cfg}
}

// Connect establishes a connection, performs the initialize handshake, and
// sends the notifications/initialized notification.
func (c *Client) Connect(ctx context.Context) error {
	transport, err := c.buildTransport()
	if err != nil {
		return fmt.Errorf("build transport: %w", err)
	}

	client := mcpsdk.NewClient(
		&mcpsdk.Implementation{
			Name:    "mcpkit",
			Version: ClientVersion,
		},
		&mcpsdk.ClientOptions{},
	)

	sess, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	c.session = sess
	c.connected = true
	return nil
}

// Disconnect gracefully closes the session.
func (c *Client) Disconnect() {
	if c.session != nil {
		_ = c.session.Close()
		c.connected = false
	}
}

// Connected reports whether the client has an active session.
func (c *Client) Connected() bool {
	return c.connected
}

// InitializeResult returns the server info from the initialize handshake.
func (c *Client) InitializeResult() *mcpsdk.InitializeResult {
	if c.session == nil {
		return nil
	}
	return c.session.InitializeResult()
}

// ListTools fetches the server's tool catalog.
func (c *Client) ListTools(ctx context.Context) ([]*mcpsdk.Tool, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}
	res, err := c.session.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("tools/list: %w", err)
	}
	return res.Tools, nil
}

// CallTool invokes a tool by name with the given arguments.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (*mcpsdk.CallToolResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}
	res, err := c.session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return nil, fmt.Errorf("tools/call %s: %w", name, err)
	}
	return res, nil
}

// ListResources fetches the server's resources.
func (c *Client) ListResources(ctx context.Context) ([]*mcpsdk.Resource, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}
	res, err := c.session.ListResources(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("resources/list: %w", err)
	}
	return res.Resources, nil
}

// ReadResource reads a resource by URI.
func (c *Client) ReadResource(ctx context.Context, uri string) (*mcpsdk.ReadResourceResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}
	res, err := c.session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: uri})
	if err != nil {
		return nil, fmt.Errorf("resources/read %s: %w", uri, err)
	}
	return res, nil
}

// ListPrompts fetches the server's prompt templates.
func (c *Client) ListPrompts(ctx context.Context) ([]*mcpsdk.Prompt, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}
	res, err := c.session.ListPrompts(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("prompts/list: %w", err)
	}
	return res.Prompts, nil
}

// GetPrompt fetches a prompt by name with the given arguments.
func (c *Client) GetPrompt(ctx context.Context, name string, args map[string]string) (*mcpsdk.GetPromptResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}
	res, err := c.session.GetPrompt(ctx, &mcpsdk.GetPromptParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return nil, fmt.Errorf("prompts/get %s: %w", name, err)
	}
	return res, nil
}

// Ping sends a ping and waits for the response.
func (c *Client) Ping(ctx context.Context) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}
	return c.session.Ping(ctx, nil)
}

// SendRaw is a placeholder — the go-sdk doesn't expose a raw request API.
// We log the message and return an explanatory error.
func (c *Client) SendRaw(_ context.Context, msg json.RawMessage) (json.RawMessage, error) {
	c.mu.Lock()
	c.messages = append(c.messages, Message{
		Direction: "send",
		Payload:   msg,
	})
	c.mu.Unlock()
	return nil, fmt.Errorf("raw JSON-RPC sending not yet supported by mcpkit (use typed methods instead)")
}

// Snapshot returns a complete view of the server's capabilities.
func (c *Client) Snapshot(ctx context.Context) (*Snapshot, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}
	init := c.session.InitializeResult()
	snap := &Snapshot{
		ServerInfo:         init.ServerInfo,
		ServerCapabilities: init.Capabilities,
		ProtocolVersion:    init.ProtocolVersion,
		Instructions:       init.Instructions,
	}

	// Try to fetch each capability — ignore errors so a server with
	// only tools still produces a useful snapshot.
	if tools, err := c.ListTools(ctx); err == nil {
		snap.Tools = tools
	}
	if resources, err := c.ListResources(ctx); err == nil {
		snap.Resources = resources
	}
	if prompts, err := c.ListPrompts(ctx); err == nil {
		snap.Prompts = prompts
	}
	return snap, nil
}

// Messages returns a copy of the captured raw JSON-RPC traffic.
func (c *Client) Messages() []Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Message, len(c.messages))
	copy(out, c.messages)
	return out
}

// buildTransport constructs the appropriate transport for the config.
func (c *Client) buildTransport() (mcpsdk.Transport, error) {
	switch c.cfg.Transport {
	case "stdio", "":
		if c.cfg.Command == "" {
			return nil, fmt.Errorf("stdio transport requires --command")
		}
		cmd := exec.Command(c.cfg.Command, c.cfg.Args...)
		if len(c.cfg.Env) > 0 {
			cmd.Env = append(cmd.Environ(), c.cfg.Env...)
		}
		return &mcpsdk.CommandTransport{Command: cmd}, nil
	case "streamable-http":
		if c.cfg.URL == "" {
			return nil, fmt.Errorf("streamable-http transport requires --url")
		}
		return &mcpsdk.StreamableClientTransport{
			Endpoint: c.cfg.URL,
		}, nil
	case "sse":
		if c.cfg.URL == "" {
			return nil, fmt.Errorf("sse transport requires --url")
		}
		return &mcpsdk.SSEClientTransport{Endpoint: c.cfg.URL}, nil
	default:
		return nil, fmt.Errorf("unknown transport: %s", c.cfg.Transport)
	}
}
