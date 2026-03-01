docs/

This directory is the source-of-truth (SoT) for OpenClaw knowledge files.

Conventions:
- Store only real, reviewed documents.
- Use Markdown with YAML frontmatter when adding real docs.
- Keep one topic per file and stable path-based keys.

Sync note:
- doc-sync currently ingests only .md files.
- Deleting a file in docs/ does not auto-delete existing ki-db records.
