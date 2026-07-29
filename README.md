# Forge

[![Latest release](https://img.shields.io/github/v/release/Iwe-Coumou/forge)](https://github.com/Iwe-Coumou/forge/releases/latest)

Forge is a scaffolding CLI. Point it at a template and a project name, and it
generates a ready-to-build project — identifiers resolved, dependencies
tidied, formatted, and optionally committed to git.

Templates are organised by language, so a template is named
`<language>/<template>` — for example `go/cli_cobra`.

## Installation

```sh
go install github.com/Iwe-Coumou/forge@latest
```

Requires Go 1.25+, and `git` on your `PATH` if you use `--git`.

## Quick start

```sh
# One-time setup: sets your default module base path
forge init

# See what templates are available
forge list

# Scaffold a new project
forge new go/cli_cobra myapp
```

This creates `./myapp` with a `go.mod`, `main.go`, and a Cobra root command
wired up, then runs `go mod tidy` and `gofmt -w` on the result.

## Commands

### `forge init`

Creates Forge's config file, used to resolve the Go module path for
scaffolded projects.

```sh
forge init            # interactive prompts
forge init --default  # skip prompts, use defaults
```

The config lives at `<user config dir>/forge/config.yaml` (e.g.
`~/.config/forge/config.yaml` on Linux, `%AppData%\forge\config.yaml` on
Windows) and currently holds a single field:

```yaml
base_module: github.com/you
```

### `forge list`

Lists the available templates by their `<language>/<template>` id.

```sh
forge list      # id and short description
forge list -v   # id and full description
```

Templates marked `(wip)` belong to a language that is registered but not yet
usable; `forge new` will refuse them and explain why.

### `forge inspect [language/template]`

Shows everything about a single template: its descriptions, the command used
to verify a generated project, and the file tree it will produce.

```sh
forge inspect go/cli_cobra
```

```
go/cli_cobra

  Cobra-based CLI application

  Scaffolds a CLI application using the Cobra framework, with a root
  command and one example subcommand already wired up.

  verify   go build ./...

  files
  ├── cmd
  │   ├── example.go
  │   └── root.go
  ├── go.mod
  └── main.go
```

The file list comes from the same traversal `forge new` uses, so it always
matches what would actually be generated.

### `forge new [language/template] [project-name]`

Scaffolds a new project from `language/template` into a directory named
`project-name`.

```sh
forge new go/cli_cobra myapp
forge new go/cli_cobra myapp --path ../projects   # scaffold elsewhere
forge new go/cli_cobra myapp --git                # git init + first commit
forge new go/cli_cobra myapp --set module_path=github.com/me/myapp
```

By default the project's module path is `<base_module>/<project-name>` if a
base module is configured, otherwise just `<project-name>`.

| Flag            | Description                                         |
| --------------- | --------------------------------------------------- |
| `-p, --path`    | Directory to scaffold into (defaults to the cwd)     |
| `-g, --git`     | Run `git init`, `git add -A`, and an initial commit  |
| `--set k=v`     | Override a language-specific value (repeatable)      |
| `-v, --verbose` | Print each file rendered and each command run        |

Values resolve in this order, first match winning:

1. `--set` on the command line
2. Forge's config file
3. The language's built-in default

Unknown `--set` keys are rejected rather than ignored, so a typo fails loudly.

| Language | Accepted `--set` keys      |
| -------- | -------------------------- |
| `go`     | `module_path`, `go_version` |
| `python` | `min_python`                |

## Available templates

| Template          | Description                                                                                       |
| ----------------- | ------------------------------------------------------------------------------------------------- |
| `go/cli_cobra`    | Cobra-based CLI application with a root command and one example subcommand, plus a pinned `go.mod`. |
| `python/cli`      | Minimal Python CLI. Work in progress — not yet usable.                                              |

## Writing a template

Templates live under `internal/forger/templates/<language>/<name>/` and are
embedded into the binary at build time. A template is:

- A `template.yaml` declaring its `language`, plus `short` and `long`
  descriptions (used by `forge list`):

  ```yaml
  language: go
  short: "Cobra-based CLI application"
  long: >
    A longer description, shown by `forge list -v`.
  ```

  The `language` must match the folder the template lives in, and must be a
  registered language — a test enforces both.

- Any number of files, laid out exactly as they should appear in the
  generated project. A `.tmpl` suffix is stripped on render.

Templates are rendered with Go's `text/template` against a per-language
context, so the fields available depend on the language. Every language
provides:

- `{{.Name}}` — the project name

Go templates additionally get:

- `{{.ModulePath}}` — the resolved module path
- `{{.GoVersion}}` — the Go version for `go.mod`

Python templates additionally get:

- `{{.DistName}}` — the project name with underscores normalised to hyphens
- `{{.ImportName}}` — the project name with hyphens normalised to underscores
- `{{.MinPython}}` — the minimum Python version

## Adding a language

Everything about a language lives in one file, `internal/forger/lang_<name>.go`.
Define a type implementing `Language`, register it in `init()`, and declare
its render context alongside it:

```go
type rustLang struct{}

func init() { RegisterLanguage(rustLang{}) }

func (rustLang) Name() string { return "rust" }

func (rustLang) Context(p *Project, cfg *config.Config) (any, error) { ... }

func (rustLang) PostProcess(dir string, verbose bool) error { ... }

// Run in the generated project to prove it is valid.
func (rustLang) VerifyCmd() []string { return []string{"cargo", "check"} }
```

Then add templates under `internal/forger/templates/rust/`. No other file
needs to change — the registry is populated at package init.

While a language is still being built out, implement the optional
`Unimplemented` interface to keep it registered but refuse to scaffold it:

```go
func (rustLang) NotImplementedReason() string { return "templates are stubs" }
```

Deleting that one method is all it takes to enable the language.

## License

MIT — see [LICENSE](LICENSE).
