package mysql

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type DiscoveredDatabase struct {
	Type        string   `json:"type"`         // "schema_file" | "legacy_server_dir"
	SourcePath  string   `json:"source_path"`  // Absolute path to .sql file or data folder
	TargetDB    string   `json:"target_db"`    // Suggested normalized database name
	FileName    string   `json:"file_name"`    // Display name (e.g. schema.sql)
	SizeBytes   int64    `json:"size_bytes"`   // Size of the schema/dump file
	Description string   `json:"description"`  // Contextual description in Arabic & English
	FoundTables []string `json:"found_tables"` // Sample tables detected inside the schema
}

var createTableRegex = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?[\x60"']?([a-zA-Z0-9_]+)[\x60"']?`)

var skipDirs = map[string]bool{
	".git":         true,
	".github":      true,
	"node_modules": true,
	"vendor":       true,
	"storage":      true,
	"dist":         true,
	"build":        true,
	"bin":          true,
	".superpowers": true,
	".agents":      true,
}

// ExtractTableNamesFromSQL parses up to first 64KB of SQL content to identify table names
func ExtractTableNamesFromSQL(content string) []string {
	matches := createTableRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var tables []string
	for _, m := range matches {
		if len(m) > 1 {
			name := strings.TrimSpace(m[1])
			if name != "" && !seen[name] {
				seen[name] = true
				tables = append(tables, name)
			}
		}
	}
	return tables
}

// NormalizeTargetDB sanitizes a project name into a valid MySQL database identifier
func NormalizeTargetDB(rawName string) string {
	name := strings.TrimSpace(strings.ToLower(rawName))
	var sb strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}

	cleaned := sb.String()
	// Collapse multiple underscores
	re := regexp.MustCompile(`_+`)
	cleaned = re.ReplaceAllString(cleaned, "_")
	cleaned = strings.Trim(cleaned, "_")

	if cleaned == "" {
		cleaned = "app_db"
	}

	// MySQL database names cannot start with pure numeric if unquoted in some contexts
	if unicode.IsDigit(rune(cleaned[0])) {
		cleaned = "app_" + cleaned
	}

	return cleaned
}

// ScanProjectForSchemas recursively inspects a project directory for candidate .sql schema files
func ScanProjectForSchemas(projectPath string) ([]DiscoveredDatabase, error) {
	cleanRoot := filepath.Clean(projectPath)
	info, err := os.Stat(cleanRoot)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project path is not a directory: %s", projectPath)
	}

	suggestedDB := NormalizeTargetDB(filepath.Base(cleanRoot))
	var results []DiscoveredDatabase

	err = filepath.Walk(cleanRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip unreadable paths safely
		}

		// Calculate relative depth
		rel, err := filepath.Rel(cleanRoot, path)
		if err != nil {
			return nil
		}

		if info.IsDir() {
			base := filepath.Base(path)
			if skipDirs[strings.ToLower(base)] {
				return filepath.SkipDir
			}
			depth := strings.Count(rel, string(filepath.Separator))
			if depth > 4 {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for .sql file extensions
		if strings.EqualFold(filepath.Ext(path), ".sql") {
			fileName := filepath.Base(path)

			// Read up to 64KB to parse table names
			tables := readSampleTables(path)

			desc := fmt.Sprintf("Previous project database schema (%s) - %d tables", fileName, len(tables))

			results = append(results, DiscoveredDatabase{
				Type:        "schema_file",
				SourcePath:  filepath.ToSlash(path),
				TargetDB:    suggestedDB,
				FileName:    fileName,
				SizeBytes:   info.Size(),
				Description: desc,
				FoundTables: tables,
			})
		}
		return nil
	})

	return results, err
}

func readSampleTables(sqlPath string) []string {
	f, err := os.Open(sqlPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	buf := make([]byte, 64*1024)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil
	}

	return ExtractTableNamesFromSQL(string(buf[:n]))
}


var systemDatabases = map[string]bool{
	"mysql":              true,
	"information_schema": true,
	"performance_schema": true,
	"sys":                true,
	"#innodb_redo":       true,
	"#innodb_temp":       true,
}

var defaultHostDataDirs = []string{
	`C:\laragon\data\mysql`,
	`D:\laragon\data\mysql`,
	`C:\xampp\mysql\data`,
	`D:\xampp\mysql\data`,
	`C:\ProgramData\MySQL\MySQL Server 8.0\Data`,
	`C:\ProgramData\MySQL\MySQL Server 8.4\Data`,
	`C:\laragal\data\mysql`,
	`data\mysql`,
}

// scanHostDirsForDatabases searches given directories for candidate MySQL databases
func scanHostDirsForDatabases(dirs []string, targetDBHint string) []DiscoveredDatabase {
	var results []DiscoveredDatabase
	seen := make(map[string]bool)

	for _, hostDir := range dirs {
		entries, err := os.ReadDir(hostDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dbName := entry.Name()
			lowerName := strings.ToLower(dbName)

			if systemDatabases[lowerName] {
				continue
			}

			// If targetDBHint is provided, prioritize matching or related names
			if targetDBHint != "" {
				hint := strings.ToLower(targetDBHint)
				if !strings.Contains(lowerName, hint) && !strings.Contains(hint, lowerName) {
					continue
				}
			}

			fullPath := filepath.Join(hostDir, dbName)
			if seen[fullPath] {
				continue
			}
			seen[fullPath] = true

			desc := fmt.Sprintf("Previous database discovered on local machine (%s)", dbName)

			results = append(results, DiscoveredDatabase{
				Type:        "legacy_server_dir",
				SourcePath:  filepath.ToSlash(fullPath),
				TargetDB:    NormalizeTargetDB(dbName),
				FileName:    dbName,
				Description: desc,
			})
		}
	}

	return results
}

// ScanHostForLegacyDatabases scans standard host directories for databases matching or related to targetDBHint
func ScanHostForLegacyDatabases(targetDBHint string) []DiscoveredDatabase {
	return scanHostDirsForDatabases(defaultHostDataDirs, targetDBHint)
}

// ScanAllDatabases discovers both project-embedded schemas and matching host legacy databases
func ScanAllDatabases(projectPath string) ([]DiscoveredDatabase, error) {
	schemas, err := ScanProjectForSchemas(projectPath)
	if err != nil {
		return nil, err
	}

	cleanRoot := filepath.Clean(projectPath)
	hint := NormalizeTargetDB(filepath.Base(cleanRoot))

	hostDbs := ScanHostForLegacyDatabases(hint)
	return append(schemas, hostDbs...), nil
}
