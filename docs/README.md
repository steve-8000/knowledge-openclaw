# docs/

This directory is the source-of-truth (SoT) for OpenClaw knowledge Markdown files.

Conventions:
- Put only real, reviewed documents here.
- Use YAML frontmatter (`title`, `doc_type`, `tags`, `owners`, `confidence`, `source`).
- Keep one topic per file.
- Use stable path-based keys (for example: `adr/001-something.md`).

Sync:
- `doc-sync` scans this directory and ingests changed files into ki-db.
- Deleting a file here does not auto-delete documents in ki-db; change status via API when needed.
