package config

import (
	"fmt"
	"path/filepath"
)

// ProfiledConfig creates a configuration manager with profile support
type ProfiledConfig struct {
	manager     *ConfigManager
	profilePath string
	profile     string
}

// NewProfiledConfig creates a configuration manager with profile support
func NewProfiledConfig(profilePath, profile string) *ProfiledConfig {
	return &ProfiledConfig{
		manager:     NewConfigManager(),
		profilePath: profilePath,
		profile:     profile,
	}
}

// LoadConfig loads configuration for the specified profile
func (pc *ProfiledConfig) LoadConfig(config interface{}) error {
	filePath := pc.buildProfilePath()

	if err := pc.manager.LoadConfig(filePath, config); err != nil {
		defaultPath := filepath.Join(pc.profilePath, "default.yaml")
		if err2 := pc.manager.LoadConfig(defaultPath, config); err2 != nil {
			return fmt.Errorf("failed to load config for profile %s: %w (also tried default: %v)",
				pc.profile, err, err2)
		}
	}

	return nil
}

// buildProfilePath builds the file path for the current profile
func (pc *ProfiledConfig) buildProfilePath() string {
	fileName := fmt.Sprintf("%s.yaml", pc.profile)
	return filepath.Join(pc.profilePath, fileName)
}
