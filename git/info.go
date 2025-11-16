package git

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/fmotalleb/go-tools/template"
)

var (
	Version = ""
	Commit  = ""
	Date    = ""
	Branch  = ""
	// ---
	version = ""
	commit  = ""
	date    time.Time
	branch  = ""
)

func freeze() {
	version = Version
	commit = Commit
	branch = Branch
	var err error
	date, err = time.Parse(time.RFC3339, Date)
	if err != nil {
		date = time.Time{}
	}
}

func init() {
	freeze()
}

func GetVersion() string {
	return version
}

func GetCommit() string {
	return commit
}

func GetDate() time.Time {
	return date
}

func GetBranch() string {
	return branch
}

func String() string {
	data := map[string]any{
		"ver":    version,
		"branch": branch,
		"hash":   commit,
		"age":    humanDuration(time.Since(date)),
		"date":   date.Format(time.RFC3339),
		"go": map[string]any{
			"ver":  runtime.Version(),
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
		},
	}
	tmpl := new(strings.Builder)
	if version != "" {
		tmpl.WriteString("{{ .ver }} ")
	} else {
		tmpl.WriteString("`go build-time args missing` ")
	}
	if branch != "" {
		tmpl.WriteString("({{ .branch }}/{{ .hash }}) ")
	}

	// somehow the default year is 1 and i believe nobody will build go apps before year 2
	// for those with time machine: dude why are u using this module?
	if time.Date(2, 0, 0, 0, 0, 0, 0, time.UTC).Before(date) {
		tmpl.WriteString("built {{ .age }} ago ({{ .date }}) ")
	}
	tmpl.WriteString("using {{ .go.ver }} for {{ .go.os }}/{{ .go.arch }}")
	out, err := template.EvaluateTemplate(tmpl.String(), data)
	if err != nil {
		panic(err)
	}
	return out
}

func humanDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	parts := []string{}
	if days := d / (24 * time.Hour); days > 0 {
		parts = append(parts, fmt.Sprintf("%d day%s", days, plural(days)))
		d %= 24 * time.Hour
	}
	if hours := d / time.Hour; hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hour%s", hours, plural(hours)))
		d %= time.Hour
	}
	if minutes := d / time.Minute; minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d min%s", minutes, plural(minutes)))
	}
	return strings.Join(parts, ", ")
}

func plural(n time.Duration) string {
	if n == 1 {
		return ""
	}
	return "s"
}
