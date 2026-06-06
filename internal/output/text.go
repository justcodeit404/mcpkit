package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette — shared across all text output.
var (
	Primary    = lipgloss.Color("#7B61FF") // Wrapped purple
	Secondary  = lipgloss.Color("#04B575") // Success green
	Danger     = lipgloss.Color("#FF5C5C") // Critical red
	Warning    = lipgloss.Color("#FFB347") // Warn orange
	Muted      = lipgloss.Color("#6B7280") // Dimmed text
	Accent     = lipgloss.Color("#00BFFF") // Bright cyan
	Background = lipgloss.Color("#1A1B26")
	Surface    = lipgloss.Color("#24283B")
	Foreground = lipgloss.Color("#C0CAF5")
	OnPrimary  = lipgloss.Color("#FFFFFF")
)

// Style definitions.
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(Muted).
			MarginTop(1).
			MarginBottom(1)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1).
			Width(40)

	PassStyle = lipgloss.NewStyle().Foreground(Secondary).Bold(true)
	FailStyle = lipgloss.NewStyle().Foreground(Danger).Bold(true)
	WarnStyle = lipgloss.NewStyle().Foreground(Warning).Bold(true)
	InfoStyle = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	CriticalStyle = lipgloss.NewStyle().Foreground(Danger).Bold(true)
	HighStyle     = lipgloss.NewStyle().Foreground(Warning).Bold(true)
	MediumStyle   = lipgloss.NewStyle().Foreground(Accent).Bold(true)
	LowStyle      = lipgloss.NewStyle().Foreground(Muted).Bold(true)

	KeyStyle   = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	ValueStyle = lipgloss.NewStyle().Foreground(Foreground)

	RuleIDStyle = lipgloss.NewStyle().Bold(true).Foreground(Primary)
	TagStyle    = lipgloss.NewStyle().Foreground(Muted).Italic(true)
	Mut         = lipgloss.NewStyle().Foreground(Muted)
)

// TextFormatter renders results as colorized terminal output.
type TextFormatter struct {
	NoColor bool
	Writer  io.Writer
}

// NewTextFormatter constructs a default TextFormatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{Writer: os.Stdout}
}

// Format dispatches to the right renderer based on the result type.
func (f *TextFormatter) Format(w io.Writer, result any) error {
	if w != nil {
		f.Writer = w
	}
	switch v := result.(type) {
	case TestResultRenderable:
		return f.renderTest(v)
	case ScanResultRenderable:
		return f.renderScan(v)
	case BenchResultRenderable:
		return f.renderBench(v)
	default:
		_, err := fmt.Fprintf(f.Writer, "%+v\n", result)
		return err
	}
}

// Banner prints a styled title block.
func (f *TextFormatter) Banner(title string, subtitle ...string) {
	t := TitleStyle.Render("🛠  " + title)
	fmt.Fprintln(f.Writer, t)
	for _, s := range subtitle {
		fmt.Fprintln(f.Writer, SubtitleStyle.Render("  "+s))
	}
	fmt.Fprintln(f.Writer)
}

// Success prints a green checkmark + message.
func (f *TextFormatter) Success(msg string) {
	fmt.Fprintln(f.Writer, PassStyle.Render("✓")+" "+msg)
}

// Failure prints a red X + message.
func (f *TextFormatter) Failure(msg string) {
	fmt.Fprintln(f.Writer, FailStyle.Render("✗")+" "+msg)
}

// Warning prints an orange ! + message.
func (f *TextFormatter) Warning(msg string) {
	fmt.Fprintln(f.Writer, WarnStyle.Render("!")+" "+msg)
}

// Info prints a cyan arrow + message.
func (f *TextFormatter) Info(msg string) {
	fmt.Fprintln(f.Writer, InfoStyle.Render("→")+" "+msg)
}

// KeyValue prints a styled key:value pair.
func (f *TextFormatter) KeyValue(key, value string) {
	fmt.Fprintf(f.Writer, "  %s %s\n",
		KeyStyle.Render(key+":"),
		ValueStyle.Render(value),
	)
}

// Section header.
func (f *TextFormatter) Section(title string) {
	fmt.Fprintln(f.Writer)
	fmt.Fprintln(f.Writer, HeaderStyle.Render(" "+title))
}

// SeverityIcon returns a colored severity label.
func (f *TextFormatter) SeverityIcon(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return CriticalStyle.Render("◆ CRITICAL")
	case "high":
		return HighStyle.Render("▲ HIGH")
	case "medium":
		return MediumStyle.Render("● MEDIUM")
	case "low":
		return LowStyle.Render("○ LOW")
	case "info":
		return LowStyle.Render("· INFO")
	case "pass":
		return PassStyle.Render("✓ PASS")
	case "fail":
		return FailStyle.Render("✗ FAIL")
	case "skip":
		return Mut.Render("- SKIP")
	default:
		return severity
	}
}

// ---- Renderable result types ----

// TestResultRenderable is a typed test result for the text formatter.
type TestResultRenderable struct {
	Server  string
	Checks  []CheckResult
	Summary TestSummary
	DurMS   int64
}

type CheckResult struct {
	ID       string
	Name     string
	Status   string // pass|fail|skip
	Message  string
	Duration string
}

type TestSummary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
}

// ScanResultRenderable is a typed scan result for the text formatter.
type ScanResultRenderable struct {
	Server   string
	Findings []FindingResult
	Summary  ScanSummary
}

type FindingResult struct {
	RuleID      string
	RuleName    string
	Severity    string
	Target      string // e.g. tool name or resource URI
	Description string
	Remediation string
}

