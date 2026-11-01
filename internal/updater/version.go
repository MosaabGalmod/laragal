package updater

import (
	"regexp"
	"strconv"
	"strings"
)

var semverRegex = regexp.MustCompile(`[vV]?(\d+\.\d+(?:\.\d+)*)`)
var phpVersionRegex = regexp.MustCompile(`(?i)PHP\s+(\d+\.\d+(?:\.\d+)*)`)

// CleanVersionString strips prefixes like 'v' or 'php' and whitespace
func CleanVersionString(v string) string {
	clean := strings.TrimSpace(v)
	clean = strings.TrimPrefix(clean, "v")
	clean = strings.TrimPrefix(clean, "V")
	return clean
}

// ExtractVersion searches text output for standard semantic version patterns
func ExtractVersion(output string) string {
	clean := strings.TrimSpace(output)
	if m := phpVersionRegex.FindStringSubmatch(clean); len(m) > 1 {
		return m[1]
	}
	matches := semverRegex.FindAllStringSubmatch(clean, -1)
	if len(matches) == 0 {
		return ""
	}

	for _, m := range matches {
		if len(m) > 1 {
			val := m[1]
			parts := strings.Split(val, ".")
			if len(parts) >= 2 {
				if _, err := strconv.Atoi(parts[0]); err == nil {
					if _, err2 := strconv.Atoi(parts[1]); err2 == nil {
						return val
					}
				}
			}
		}
	}

	return ""
}

func parseVersionComponents(v string) ([]int, int) {
	clean := CleanVersionString(v)
	build := 0
	if idx := strings.Index(clean, "-"); idx != -1 {
		suffix := clean[idx+1:]
		clean = clean[:idx]
		if b, err := strconv.Atoi(suffix); err == nil {
			build = b
		}
	}
	parts := strings.Split(clean, ".")
	var res []int
	for _, p := range parts {
		if num, err := strconv.Atoi(p); err == nil {
			res = append(res, num)
		}
	}
	return res, build
}

// IsVersionNewer compares two version strings and returns true if latest is strictly greater than current.
// Handles build suffixes such as "-1" correctly ("8.4.27" < "8.4.27-1" < "8.4.28").
func IsVersionNewer(current string, latest string) bool {
	cleanLatest := strings.TrimSpace(latest)
	if cleanLatest == "" {
		return false
	}
	cleanCurrent := strings.TrimSpace(current)
	if cleanCurrent == "" {
		return true
	}

	curParts, curBuild := parseVersionComponents(cleanCurrent)
	latParts, latBuild := parseVersionComponents(cleanLatest)

	maxLen := len(curParts)
	if len(latParts) > maxLen {
		maxLen = len(latParts)
	}

	for i := 0; i < maxLen; i++ {
		curVal := 0
		if i < len(curParts) {
			curVal = curParts[i]
		}
		latVal := 0
		if i < len(latParts) {
			latVal = latParts[i]
		}

		if latVal > curVal {
			return true
		}
		if latVal < curVal {
			return false
		}
	}

	return latBuild > curBuild
}
