package storage

import (
	"os/exec"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// GetGitAuthor runs `git config user.name` and returns the configured name.
// Returns empty string if git is not available or user.name is not configured.
func GetGitAuthor() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GenerateTitle creates a title from the prompt by taking approximately the first 60 characters,
// cleaning up whitespace, and truncating at a word boundary.
func GenerateTitle(prompt string) string {
	// normalize whitespace
	prompt = strings.TrimSpace(prompt)
	prompt = regexp.MustCompile(`\s+`).ReplaceAllString(prompt, " ")

	if len(prompt) <= 60 {
		return prompt
	}

	// truncate to ~60 chars at word boundary
	truncated := prompt[:60]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return truncated
}

// GenerateFilename creates a filename from title and timestamp in the format:
// YYYY-MM-DD-slugified-title.md
func GenerateFilename(title string, timestamp time.Time) string {
	datePrefix := timestamp.Format("2006-01-02")
	slug := Slugify(title)
	return datePrefix + "-" + slug + ".md"
}

// Slugify converts a string to a URL-friendly slug by:
// - converting to lowercase
// - replacing spaces with hyphens
// - removing non-alphanumeric characters (except hyphens)
// - collapsing multiple hyphens
// - trimming leading/trailing hyphens
// - limiting to ~50 characters
func Slugify(s string) string {
	// convert to lowercase
	s = strings.ToLower(s)

	// replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// remove non-alphanumeric characters except hyphens
	var builder strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			builder.WriteRune(r)
		}
	}
	s = builder.String()

	// collapse multiple hyphens
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")

	// trim leading/trailing hyphens
	s = strings.Trim(s, "-")

	// limit to ~50 characters at hyphen boundary
	if len(s) > 50 {
		truncated := s[:50]
		lastHyphen := strings.LastIndex(truncated, "-")
		if lastHyphen > 0 {
			s = truncated[:lastHyphen]
		} else {
			s = truncated
		}
	}

	return s
}

// GetTimestamp returns the current time.
func GetTimestamp() time.Time {
	return time.Now()
}
