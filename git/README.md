# Git metadata

## ✅ Features

- Access your applications git metadata using buildtime variables.
- Parses and freezes injected values at init-time.

## 📦 Import

```go
import "github.com/fmotalleb/go-tools/git"
```

## 🛠️ Example GoReleaser Configuration

```yaml
version: 2

before:
  hooks:
    - go mod download

builds:
  - id: default
    env:
      - CGO_ENABLED=0
    ldflags:
      - "-s"
      - "-w"
      - "-X github.com/fmotalleb/go-tools/git.Version={{.Version}}"
      - "-X github.com/fmotalleb/go-tools/git.Commit={{.ShortCommit}}"
      - "-X github.com/fmotalleb/go-tools/git.Date={{.Date}}"
      - "-X github.com/fmotalleb/go-tools/git.Branch={{.Branch}}"
```

## 🛠️ Example Go Build Script

```bash
#!/usr/bin/env bash
set -euo pipefail

PACKAGE="github.com/fmotalleb/go-tools/git"

VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0-dev")
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BRANCH=$(git rev-parse --abbrev-ref HEAD)

go build -trimpath -ldflags "-s -w \
    -X ${PACKAGE}.Version=${VERSION} \
    -X ${PACKAGE}.Commit=${COMMIT} \
    -X ${PACKAGE}.Date=${DATE} \
    -X ${PACKAGE}.Branch=${BRANCH}" \
    -o "${OUTPUT}"
```

## 🔧 Provided Variables (Injected via `ldflags`)

| Variable  | Purpose                       | Default           |
| --------- | ----------------------------- | ----------------- |
| `Version` | Semantic version (latest tag) | `v0.0.0-dev`      |
| `Commit`  | Short git commit              | `--`              |
| `Date`    | Build date (RFC3339)          | `2025-06-21T...Z` |
| `Branch`  | Git branch name               | `dev-branch`      |

## 🧪 Example Usage

```go
fmt.Println(git.GetVersion())  // e.g. v1.2.3
fmt.Println(git.GetCommit())   // e.g. 4ac0ffee
fmt.Println(git.GetDate())     // time.Time object
fmt.Println(git.GetBranch())   // e.g. main
fmt.Println(git.String())      // v1.2.3 (main/4ac0ffee) 7m0s ago
```

> [!CAUTION]
> Do not modify public variables ins this module
>
> It will produce unpredictable consequences.