type ScanSummary struct {
	Total    int
	Critical int
	High     int
	Medium   int
	Low      int
	Info     int
}

// BenchResultRenderable is a typed bench result for the text formatter.
type BenchResultRenderable struct {
	Server    string
	Method    string
	Metrics   BenchMetrics
	Histogram []HistogramBucket
}

type BenchMetrics struct {
	Iterations int
	Errors     int
	Min        string
	Max        string
	Mean       string
	Median     string
	P75        string
	P90        string
	P95        string
	P99        string
	Stddev     string
	Throughput string
	TotalDur   string
}

type HistogramBucket struct {
	From  string
	To    string
	Count int
}

func (f *TextFormatter) renderTest(r TestResultRenderable) error {
	f.Banner("mcpkit test", "MCP server protocol compliance")
	f.KeyValue("Server", r.Server)

	// Checks table
	rows := make([][]string, 0, len(r.Checks))
	rows = append(rows, []string{"ID", "Check", "Status", "Time", "Detail"})
	for _, c := range r.Checks {
		rows = append(rows, []string{
			c.ID,
			c.Name,
			f.SeverityIcon(c.Status),
			c.Duration,
			c.Message,
		})
	}
	f.printTable(rows)

	// Summary
	f.Section("Summary")
	fmt.Fprintf(f.Writer, "  %s  passed  %s  failed  %s  skipped  (%d total)\n",
		PassStyle.Render(fmt.Sprintf("%d", r.Summary.Passed)),
		FailStyle.Render(fmt.Sprintf("%d", r.Summary.Failed)),
		Mut.Render(fmt.Sprintf("%d", r.Summary.Skipped)),
		r.Summary.Total,
	)
	fmt.Fprintf(f.Writer, "  %s\n", Mut.Render(fmt.Sprintf("Completed in %dms", r.DurMS)))
	fmt.Fprintln(f.Writer)
	return nil
}

func (f *TextFormatter) renderScan(r ScanResultRenderable) error {
	f.Banner("mcpkit scan", "MCP server security scan")
	f.KeyValue("Server", r.Server)

	if len(r.Findings) == 0 {
		f.Success("No security findings — looks clean.")
		return nil
	}

	// Findings table
	rows := [][]string{{"Rule", "Severity", "Target", "Finding"}}
	for _, fi := range r.Findings {
		rows = append(rows, []string{
			RuleIDStyle.Render(fi.RuleID),
			f.SeverityIcon(fi.Severity),
			ValueStyle.Render(fi.Target),
			fi.Description,
		})
	}
	f.printTable(rows)

	// Summary
	f.Section("Summary")
	fmt.Fprintf(f.Writer, "  %s\n", Mut.Render(fmt.Sprintf(
		"%d findings: %d critical, %d high, %d medium, %d low, %d info",
		r.Summary.Total, r.Summary.Critical, r.Summary.High,
		r.Summary.Medium, r.Summary.Low, r.Summary.Info,
	)))
	fmt.Fprintln(f.Writer)
	return nil
}

func (f *TextFormatter) renderBench(r BenchResultRenderable) error {
	f.Banner("mcpkit bench", "MCP server performance benchmark")
	f.KeyValue("Server", r.Server)
	f.KeyValue("Method", r.Method)

	// Metrics
	f.Section("Latency")
	m := r.Metrics
	f.KeyValue("min", m.Min)
	f.KeyValue("max", m.Max)
	f.KeyValue("mean", m.Mean)
	f.KeyValue("p50", m.Median)
	f.KeyValue("p90", m.P90)
	f.KeyValue("p95", m.P95)
	f.KeyValue("p99", m.P99)
	f.KeyValue("stddev", m.Stddev)

	f.Section("Throughput")
	f.KeyValue("iterations", fmt.Sprintf("%d", m.Iterations))
	f.KeyValue("errors", fmt.Sprintf("%d", m.Errors))
	f.KeyValue("throughput", m.Throughput)
	f.KeyValue("total duration", m.TotalDur)

	// Histogram
	if len(r.Histogram) > 0 {
		f.Section("Latency Distribution")
		maxCount := 0
		for _, b := range r.Histogram {
			if b.Count > maxCount {
				maxCount = b.Count
			}
		}
		for _, b := range r.Histogram {
			barLen := 0
			if maxCount > 0 {
				barLen = int(float64(b.Count) / float64(maxCount) * 30)
			}
			bar := strings.Repeat("█", barLen)
			fmt.Fprintf(f.Writer, "  %s-%s  %s %d\n",
				ValueStyle.Render(pad(b.From, 8)),
				ValueStyle.Render(pad(b.To, 8)),
				AccentStyle.Render(bar),
				b.Count,
			)
		}
	}
	fmt.Fprintln(f.Writer)
	return nil
}

var AccentStyle = lipgloss.NewStyle().Foreground(Accent)

func (f *TextFormatter) printTable(rows [][]string) {
	if len(rows) == 0 {
		return
	}
	// Compute column widths
	cols := len(rows[0])
	widths := make([]int, cols)
	for _, row := range rows {
		for i, c := range row {
			w := lipgloss.Width(c)
			if w > widths[i] {
				widths[i] = w
			}
		}
	}
	for _, row := range rows {
		for j, c := range row {
			if j == cols-1 {
				fmt.Fprint(f.Writer, c)
			} else {
				fmt.Fprintf(f.Writer, "%-*s  ", widths[j], c)
			}
		}
		fmt.Fprintln(f.Writer)
	}
	fmt.Fprintln(f.Writer)
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
