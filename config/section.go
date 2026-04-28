package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigSection allows loading specific sections of a config file
type ConfigSection struct {
	sectionPath string
}

// NewConfigSection creates a new config section loader
func NewConfigSection(sectionPath string) *ConfigSection {
	return &ConfigSection{sectionPath: sectionPath}
}

// LoadSection loads a specific section from YAML config
func (cs *ConfigSection) LoadSection(filePath string, config interface{}) error {
	if err := validateConfigTarget(config); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var fullConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &fullConfig); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	section, err := cs.getSection(fullConfig)
	if err != nil {
		return err
	}

	sectionData, err := yaml.Marshal(section)
	if err != nil {
		return fmt.Errorf("failed to marshal section: %w", err)
	}

	if err := yaml.Unmarshal(sectionData, config); err != nil {
		return fmt.Errorf("failed to parse section YAML: %w", err)
	}

	return nil
}

// getSection navigates to the specified section in the config map
func (cs *ConfigSection) getSection(config map[string]interface{}) (interface{}, error) {
	parts := strings.Split(cs.sectionPath, ".")
	current := interface{}(config)

	for _, part := range parts {
		if currentMap, ok := current.(map[string]interface{}); ok {
			if value, exists := currentMap[part]; exists {
				current = value
			} else {
				return nil, fmt.Errorf("section path %s not found at part %s", cs.sectionPath, part)
			}
		} else {
			return nil, fmt.Errorf("cannot navigate through non-map value at %s", part)
		}
	}

	return current, nil
}
