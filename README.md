# Text Fill Action

This GitHub Action replaces placeholder strings in your project files with the full contents of matching files from a source directory.

Each file in `source_dir` defines one replacement: the filename without its extension is the search string, and the file contents are the replacement text.

This was heavily inspired by [gha-find-replace made by jacobtomlinson](https://github.com/jacobtomlinson/gha-find-replace).

## Usage

### Example workflow

```yaml
name: Fill placeholders
on: [push]
jobs:
  fill:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Fill text from snippets
        uses: Zisomerism/replace-from-file@v1
        with:
          source_dir: "content/snippets"
          include: "**/*.md"
          exclude: ".git/**"
```

If `content/snippets/header.txt` contains `# My Title`, every literal `header` token in matched `.md` files is replaced with that full block.

### Inputs

| Input | Required | Default | Description |
| --- | --- | --- | --- |
| `source_dir` | yes | — | Directory whose file names (without extension) are search strings and whose contents are replacements |
| `include` | no | `**` | Glob of repo files to search |
| `exclude` | no | `.git/**` | Glob of repo files to skip |

### Outputs

| Output | Description |
| --- | --- |
| `modifiedFiles` | Number of files written |

## How it works

1. Read each file in `source_dir` (direct children only).
2. Use the filename stem as the search string (e.g. `intro.txt` → `intro`).
3. Walk the repo and apply literal `strings.ReplaceAll` for every mapping.
4. Stems are applied longest-first so shorter tokens do not break longer ones.
5. Files under `source_dir` are always excluded from search targets.

## Design defaults

| Decision | Behavior |
| --- | --- |
| Match type | Literal string |
| `source_dir` depth | Flat (direct children only) |
| `source_dir` in search scope | Always excluded |
| Replacement scope | All occurrences in each matched file |
| Stem collision | Error (e.g. `foo.txt` and `foo.md`) |

## Development

```bash
go test -v ./...
```

Integration tests run via `.github/workflows/integration.yml` using the local `Dockerfile`.
