package output

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var stderr io.Writer = os.Stderr

var (
	green   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	red     = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	muted   = lipgloss.NewStyle().Faint(true)
	bold    = lipgloss.NewStyle().Bold(true)
	pending = lipgloss.NewStyle().Faint(true)
)

func Init() {
	if os.Getenv("NO_COLOR") != "" {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

func Writer() io.Writer {
	return stderr
}

func Successf(format string, args ...interface{}) {
	printf(green, "◆", format, args...)
}

func Errorf(format string, args ...interface{}) {
	printf(red, "◆", format, args...)
}

func Bulletf(format string, args ...interface{}) {
	printf(cyan, "◇", format, args...)
}

func Pendingf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(stderr, " %s %s\n", pending.Render("◆"), pending.Render(msg))
}

func Headerf(format string, args ...interface{}) {
	msg := bold.Render(fmt.Sprintf(format, args...))
	hline := muted.Render(repeat("─", 40))
	fmt.Fprintf(stderr, "\n %s %s %s\n", hline, msg, hline)
}

func Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(stderr, "%s%s\n", muted.Render("  · "), msg)
}

func printf(style lipgloss.Style, icon, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(stderr, " %s %s\n", style.Render(icon), style.Render(msg))
}

func repeat(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}

// Step tracks a timed operation.
type Step struct {
	label string
	start time.Time
}

func NewStep(label string) *Step {
	fmt.Fprintf(stderr, " %s %s\n", pending.Render("◆"), pending.Render(label))
	return &Step{label: label, start: time.Now()}
}

func (s *Step) Done() {
	s.done(green, "", "")
}

func (s *Step) Donef(detail string) {
	s.done(green, "  ", detail)
}

func (s *Step) Fail() {
	s.done(red, "", "")
}

func (s *Step) done(style lipgloss.Style, sep, detail string) {
	elapsed := time.Since(s.start)
	rest := s.label
	if detail != "" {
		rest += sep + muted.Render(detail)
	}
	rest += "  " + muted.Render(fmtDuration(elapsed))
	fmt.Fprintf(stderr, " %s %s\n", style.Render("◆"), muted.Render(rest))
}

func fmtDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return d.Round(10 * time.Millisecond).String()
	default:
		return d.Round(time.Second).String()
	}
}
