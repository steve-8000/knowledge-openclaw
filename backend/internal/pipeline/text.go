package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
)

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)

type Chunk struct {
	HeadingPath string
	Text        string
	TokenCount  int
	SHA256      []byte
}

func NormalizeWhitespace(input string) string {
	lines := strings.Split(input, "\n")
	clean := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.Join(strings.Fields(line), " ")
		if trimmed == "" {
			if len(clean) == 0 || clean[len(clean)-1] == "" {
				continue
			}
			clean = append(clean, "")
			continue
		}
		clean = append(clean, trimmed)
	}
	return strings.TrimSpace(strings.Join(clean, "\n"))
}

func ExtractMarkdownLinks(text string) []string {
	matches := markdownLinkPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	links := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		link := strings.TrimSpace(match[1])
		if link == "" {
			continue
		}
		if _, ok := seen[link]; ok {
			continue
		}
		seen[link] = struct{}{}
		links = append(links, link)
	}
	sort.Strings(links)
	return links
}

func ChunkByHeadings(text string, maxTokens int) []Chunk {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	if maxTokens <= 0 {
		maxTokens = 512
	}

	lines := strings.Split(text, "\n")
	h2 := ""
	h3 := ""
	currentPath := ""
	buf := make([]string, 0)
	sections := make([]Chunk, 0)

	flush := func() {
		content := strings.TrimSpace(strings.Join(buf, "\n"))
		buf = buf[:0]
		if content == "" {
			return
		}
		tokens := strings.Fields(content)
		for start := 0; start < len(tokens); start += maxTokens {
			end := start + maxTokens
			if end > len(tokens) {
				end = len(tokens)
			}
			chunkText := strings.Join(tokens[start:end], " ")
			sum := sha256.Sum256([]byte(chunkText))
			sections = append(sections, Chunk{
				HeadingPath: currentPath,
				Text:        chunkText,
				TokenCount:  end - start,
				SHA256:      sum[:],
			})
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "## "):
			flush()
			h2 = strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			h3 = ""
			currentPath = h2
			continue
		case strings.HasPrefix(trimmed, "### "):
			flush()
			h3 = strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			if h2 != "" {
				currentPath = h2 + " > " + h3
			} else {
				currentPath = h3
			}
			continue
		}
		buf = append(buf, line)
	}
	flush()
	return sections
}

func StableKeyFromLink(link string) string {
	raw := strings.TrimSpace(link)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return ""
		}
		if parsed.Host == "" {
			return ""
		}
		raw = parsed.Path
	}
	if strings.HasPrefix(raw, "doc:") {
		return strings.TrimSpace(strings.TrimPrefix(raw, "doc:"))
	}
	if strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "mailto:") {
		return ""
	}
	base := path.Base(strings.TrimSuffix(raw, "/"))
	if base == "." || base == "/" || base == "" {
		return ""
	}
	base = strings.TrimSuffix(base, path.Ext(base))
	return strings.TrimSpace(base)
}

func HashHex(data []byte) string {
	return hex.EncodeToString(data)
}

func SimilarityJaccard(a string, b string) float64 {
	aSet := toWordSet(a)
	bSet := toWordSet(b)
	if len(aSet) == 0 && len(bSet) == 0 {
		return 1
	}
	inter := 0
	union := make(map[string]struct{}, len(aSet)+len(bSet))
	for w := range aSet {
		union[w] = struct{}{}
		if _, ok := bSet[w]; ok {
			inter++
		}
	}
	for w := range bSet {
		union[w] = struct{}{}
	}
	return float64(inter) / float64(len(union))
}

func toWordSet(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, token := range strings.Fields(strings.ToLower(s)) {
		token = strings.Trim(token, ".,;:!?()[]{}\"'`")
		if token == "" {
			continue
		}
		out[token] = struct{}{}
	}
	return out
}
