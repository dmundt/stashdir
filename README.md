# stashdir

[![CI](https://github.com/dmundt/stashdir/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/dmundt/stashdir/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmundt/stashdir.svg)](https://pkg.go.dev/github.com/dmundt/stashdir)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dmundt/stashdir)](https://github.com/dmundt/stashdir/blob/main/go.mod)
[![License: MIT](https://img.shields.io/github/license/dmundt/stashdir)](LICENSE)

A simple Go CLI to bookmark directories and quickly jump to them.

Features:
- add: store the current or given path
- copy: interactively copy a saved path to the clipboard, or by index
- list: show all stored paths
- path: print the absolute path of the database file
- remove: delete a stored path by index or by path
- select: interactively pick a path with arrow keys, or by index

Data is stored in a JSON file under your user config directory (on Windows: `%AppData%/stashdir/config.json`).

## Build / Install

Install the latest released binary to your `GOBIN`/`GOPATH/bin` (recommended):

```
go install github.com/dmundt/stashdir/cmd/stashdir@latest
```

From a local clone, you can also build or install:

```
go build -o stashdir.exe ./cmd/stashdir
go install ./cmd/stashdir
```

Then run `stashdir` from anywhere on your PATH.

## Usage

```cmd
stashdir add                # add current working directory
stashdir add C:\\Some\\Path # add a specific path
stashdir copy               # interactive selection; copies the chosen path to clipboard
stashdir copy 2             # copy by 1-based index
stashdir list               # list saved paths (persistently sorted)
stashdir path               # show absolute path to the database file
stashdir remove 2           # remove by index (1-based)
stashdir remove C:\\Some\\Path  # remove by exact path
stashdir select             # interactive selection; prints the chosen path
stashdir select 3           # select by 1-based index; prints the path
```

### Change directory in cmd.exe

Because a process cannot change its parent's working directory, either use the one-liner or optional helper:

```cmd
for /f "usebackq delims=" %i in (`stashdir select`) do cd /d "%i"
```

Optional helper script `stashdir-jump.cmd` is included for convenience.

On PowerShell:

```powershell
Set-Location (stashdir select)
```

## Notes

- Paths are persisted in alphabetical order.
- Selection is keyboard-navigable with arrow keys and Enter; ESC/Ctrl+C cancels quietly.
- Paths are normalized; duplicates won't be added.
- The tool prints paths to stdout so you can compose with your shell.

## CI

This repo includes a GitHub Actions workflow that runs on every push and pull request:

- OS matrix: `ubuntu-latest`, `windows-latest`
- Go versions: `1.21.x`
- Steps:
	- Build: `go build ./...`
	- Vet: `go vet ./...`
	- Test: `go test ./...`
	- Static analysis: `staticcheck ./...`

You can adjust the matrix to speed up CI (e.g., limit to the latest Go release) if needed.

## License

MIT License. See [LICENSE](LICENSE) for details.