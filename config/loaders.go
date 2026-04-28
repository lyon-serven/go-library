package config

import (
	"fmt"
	"os"
)

// LoadYAMLConfig loads YAML configuration from file
func LoadYAMLConfig(filePath string, config interface{}) error {
	manager := NewConfigManager()
	return manager.LoadConfig(filePath, config)
}

// LoadYAMLFromBytes loads YAML configuration from bytes
func LoadYAMLFromBytes(data []byte, config interface{}) error {
	manager := NewConfigManager()
	return manager.LoadConfigFromBytes(data, config)
}

// SaveYAMLConfig saves configuration to YAML file
func SaveYAMLConfig(filePath string, config interface{}) error {
	manager := NewConfigManager()
	return manager.SaveConfig(filePath, config)
}

// Environment variable substitution
func ExpandEnvVars(data []byte) []byte {
	content := string(data)
	result := os.ExpandEnv(content)
	return []byte(result)
}

// LoadYAMLConfigWithEnv loads YAML configuration with environment variable expansion
func LoadYAMLConfigWithEnv(filePath string, config interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	expandedData := ExpandEnvVars(data)
	return LoadYAMLFromBytes(expandedData, config)
}
