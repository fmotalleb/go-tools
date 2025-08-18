package git

import (
	"runtime"
	"time"

	"github.com/fmotalleb/go-tools/template"
)

var (
	Version = "v0.0.0-dev"
	Commit  = "--"
	Date    = "2025-06-21T15:24:40Z"
	Branch  = "dev-branch"
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
		"age":    time.Since(date).Round(time.Minute),
		"go": map[string]any{
			"ver":  runtime.Version(),
			"OS":   runtime.GOOS,
			"arch": runtime.GOARCH,
		},
	}
	out, err := template.EvaluateTemplate("{{ .ver }} ({{ .branch }}/{{ .hash }}) {{ .age }} ago built using {{ .go.ver }} for {{ .go.OS }} {{ .go.arch }}", data)
	if err != nil {
		panic(err)
	}
	return out
}
