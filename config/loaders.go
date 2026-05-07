package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"
)

// LoadConfig loads configuration by file extension and target type.
func LoadConfig(filePath string, target any) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	if msg, ok := target.(proto.Message); ok {
		switch ext {
		case ".yaml", ".yml", "":
			return LoadProtoYAMLConfig(filePath, msg)
		case ".json":
			return LoadProtoJSONConfig(filePath, msg)
		default:
			return fmt.Errorf("unsupported proto config file extension %s", ext)
		}
	}

	switch ext {
	case ".yaml", ".yml", "":
		return LoadYAMLConfig(filePath, target)
	case ".json":
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read config file %s: %w", filePath, err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported config file extension %s", ext)
	}
}

// LoadYAMLConfig loads YAML configuration from file.
func LoadYAMLConfig(filePath string, target any) error {
	manager := NewConfigManager()
	return manager.LoadConfig(filePath, target)
}

// LoadYAMLFromBytes loads YAML configuration from bytes.
func LoadYAMLFromBytes(data []byte, target any) error {
	manager := NewConfigManager()
	return manager.LoadConfigFromBytes(data, target)
}

// SaveYAMLConfig saves configuration to YAML file.
func SaveYAMLConfig(filePath string, target any) error {
	manager := NewConfigManager()
	return manager.SaveConfig(filePath, target)
}

// Environment variable substitution
func ExpandEnvVars(data []byte) []byte {
	content := string(data)
	result := os.ExpandEnv(content)
	return []byte(result)
}

// LoadYAMLConfigWithEnv loads YAML configuration with environment variable expansion.
func LoadYAMLConfigWithEnv(filePath string, target any) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	expandedData := ExpandEnvVars(data)
	return LoadYAMLFromBytes(expandedData, target)
}
