package config

// Config represents a source configuration (e.g., from a file, CLI, or defaults).
// All fields use pointers or slices so that unset values can be distinguished from zero values.
type Config struct {
	// Excludes specifies glob patterns of files/directories to ignore.
	// We use the slice itself to determine presence (nil vs empty vs populated).
	Excludes []string

	// MaxFileSizeBytes specifies the maximum size of a file to be analyzed.
	MaxFileSizeBytes *int64

	// EnableDefaultAnalyzers determines whether built-in analyzers should run.
	EnableDefaultAnalyzers *bool
}

// EffectiveConfiguration represents the final resolved configuration for a scan.
// All values are concrete primitives.
type EffectiveConfiguration struct {
	Excludes               []string
	MaxFileSizeBytes       int64
	EnableDefaultAnalyzers bool
}

// Resolve merges a slice of Config objects into a single EffectiveConfiguration.
// The precedence rule is simple: "Replacement" (higher index in the slice overrides lower index).
// For example, if configs = [defaults, file, cli], then cli overrides file, which overrides defaults.
func Resolve(configs []Config) EffectiveConfiguration {
	var effective EffectiveConfiguration

	for _, cfg := range configs {
		if cfg.Excludes != nil {
			// For slices, we replace the entire slice (higher wins) rather than appending.
			effective.Excludes = cfg.Excludes
		}
		if cfg.MaxFileSizeBytes != nil {
			effective.MaxFileSizeBytes = *cfg.MaxFileSizeBytes
		}
		if cfg.EnableDefaultAnalyzers != nil {
			effective.EnableDefaultAnalyzers = *cfg.EnableDefaultAnalyzers
		}
	}

	return effective
}
