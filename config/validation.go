package config

import (
	"fmt"
	"reflect"
)

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

// Validation interface for configuration structs
type Validator interface {
	Validate() error
}

// LoadYAMLConfigWithValidation loads YAML configuration and validates it
func LoadYAMLConfigWithValidation(filePath string, config interface{}) error {
	if err := LoadYAMLConfig(filePath, config); err != nil {
		return err
	}

	if validator, ok := config.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}

	return nil
}
