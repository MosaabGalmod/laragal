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

// ExtractSeries extracts the major.minor series string from a version (e.g. "11.8.9" -> "11.8", "10.11.19" -> "10.11")
func ExtractSeries(v string) string {
	clean := CleanVersionString(v)
	parts := strings.Split(clean, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return clean
}

func parseVersionComponents(v string) ([]int, int, bool) {
	clean := CleanVersionString(v)
	build := 0
	hasBuild := false
	if idx := strings.Index(clean, "-"); idx != -1 {
		suffix := clean[idx+1:]
		clean = clean[:idx]
		if b, err := strconv.Atoi(suffix); err == nil {
			build = b
			hasBuild = true
		}
	}
	parts := strings.Split(clean, ".")
	var res []int
	for _, p := range parts {
		if num, err := strconv.Atoi(p); err == nil {
			res = append(res, num)
		}
	}
	return res, build, hasBuild
}

// IsVersionNewer compares two version strings and returns true if latest is strictly greater than current.
// If current has NO build suffix and feed latest has a build suffix and numeric parts are equal,
// they are treated as equal (not newer) because local CLI outputs (like `php -v`) omit Windows rebuild suffixes.
// Keep: "8.5.12-1" < "8.5.12-2" when both have suffixes; "8.5.12" < "8.5.13-1".
func IsVersionNewer(current string, latest string) bool {
	cleanLatest := strings.TrimSpace(latest)
	if cleanLatest == "" {
		return false
	}
	cleanCurrent := strings.TrimSpace(current)
	if cleanCurrent == "" {
		return true
	}

	curParts, curBuild, curHasBuild := parseVersionComponents(cleanCurrent)
	latParts, latBuild, latHasBuild := parseVersionComponents(cleanLatest)

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

	// Numeric parts are equal.
	// When current has NO build suffix, do not rank a suffixed feed version as newer.
	if !curHasBuild || !latHasBuild {
		return false
	}

	return latBuild > curBuild
}

// ExtractMajor returns the major version number from a version string (e.g. "4.8.1" -> 4).
func ExtractMajor(v string) (int, error) {
	parts, _, _ := parseVersionComponents(v)
	if len(parts) == 0 {
		return 0, strconv.ErrSyntax
	}
	return parts[0], nil
}

// IsMajorChange compares the major version of current (or catalog) vs target.
// Returns true if both versions have parsed majors and they differ.
func IsMajorChange(currentOrCatalog string, target string) bool {
	curMajor, err1 := ExtractMajor(currentOrCatalog)
	tgtMajor, err2 := ExtractMajor(target)
	if err1 != nil || err2 != nil {
		return false
	}
	return curMajor != tgtMajor
}

