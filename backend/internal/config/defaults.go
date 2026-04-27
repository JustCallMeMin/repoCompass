package config

// GetDefaults returns the base system default configuration.
// These represent the baseline fallback values when no other configuration is provided.
func GetDefaults() Config {
	maxFileSize := int64(1048576) // 1MB
	enableDefaultAnalyzers := true

	return Config{
		Excludes:               []string{".git", "node_modules"},
		MaxFileSizeBytes:       &maxFileSize,
		EnableDefaultAnalyzers: &enableDefaultAnalyzers,
	}
}
