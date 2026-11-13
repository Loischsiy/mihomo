# mihomo

Instructions for coding agents (Notion, Dust, Codex, Claude Code, OpenCode, Cursor and other MCP clients). They cover the whole tree; a nested AGENTS.md overrides this file inside its own subtree.

> Draft generated on 2026-11-08 from the repository layout (1047 files scanned). Verify every fact, replace the TODOs, delete whatever does not apply.

## Project overview

<h1 align="center">

## Tech stack

- Languages by file count: Go (982), JSON (17), Shell (6), YAML (6), Markdown (5), YAML (4).
- Key files: go.mod, Makefile, Dockerfile, README.md.

## Project structure

| Path | Contents |
| --- | --- |
| `transport/` | TODO: describe what lives here; 318 files |
| `component/` | TODO: describe what lives here; 172 files |
| `listener/` | TODO: describe what lives here; 150 files |
| `common/` | TODO: describe what lives here; 132 files |
| `adapter/` | TODO: describe what lives here; 70 files |
| `test/` | Tests; 39 files |
| `rules/` | TODO: describe what lives here; 35 files |
| `constant/` | TODO: describe what lives here; 32 files |
| `hub/` | TODO: describe what lives here; 23 files |
| `dns/` | TODO: describe what lives here; 22 files |
| `.github/` | CI/CD configuration; 14 files |
| `tunnel/` | TODO: describe what lives here; 7 files |
| `ntp/` | TODO: describe what lives here; 5 files |
| `config/` | Configuration; 4 files |
| `go.mod` | Go module |
| `Makefile` | Make targets |
| `Dockerfile` | Docker image |
| `README.md` | Human-facing documentation |

## Detailed guides

This file is the router. When a topic needs more than a few lines, move it to `.agents/<topic>.md` (architecture, backend, frontend, testing, workflows/deploy) and leave a one-line pointer here.

## Commands

```bash
go build ./...
go test ./...
```

Run servers and long builds as background jobs (`start_job`), never as a blocking command.

## Code style

- Match the existing style, directory layout and import conventions; never reformat untouched code.
- Keep changes scoped to the task: no new dependencies, build tools or drive-by refactors.
- TODO: repository-specific rules only (naming, layering, forbidden APIs); skip advice any model already knows.

## Testing

- Run before finishing: go build ./..., go test ./....
- Add or update tests for behaviour you change, even when nobody asked.

## Safety and permissions

- Allowed without asking: read and search files, run the checks listed above.
- Ask first: installing packages, `git push`, deleting files, touching CI, release or infrastructure config.
- Never commit secrets, `.env` files or local build artifacts.

## Definition of done

- [ ] Change stays inside the task and unrelated edits are untouched.
- [ ] Checks pass: go build ./..., go test ./....
- [ ] Tests and docs updated when behaviour changed.
- [ ] Full diff reviewed, including new files.

## Maintaining this file

AGENTS.md is project memory, not a one-off document. Treat it like code.

- Update this file in the same change that alters structure, commands, stack or invariants, so the map never goes stale.
- Use `update_agents_md` for edits (it verifies sha256) and `init_agents_md` for a new subdirectory.
- Keep this file a router: when a section outgrows a few lines of detail, move it to `.agents/<topic>.md` and link it from "Detailed guides".
- Write instruction files in English, in plain Markdown, with no frontmatter: agents re-read them on every request, and English costs the fewest tokens.
- Keep the file under ~150 lines. Link to docs instead of duplicating them, and drop anything a competent model already knows.
- Prefer short imperative rules, exact commands and small code snippets over prose paragraphs.
- Document only project-specific facts: layout, commands, conventions, invariants, boundaries.
- Use the conventional sections: project overview, tech stack, project structure, commands, code style, testing, safety and permissions.
- Put subtree-specific rules in a nested AGENTS.md; the file closest to the edited code wins.
- When a file outgrows the limit, keep it as a router: short sections plus one-line pointers to .agents/<topic>.md guides that are read only when the task touches them.
- Never store secrets, tokens or personal data in an instruction file.
