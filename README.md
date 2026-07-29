# Forge

[![Latest release](https://img.shields.io/github/v/release/Iwe-Coumou/forge)](https://github.com/Iwe-Coumou/forge/releases/latest)

Forge is a scaffolding CLI for Go projects. Point it at a template and a
project name, and it generates a ready-to-build Go module — module path
resolved, dependencies tidied, formatted, and optionally committed to git.

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
forge new cli_cobra myapp
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

Lists the available templates.

```sh
forge list      # name and short description
forge list -v   # name and full description
```

### `forge new [template] [project-name]`

Scaffolds a new project from `template` into a directory named
`project-name`.

```sh
forge new cli_cobra myapp
forge new cli_cobra myapp --path ../projects   # scaffold elsewhere
forge new cli_cobra myapp --git                # git init + first commit
```

The project's module path is `<base_module>/<project-name>` if a base
module is configured, otherwise just `<project-name>`.

| Flag           | Description                                            |
| -------------- | ------------------------------------------------------- |
| `-p, --path`   | Directory to scaffold into (defaults to the cwd)         |
| `-g, --git`    | Run `git init`, `git add -A`, and an initial commit      |
| `-v, --verbose`| Print each file rendered and each command run            |

## Available templates

| Template    | Description                                                                                          |
| ----------- | ----------------------------------------------------------------------------------------------------- |
| `cli_cobra` | Cobra-based CLI application with a root command and one example subcommand, plus a pinned `go.mod`.    |

## Writing a template

Templates live under `internal/forger/templates/<name>/` and are embedded
into the binary at build time. A template is:

- A `template.yaml` with `short` and `long` descriptions (used by `forge list`).
- Any number of `*.go.tmpl` (or other) files, laid out exactly as they
  should appear in the generated project. The `.tmpl` suffix is stripped
  on render.

Templates are rendered with Go's `text/template` against a `Project`
struct, so files can reference:

- `{{.Name}}` — the project name
- `{{.ModulePath}}` — the resolved Go module path
- `{{.OutputDir}}` — the absolute output directory
- `{{.Template}}` — the template name

## License

MIT — see [LICENSE](LICENSE).
