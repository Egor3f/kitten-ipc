# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**kitten-ipc** is a cross-language IPC (inter-process communication) framework. It consists of two parts:

1. **`kitcom`** — a code generator CLI that parses annotated source files and generates IPC stub code for the other language
2. **`lib/`** — runtime IPC libraries (Go and TypeScript) that handle the actual communication over Unix domain sockets using newline-delimited JSON messages

The workflow: annotate structs/classes with `kittenipc:api` (Go) or `@kittenipc api` (TS), run `kitcom` to generate remote API stubs, then use the runtime library (`ParentIPC`/`ChildIPC`) to make cross-process calls.

## Build & Test Commands

```bash
# Go workspace covers: kitcom, lib/golang, example/golang
go build ./kitcom              # build the code generator
go test ./lib/golang/...       # run runtime library tests
go test ./kitcom/...           # run kitcom tests
go test ./...                  # run all tests

# Run a single test
go test ./lib/golang/... -run TestName

# Code generation (from example/)
cd example && make ipc

# kitcom usage
kitcom -src path/to/source -dest path/to/output [-pkg packageName]
# -pkg is required when generating Go output
```

## Architecture

### Code Generator (`kitcom/`)

- `main.go` — CLI entry point; selects parser/generator by file extension
- `internal/api/` — language-neutral API model (`Api` → `Endpoint` → `Method` → `Val` with types: int, string, bool, blob)
- `internal/golang/` — Go parser (`goparser.go`) reads `// kittenipc:api` annotations; Go generator (`gogen.go`) emits remote API stubs
- `internal/ts/` — TypeScript parser (`tsparser.go`) reads `@kittenipc api` JSDoc annotations; TS generator (`tsgen.go`) emits stubs
- `internal/tsgo/` — vendored subset of a Go-based TypeScript parser (scanner, AST, etc.)
- `internal/common/` — shared parser/writer utilities

### Runtime Library (`lib/`)

**Go** (`lib/golang/`):
- `parent.go` — `ParentIPC`: spawns a child process, listens on a Unix socket, accepts connection
- `child.go` — `ChildIPC`: connects to parent's socket (path passed via `--ipc-socket` arg)
- `common.go` — shared IPC logic: message send/receive, method dispatch via reflection, call tracking
- `serialize.go` — type serialization/deserialization (handles blob as base64)
- `protocol.go` — message format: `Message{Type, Id, Method, Args, Result, Error}`

**TypeScript** (`lib/ts/`): mirror of the Go runtime

### IPC Protocol

- Transport: Unix domain socket, newline-delimited JSON
- Two message types: `MsgCall` (1) and `MsgResponse` (2)
- Methods are addressed as `EndpointName.MethodName`
- Parent creates socket → starts child process → child connects back
- Bidirectional: both sides can call methods on the other

## Go Workspace

Uses `go.work` with three modules: `kitcom`, `lib/golang`, `example/golang`.
