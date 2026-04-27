package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Resolver defines the interface for resolving scan configuration.
type Resolver interface {
	ResolveConfig(ctx context.Context, repoPath string, overrides Config) (EffectiveConfiguration, error)
}

// LocalConfigurationResolver implements Resolver for local directories.
type LocalConfigurationResolver struct{}

// NewLocalConfigurationResolver creates a new LocalConfigurationResolver.
func NewLocalConfigurationResolver() *LocalConfigurationResolver {
	return &LocalConfigurationResolver{}
}

// ResolveConfig reads the defaults, repository config file, and applies overrides.
func (r *LocalConfigurationResolver) ResolveConfig(ctx context.Context, repoPath string, overrides Config) (EffectiveConfiguration, error) {
	if ctx == nil {
		return EffectiveConfiguration{}, fmt.Errorf("context cannot be nil")
	}

	info, err := os.Stat(repoPath)
	if err != nil {
		return EffectiveConfiguration{}, fmt.Errorf("invalid repository path: %w", err)
	}
	if !info.IsDir() {
		return EffectiveConfiguration{}, fmt.Errorf("repository path must be a directory")
	}

	defaults := GetDefaults()
	
	// Check for repo configuration file
	var fileCfg Config
	var foundFile bool
	
	for _, filename := range []string{".repocompass.yaml", ".repocompass.yml"} {
		configPath := filepath.Join(repoPath, filename)
		data, err := os.ReadFile(configPath)
		if err == nil {
			// File exists, attempt to parse
			if err := yaml.Unmarshal(data, &fileCfg); err != nil {
				return EffectiveConfiguration{}, fmt.Errorf("failed to parse configuration file %s: %w", filename, err)
			}
			foundFile = true
			break
		}
		if !os.IsNotExist(err) {
			// Some other error like permission denied
			return EffectiveConfiguration{}, fmt.Errorf("error reading configuration file %s: %w", filename, err)
		}
	}

	configs := []Config{defaults}
	if foundFile {
		configs = append(configs, fileCfg)
	}
	configs = append(configs, overrides)

	return Resolve(configs), nil
}
