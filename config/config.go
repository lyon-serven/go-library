// Package config provides utilities for loading and managing configuration from YAML files.
package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"gitee.com/wangsoft/go-library/util/cryptoutil"

	"gopkg.in/yaml.v3"
)

// ConfigManager manages configuration loading and caching
type ConfigManager struct {
	mu            sync.RWMutex
	configs       map[string]interface{}
	watcher       *FileWatcher
	enableDecrypt bool
	priKey        string // 秘钥
}

// ConfigOption defines the option function type
type ConfigOption func(*ConfigManager)

// WithDecryption enables configuration decryption with the provided key
func WithDecryption(priKey string) ConfigOption {
	return func(cm *ConfigManager) {
		cm.enableDecrypt = true
		cm.priKey = priKey
	}
}
func WithEnableDecryption() ConfigOption {
	return func(cm *ConfigManager) {
		cm.enableDecrypt = true
	}
}
func WithDisabledWatcher() ConfigOption {
	return func(cm *ConfigManager) {
		cm.watcher = nil
	}
}

// NewConfigManager creates a new configuration manager with options
func NewConfigManager(opts ...ConfigOption) *ConfigManager {
	cm := &ConfigManager{
		configs:       make(map[string]interface{}),
		watcher:       NewFileWatcher(),
		priKey:        "Lyon123!@#",
		enableDecrypt: false,
	}

	// Apply all options
	for _, opt := range opts {
		opt(cm)
	}

	return cm
}

// LoadConfig loads configuration from YAML file into the specified struct
func (cm *ConfigManager) LoadConfig(filePath string, config interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate that config is a pointer to a struct
	if err := validateConfigTarget(config); err != nil {
		return err
	}

	// Read YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Decrypt sensitive fields
	if cm.enableDecrypt {
		cm.decryptConfigFields(reflect.ValueOf(config))
	}
	// Cache the config
	cm.configs[filePath] = config

	return nil
}

// 递归解密函数
func (cm *ConfigManager) decryptConfigFields(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}
		switch field.Kind() {
		case reflect.String:
			// 处理string类型且以Des结尾的字段
			fieldName := fieldType.Name
			fieldValue := field.String()
			if strings.HasSuffix(fieldName, "Des") && fieldValue != "" {
				decryptedValue, err := cryptoutil.DESDecryptHex(fieldValue, cm.priKey)
				if err != nil {
					log.Printf("字段 %s 解密失败: %v", fieldName, err)
					continue
				}
				field.SetString(decryptedValue)
			}
		case reflect.Struct:
			// 递归处理嵌套结构体
			cm.decryptConfigFields(field)
		case reflect.Ptr:
			// 递归处理指针类型的嵌套结构体
			if !field.IsNil() {
				cm.decryptConfigFields(field)
			}
		case reflect.Slice, reflect.Array:
			// 处理切片或数组中的结构体元素
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.Kind() == reflect.Struct || elem.Kind() == reflect.Ptr {
					cm.decryptConfigFields(elem)
				}
			}
		}
	}
}

// LoadConfigFromBytes loads configuration from YAML bytes into the specified struct
func (cm *ConfigManager) LoadConfigFromBytes(data []byte, config interface{}) error {
	// Validate that config is a pointer to a struct
	if err := validateConfigTarget(config); err != nil {
		return err
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return nil
}

// LoadConfigFromReader loads configuration from a reader into the specified struct
func (cm *ConfigManager) LoadConfigFromReader(reader io.Reader, config interface{}) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read config data: %w", err)
	}

	return cm.LoadConfigFromBytes(data, config)
}

// ReloadConfig reloads configuration from file
func (cm *ConfigManager) ReloadConfig(filePath string, config interface{}) error {
	return cm.LoadConfig(filePath, config)
}

// WatchConfig watches a config file for changes and automatically reloads
func (cm *ConfigManager) WatchConfig(filePath string, config interface{}, callback func()) error {
	// Load initial config
	if err := cm.LoadConfig(filePath, config); err != nil {
		return err
	}

	// Start watching file
	return cm.watcher.Watch(filePath, func() {
		if err := cm.ReloadConfig(filePath, config); err != nil {
			fmt.Printf("Error reloading config %s: %v\n", filePath, err)
		} else {
			if callback != nil {
				callback()
			}
		}
	})
}

// GetCachedConfig retrieves cached configuration
func (cm *ConfigManager) GetCachedConfig(filePath string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	config, exists := cm.configs[filePath]
	return config, exists
}

// SaveConfig saves configuration struct to YAML file
func (cm *ConfigManager) SaveConfig(filePath string, config interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Update cache
	cm.mu.Lock()
	cm.configs[filePath] = config
	cm.mu.Unlock()

	return nil
}

// Close closes the configuration manager and stops file watchers
func (cm *ConfigManager) Close() error {
	if cm.watcher != nil {
		return cm.watcher.Close()
	}
	return nil
}

// Simple functions for direct usage without manager

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

// ConfigWithDefaults allows loading config with default values
type ConfigWithDefaults struct {
	config   interface{}
	defaults map[string]interface{}
}

