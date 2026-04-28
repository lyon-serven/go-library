package config

import (
	"fmt"
	"reflect"
	"strings"
)

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
	if err := LoadYAMLConfig(filePath, cwd.config); err != nil {
		return err
	}

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

	for i, part := range parts {
		field, found := currentType.FieldByName(part)
		if !found {
			return fmt.Errorf("field %s not found", part)
		}

		fieldValue := currentValue.FieldByName(part)
		if !fieldValue.CanSet() {
			return fmt.Errorf("field %s cannot be set", part)
		}

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

		if field.Type.Kind() == reflect.Struct {
			currentValue = fieldValue
			currentType = field.Type
		} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			if fieldValue.IsNil() {
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
