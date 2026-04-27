package config

import (
	"reflect"
	"testing"
)

func TestGetDefaults(t *testing.T) {
	defaults := GetDefaults()

	if !reflect.DeepEqual(defaults.Excludes, []string{".git", "node_modules"}) {
		t.Fatalf("unexpected default excludes: got %v", defaults.Excludes)
	}
	if defaults.MaxFileSizeBytes == nil {
		t.Fatal("expected default MaxFileSizeBytes to be set")
	}
	if *defaults.MaxFileSizeBytes != 1048576 {
		t.Fatalf("unexpected default MaxFileSizeBytes: got %d", *defaults.MaxFileSizeBytes)
	}
	if defaults.EnableDefaultAnalyzers == nil {
		t.Fatal("expected default EnableDefaultAnalyzers to be set")
	}
	if !*defaults.EnableDefaultAnalyzers {
		t.Fatal("expected default EnableDefaultAnalyzers to be true")
	}
}

func TestResolve(t *testing.T) {
	defaultMaxFileSize := int64(1048576) // 1MB
	cliMaxFileSize := int64(2048576)     // 2MB

	defaultEnable := true
	cliEnable := false

	tests := []struct {
		name    string
		configs []Config
		want    EffectiveConfiguration
	}{
		{
			name:    "empty configs returns zero values",
			configs: []Config{},
			want: EffectiveConfiguration{
				Excludes:               nil,
				MaxFileSizeBytes:       0,
				EnableDefaultAnalyzers: false,
			},
		},
		{
			name: "single config",
			configs: []Config{
				{
					Excludes:               []string{"*.log"},
					MaxFileSizeBytes:       &defaultMaxFileSize,
					EnableDefaultAnalyzers: &defaultEnable,
				},
			},
			want: EffectiveConfiguration{
				Excludes:               []string{"*.log"},
				MaxFileSizeBytes:       defaultMaxFileSize,
				EnableDefaultAnalyzers: defaultEnable,
			},
		},
		{
			name: "higher index overrides lower index",
			configs: []Config{
				{
					// Defaults
					Excludes:               []string{"*.log"},
					MaxFileSizeBytes:       &defaultMaxFileSize,
					EnableDefaultAnalyzers: &defaultEnable,
				},
				{
					// CLI overrides
					Excludes:               []string{"vendor/*", "*.tmp"},
					MaxFileSizeBytes:       &cliMaxFileSize,
					EnableDefaultAnalyzers: &cliEnable,
				},
			},
			want: EffectiveConfiguration{
				Excludes:               []string{"vendor/*", "*.tmp"},
				MaxFileSizeBytes:       cliMaxFileSize,
				EnableDefaultAnalyzers: cliEnable,
			},
		},
		{
			name: "partial override leaves lower values intact",
			configs: []Config{
				{
					Excludes:               []string{"*.log"},
					MaxFileSizeBytes:       &defaultMaxFileSize,
					EnableDefaultAnalyzers: &defaultEnable,
				},
				{
					// Only override MaxFileSizeBytes
					MaxFileSizeBytes: &cliMaxFileSize,
				},
			},
			want: EffectiveConfiguration{
				Excludes:               []string{"*.log"},
				MaxFileSizeBytes:       cliMaxFileSize,
				EnableDefaultAnalyzers: defaultEnable,
			},
		},
		{
			name: "empty slice overrides populated slice",
			configs: []Config{
				{
					Excludes: []string{"*.log"},
				},
				{
					Excludes: []string{}, // Empty slice specifically set
				},
			},
			want: EffectiveConfiguration{
				Excludes: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.configs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resolve() = \n%v\nwant \n%v", got, tt.want)
			}
		})
	}
}
