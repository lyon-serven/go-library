// Package config provides utilities for loading and managing configuration from YAML files.
package config

import (
	"fmt"
	"github.com/lyon-serven/go-library/util/cryptoutil"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

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

	for _, opt := range opts {
		opt(cm)
	}

	return cm
}

// LoadConfig loads configuration from YAML file into the specified struct
func (cm *ConfigManager) LoadConfig(filePath string, config interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := validateConfigTarget(config); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	if cm.enableDecrypt {
		cm.decryptConfigFields(reflect.ValueOf(config))
	}

	cm.configs[filePath] = config
	return nil
}

// LoadConfigFromBytes loads configuration from YAML bytes into the specified struct
func (cm *ConfigManager) LoadConfigFromBytes(data []byte, config interface{}) error {
	if err := validateConfigTarget(config); err != nil {
		return err
	}

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
	if err := cm.LoadConfig(filePath, config); err != nil {
		return err
	}

	return cm.watcher.Watch(filePath, func() {
		if err := cm.ReloadConfig(filePath, config); err != nil {
			fmt.Printf("Error reloading config %s: %v\n", filePath, err)
		} else if callback != nil {
			callback()
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

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

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
			cm.decryptConfigFields(field)
		case reflect.Ptr:
			if !field.IsNil() {
				cm.decryptConfigFields(field)
			}
		case reflect.Slice, reflect.Array:
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.Kind() == reflect.Struct || elem.Kind() == reflect.Ptr {
					cm.decryptConfigFields(elem)
				}
			}
		}
	}
}
