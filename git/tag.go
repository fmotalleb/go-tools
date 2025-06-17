package git

//go:generate bash -c "git describe --tags --abbrev=0 > latest.tag.tmp"

import (
	_ "embed"
)

//go:embed latest.tag.tmp
var LastTag string
