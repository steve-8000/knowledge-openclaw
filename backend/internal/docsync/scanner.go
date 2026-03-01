package docsync

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter holds the YAML frontmatter metadata from a markdown document.
type Frontmatter struct {
	Title      string            `yaml:"title"`
	DocType    string            `yaml:"doc_type"`
	Tags       []string          `yaml:"tags"`
	Owners     []string          `yaml:"owners"`
	Confidence string            `yaml:"confidence"`
	Status     string            `yaml:"status"`
	Source     map[string]string `yaml:"source"`
}

// DocFile represents a parsed markdown document from the filesystem.
type DocFile struct {
	// StableKey derived from file path (e.g., "docs/adr/001-use-postgres")
	StableKey string
	// FilePath is the absolute path to the .md file
	FilePath string
	// Meta is the parsed YAML frontmatter
	Meta Frontmatter
	// Body is the raw markdown content (after frontmatter)
	Body string
	// ContentHash is SHA-256 of the full file content
	ContentHash string
}

// ScanDocs recursively scans a directory for .md files and parses them.
func ScanDocs(rootDir string) ([]DocFile, error) {
	var docs []DocFile

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		doc, err := parseDocFile(rootDir, path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		docs = append(docs, doc)
		return nil
	})

	return docs, err
}

func parseDocFile(rootDir, filePath string) (DocFile, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return DocFile{}, err
	}

	// Compute hash of entire file
	h := sha256.Sum256(raw)
	hash := hex.EncodeToString(h[:])

	// Derive stable_key from relative path (strip root and .md extension)
	relPath, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return DocFile{}, err
	}
	stableKey := strings.TrimSuffix(relPath, ".md")
	// Normalize path separators to forward slashes
	stableKey = filepath.ToSlash(stableKey)

	// Parse frontmatter
	meta, body, err := parseFrontmatter(string(raw))
	if err != nil {
		return DocFile{}, fmt.Errorf("frontmatter: %w", err)
	}

	return DocFile{
		StableKey:   stableKey,
		FilePath:    filePath,
		Meta:        meta,
		Body:        body,
		ContentHash: hash,
	}, nil
}

// parseFrontmatter splits YAML frontmatter from markdown body.
// Expects the file to start with "---\n" and have a closing "---\n".
func parseFrontmatter(content string) (Frontmatter, string, error) {
	var meta Frontmatter

	scanner := bufio.NewScanner(strings.NewReader(content))
	if !scanner.Scan() {
		return meta, content, nil
	}

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "---" {
		// No frontmatter — return entire content as body
		return meta, content, nil
	}

	// Collect YAML lines until closing "---"
	var yamlLines []string
	foundClose := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundClose = true
			break
		}
		yamlLines = append(yamlLines, line)
	}

	if !foundClose {
		return meta, content, fmt.Errorf("unclosed frontmatter (missing closing ---)")
	}

	// Parse YAML
	yamlContent := strings.Join(yamlLines, "\n")
	if err := yaml.Unmarshal([]byte(yamlContent), &meta); err != nil {
		return meta, "", fmt.Errorf("yaml parse: %w", err)
	}

	// Rest is body
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	body := strings.Join(bodyLines, "\n")
	body = strings.TrimLeft(body, "\n")

	return meta, body, nil
}
