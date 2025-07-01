package git

import (
	"fmt"
	"time"
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
	out := fmt.Sprintf(
		"%s (%s/%s) %s ago",
		version,
		branch,
		commit,
		time.Since(date).Round(time.Minute),
	)
	return out
}
