package config

import (
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// LoadProtoYAMLConfig loads YAML configuration into a protobuf message.
func LoadProtoYAMLConfig(filePath string, target proto.Message) error {
	manager := NewConfigManager()
	return manager.LoadProtoYAMLConfig(filePath, target)
}

// LoadProtoJSONConfig loads JSON configuration into a protobuf message.
func LoadProtoJSONConfig(filePath string, target proto.Message) error {
	manager := NewConfigManager()
	return manager.LoadProtoJSONConfig(filePath, target)
}

// LoadProtoYAMLFromBytes loads YAML configuration bytes into a protobuf message.
func LoadProtoYAMLFromBytes(data []byte, target proto.Message) error {
	manager := NewConfigManager()
	return manager.LoadProtoYAMLConfigFromBytes(data, target)
}

// LoadProtoJSONFromBytes loads JSON configuration bytes into a protobuf message.
func LoadProtoJSONFromBytes(data []byte, target proto.Message) error {
	manager := NewConfigManager()
	return manager.LoadProtoJSONConfigFromBytes(data, target)
}

// LoadProtoYAMLConfig loads YAML configuration from file into a protobuf message.
func (cm *ConfigManager) LoadProtoYAMLConfig(filePath string, target proto.Message) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}
	if err := cm.loadProtoYAMLConfigFromBytes(data, target); err != nil {
		return err
	}
	cm.configs[filePath] = target
	return nil
}

// LoadProtoJSONConfig loads JSON configuration from file into a protobuf message.
func (cm *ConfigManager) LoadProtoJSONConfig(filePath string, target proto.Message) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}
	if err := cm.loadProtoJSONConfigFromBytes(data, target); err != nil {
		return err
	}
	cm.configs[filePath] = target
	return nil
}

// LoadProtoYAMLConfigFromBytes loads YAML configuration bytes into a protobuf message.
func (cm *ConfigManager) LoadProtoYAMLConfigFromBytes(data []byte, target proto.Message) error {
	return cm.loadProtoYAMLConfigFromBytes(data, target)
}

// LoadProtoJSONConfigFromBytes loads JSON configuration bytes into a protobuf message.
func (cm *ConfigManager) LoadProtoJSONConfigFromBytes(data []byte, target proto.Message) error {
	return cm.loadProtoJSONConfigFromBytes(data, target)
}

func (cm *ConfigManager) loadProtoYAMLConfigFromBytes(data []byte, target proto.Message) error {
	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}
	jsonData, err := json.Marshal(normalizeYAMLValue(raw))
	if err != nil {
		return fmt.Errorf("failed to convert YAML config to JSON: %w", err)
	}
	return cm.loadProtoJSONConfigFromBytes(jsonData, target)
}

func (cm *ConfigManager) loadProtoJSONConfigFromBytes(data []byte, target proto.Message) error {
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf config: %w", err)
	}
	return nil
}

func normalizeYAMLValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(val))
		for k, item := range val {
			m[k] = normalizeYAMLValue(item)
		}
		return m
	case map[any]any:
		m := make(map[string]any, len(val))
		for k, item := range val {
			m[fmt.Sprint(k)] = normalizeYAMLValue(item)
		}
		return m
	case []any:
		items := make([]any, len(val))
		for i, item := range val {
			items[i] = normalizeYAMLValue(item)
		}
		return items
	default:
		return val
	}
}
