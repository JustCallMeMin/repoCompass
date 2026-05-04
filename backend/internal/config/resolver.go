package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
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
		return EffectiveConfiguration{}, rcerr.New(rcerr.CodeInvalidSource,
			fmt.Sprintf("invalid repository path: %s", repoPath), err)
	}
	if !info.IsDir() {
		return EffectiveConfiguration{}, rcerr.New(rcerr.CodeInvalidSource,
			fmt.Sprintf("repository path must be a directory: %s", repoPath), nil)
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
				return EffectiveConfiguration{}, rcerr.New(rcerr.CodeConfigResolveFailed,
					fmt.Sprintf("failed to parse configuration file %s", filename), err)
			}
			foundFile = true
			break
		}
		if !os.IsNotExist(err) {
			return EffectiveConfiguration{}, rcerr.New(rcerr.CodeConfigResolveFailed,
				fmt.Sprintf("error reading configuration file %s", filename), err)
		}
	}

	configs := []Config{defaults}
	if foundFile {
		configs = append(configs, fileCfg)
	}
	configs = append(configs, overrides)

	return Resolve(configs), nil
}