// NewConfigWithDefaults creates a new config with defaults
func NewConfigWithDefaults(config interface{}) *ConfigWithDefaults {
	return &ConfigWithDefaults{
		config:   config,
		defaults: make(map[string]interface{}),
	}
}

// SetDefault sets a default value for a field path
func (cwd *ConfigWithDefaults) SetDefault(fieldPath string, value interface{}) *ConfigWithDefaults {
	cwd.defaults[fieldPath] = value
	return cwd
}

// Load loads the configuration and applies defaults
func (cwd *ConfigWithDefaults) Load(filePath string) error {
	// Load config first
	if err := LoadYAMLConfig(filePath, cwd.config); err != nil {
		return err
	}

	// Apply defaults
	return cwd.applyDefaults()
}

// applyDefaults applies default values to unset fields
func (cwd *ConfigWithDefaults) applyDefaults() error {
	configValue := reflect.ValueOf(cwd.config)
	if configValue.Kind() != reflect.Ptr || configValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	configValue = configValue.Elem()
	configType := configValue.Type()

	for fieldPath, defaultValue := range cwd.defaults {
		if err := setFieldDefault(configValue, configType, fieldPath, defaultValue); err != nil {
			return fmt.Errorf("failed to set default for %s: %w", fieldPath, err)
		}
	}

	return nil
}

// setFieldDefault sets a default value for a nested field
func setFieldDefault(configValue reflect.Value, configType reflect.Type, fieldPath string, defaultValue interface{}) error {
	parts := strings.Split(fieldPath, ".")
	currentValue := configValue
	currentType := configType

	// Navigate to the target field
	for i, part := range parts {
		field, found := currentType.FieldByName(part)
		if !found {
			return fmt.Errorf("field %s not found", part)
		}

		fieldValue := currentValue.FieldByName(part)
		if !fieldValue.CanSet() {
			return fmt.Errorf("field %s cannot be set", part)
		}

		// If this is the last part, set the value
		if i == len(parts)-1 {
			if fieldValue.IsZero() {
				defaultVal := reflect.ValueOf(defaultValue)
				if !defaultVal.Type().ConvertibleTo(fieldValue.Type()) {
					return fmt.Errorf("default value type %s is not convertible to field type %s",
						defaultVal.Type(), fieldValue.Type())
				}
				fieldValue.Set(defaultVal.Convert(fieldValue.Type()))
			}
			return nil
		}

		// Continue navigation for nested structs
		if field.Type.Kind() == reflect.Struct {
			currentValue = fieldValue
			currentType = field.Type
		} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			if fieldValue.IsNil() {
				// Create new instance for nil pointer
				newValue := reflect.New(field.Type.Elem())
				fieldValue.Set(newValue)
			}
			currentValue = fieldValue.Elem()
			currentType = field.Type.Elem()
		} else {
			return fmt.Errorf("cannot navigate through non-struct field %s", part)
		}
	}

	return nil
}

// validateConfigTarget validates that the target is a pointer to a struct
func validateConfigTarget(config interface{}) error {
	if config == nil {
		return fmt.Errorf("config target cannot be nil")
	}

	value := reflect.ValueOf(config)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("config target must be a pointer")
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("config target must be a pointer to a struct")
	}

	if !elem.CanSet() {
		return fmt.Errorf("config target must be settable")
	}

	return nil
}

// Environment variable substitution
func ExpandEnvVars(data []byte) []byte {
	content := string(data)

	// Simple environment variable expansion: ${VAR_NAME} or $VAR_NAME
	result := os.ExpandEnv(content)

	return []byte(result)
}

// LoadYAMLConfigWithEnv loads YAML configuration with environment variable expansion
func LoadYAMLConfigWithEnv(filePath string, config interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Expand environment variables
	expandedData := ExpandEnvVars(data)

	return LoadYAMLFromBytes(expandedData, config)
}

// Profile-based configuration
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
	// Build profile-specific config file path
	filePath := pc.buildProfilePath()

	// Try to load profile-specific config first
	if err := pc.manager.LoadConfig(filePath, config); err != nil {
		// Fall back to default config if profile-specific doesn't exist
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

// Validation interface for configuration structs
type Validator interface {
	Validate() error
}

// LoadYAMLConfigWithValidation loads YAML configuration and validates it
func LoadYAMLConfigWithValidation(filePath string, config interface{}) error {
	// Load config
	if err := LoadYAMLConfig(filePath, config); err != nil {
		return err
	}

	// Validate if config implements Validator interface
	if validator, ok := config.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}

	return nil
}

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
	// Validate that config is a pointer to a struct
	if err := validateConfigTarget(config); err != nil {
		return err
	}

	// Read YAML file directly
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Parse entire config into a map
	var fullConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &fullConfig); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Navigate to the section
	section, err := cs.getSection(fullConfig)
	if err != nil {
		return err
	}

	// Marshal section back to YAML and then unmarshal to target config
	sectionData, err := yaml.Marshal(section)
	if err != nil {
		return fmt.Errorf("failed to marshal section: %w", err)
	}

	// Parse YAML directly
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
