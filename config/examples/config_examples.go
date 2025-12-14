// Package config demonstrates usage examples for the configuration management system.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"mylib/config"
)

// Example configuration structs

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	Timeout      string `yaml:"timeout"`
}

// Validate implements the Validator interface
func (dc *DatabaseConfig) Validate() error {
	if dc.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if dc.Port <= 0 {
		return fmt.Errorf("database port must be positive")
	}
	if dc.Username == "" {
		return fmt.Errorf("database username is required")
	}
	return nil
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Address   string        `yaml:"address"`
	Port      int           `yaml:"port"`
	TLS       TLSConfig     `yaml:"tls"`
	Timeouts  TimeoutConfig `yaml:"timeouts"`
	RateLimit RateLimit     `yaml:"rate_limit"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// TimeoutConfig represents timeout configuration
type TimeoutConfig struct {
	Read  time.Duration `yaml:"read"`
	Write time.Duration `yaml:"write"`
	Idle  time.Duration `yaml:"idle"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	Enabled bool `yaml:"enabled"`
	RPS     int  `yaml:"rps"`
	Burst   int  `yaml:"burst"`
}

// AppConfig represents the complete application configuration
type AppConfig struct {
	App      AppInfo        `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
	Logging  LoggingConfig  `yaml:"logging"`
	Features FeatureFlags   `yaml:"features"`
}

// AppInfo represents application metadata
type AppInfo struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// FeatureFlags represents feature toggle configuration
type FeatureFlags struct {
	EnableMetrics bool            `yaml:"enable_metrics"`
	EnableTracing bool            `yaml:"enable_tracing"`
	EnableDebug   bool            `yaml:"enable_debug"`
	NewFeatures   map[string]bool `yaml:"new_features"`
}

// Validate implements the Validator interface for AppConfig
func (ac *AppConfig) Validate() error {
	if ac.App.Name == "" {
		return fmt.Errorf("app name is required")
	}
	if ac.App.Environment == "" {
		return fmt.Errorf("app environment is required")
	}
	return ac.Database.Validate()
}

func main() {
	fmt.Println("=== Config Package Examples ===\n")

	// Create config directory for examples
	configDir := "./examples/configs"
	os.MkdirAll(configDir, 0755)

	// Example 1: Basic Configuration Loading
	fmt.Println("1. Basic Configuration Loading")
	basicConfigExample(configDir)

	// Example 2: Configuration with Defaults
	fmt.Println("\n2. Configuration with Defaults")
	defaultsExample(configDir)

	// Example 3: Environment Variable Substitution
	fmt.Println("\n3. Environment Variable Substitution")
	envVarsExample(configDir)

	// Example 4: Profile-based Configuration
	fmt.Println("\n4. Profile-based Configuration")
	profilesExample(configDir)

	// Example 5: Configuration Validation
	fmt.Println("\n5. Configuration Validation")
	validationExample(configDir)

	// Example 6: Configuration Sections
	fmt.Println("\n6. Configuration Sections")
	sectionsExample(configDir)

	// Example 7: File Watching
	fmt.Println("\n7. File Watching")
	watchingExample(configDir)

	// Example 8: Configuration Manager
	fmt.Println("\n8. Configuration Manager")
	managerExample(configDir)

	fmt.Println("\n=== All Examples Completed ===")
}

// basicConfigExample demonstrates basic configuration loading
func basicConfigExample(configDir string) {
	// Create sample config file
	configFile := filepath.Join(configDir, "app.yaml")
	yamlContent := `app:
  name: "MyApplication"
  version: "1.0.0"
  environment: "development"
  debug: true

database:
  host: "localhost"
  port: 5432
  username: "admin"
  password: "password123"
  database: "myapp"
  max_open_conns: 10
  max_idle_conns: 5
  timeout: "30s"

server:
  address: "0.0.0.0"
  port: 8080
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  timeouts:
    read: "10s"
    write: "10s"
    idle: "60s"
  rate_limit:
    enabled: true
    rps: 1000
    burst: 100

logging:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: 100
  max_backups: 3
  max_age: 28

features:
  enable_metrics: true
  enable_tracing: false
  enable_debug: true
  new_features:
    feature_a: true
    feature_b: false
`

	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		log.Printf("Failed to create config file: %v", err)
		return
	}

	// Load configuration
	var appConfig AppConfig
	if err := config.LoadYAMLConfig(configFile, &appConfig); err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}

	fmt.Printf("Loaded config for app: %s v%s\n", appConfig.App.Name, appConfig.App.Version)
	fmt.Printf("Database: %s:%d\n", appConfig.Database.Host, appConfig.Database.Port)
	fmt.Printf("Server: %s:%d\n", appConfig.Server.Address, appConfig.Server.Port)
}

// defaultsExample demonstrates configuration with defaults
func defaultsExample(configDir string) {
	// Create minimal config file
	configFile := filepath.Join(configDir, "minimal.yaml")
	yamlContent := `app:
  name: "MinimalApp"
  environment: "production"

database:
  host: "prod-db.example.com"
  username: "prod_user"
`

	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		log.Printf("Failed to create minimal config file: %v", err)
		return
	}

	// Load config with defaults
	var appConfig AppConfig
	configWithDefaults := config.NewConfigWithDefaults(&appConfig).
		SetDefault("App.Version", "1.0.0").
		SetDefault("App.Debug", false).
		SetDefault("Database.Port", 5432).
		SetDefault("Database.Password", "default_password").
		SetDefault("Database.MaxOpenConns", 25).
		SetDefault("Database.MaxIdleConns", 10).
		SetDefault("Server.Port", 8080).
		SetDefault("Logging.Level", "info")

	if err := configWithDefaults.Load(configFile); err != nil {
		log.Printf("Failed to load config with defaults: %v", err)
		return
	}

	fmt.Printf("App: %s v%s (Debug: %v)\n", appConfig.App.Name, appConfig.App.Version, appConfig.App.Debug)
	fmt.Printf("Database port (default): %d\n", appConfig.Database.Port)
	fmt.Printf("Server port (default): %d\n", appConfig.Server.Port)
}

// envVarsExample demonstrates environment variable substitution
func envVarsExample(configDir string) {
	// Set environment variables
	os.Setenv("DB_HOST", "env-database.example.com")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_PASSWORD", "env_secret_password")
	os.Setenv("APP_DEBUG", "false")

	// Create config with environment variables
	configFile := filepath.Join(configDir, "env.yaml")
	yamlContent := `app:
  name: "EnvApp"
  version: "2.0.0"
  environment: "${APP_ENV:-development}"
  debug: false

database:
  host: "${DB_HOST}"
  port: 3306
  username: "${DB_USER:-admin}"
  password: "${DB_PASSWORD}"
  database: "myapp"
  max_open_conns: 10
  max_idle_conns: 5
  timeout: "30s"
`

	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		log.Printf("Failed to create env config file: %v", err)
		return
	}

	// Load configuration with environment variable expansion
	var appConfig AppConfig
	if err := config.LoadYAMLConfigWithEnv(configFile, &appConfig); err != nil {
		log.Printf("Failed to load config with env vars: %v", err)
		return
	}

	fmt.Printf("App debug from env: %v\n", appConfig.App.Debug)
	fmt.Printf("Database host from env: %s:%d\n", appConfig.Database.Host, appConfig.Database.Port)
	fmt.Printf("Database password from env: %s\n", appConfig.Database.Password)

	// Clean up environment variables
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("APP_DEBUG")
}

// profilesExample demonstrates profile-based configuration
func profilesExample(configDir string) {
	profileDir := filepath.Join(configDir, "profiles")
	os.MkdirAll(profileDir, 0755)

	// Create default profile
	defaultConfig := `app:
  name: "ProfileApp"
  version: "1.0.0"
  debug: false

database:
  host: "localhost"
  port: 5432
  max_open_conns: 10
`

	// Create development profile
	devConfig := `app:
  name: "ProfileApp"
  version: "1.0.0-dev"
  debug: true

database:
  host: "dev-db.local"
  port: 5432
  max_open_conns: 5
`

	// Create production profile
	prodConfig := `app:
  name: "ProfileApp"
  version: "1.0.0"
  debug: false

database:
  host: "prod-db.example.com"
  port: 5432
  max_open_conns: 50
`

	// Write profile configs
	os.WriteFile(filepath.Join(profileDir, "default.yaml"), []byte(defaultConfig), 0644)
	os.WriteFile(filepath.Join(profileDir, "development.yaml"), []byte(devConfig), 0644)
	os.WriteFile(filepath.Join(profileDir, "production.yaml"), []byte(prodConfig), 0644)

	// Test different profiles
	profiles := []string{"development", "production", "staging"} // staging doesn't exist, should fallback

	for _, profile := range profiles {
		var appConfig AppConfig
		profiledConfig := config.NewProfiledConfig(profileDir, profile)

		if err := profiledConfig.LoadConfig(&appConfig); err != nil {
			log.Printf("Failed to load profile %s: %v", profile, err)
			continue
		}

		fmt.Printf("Profile %s: %s v%s (Debug: %v, DB: %s, Conns: %d)\n",
			profile, appConfig.App.Name, appConfig.App.Version,
			appConfig.App.Debug, appConfig.Database.Host, appConfig.Database.MaxOpenConns)
	}
}

// validationExample demonstrates configuration validation
func validationExample(configDir string) {
	// Create invalid config
	configFile := filepath.Join(configDir, "invalid.yaml")
	yamlContent := `app:
  name: ""  # Invalid: empty name
  version: "1.0.0"
  environment: ""  # Invalid: empty environment

database:
  host: ""  # Invalid: empty host
  port: -1  # Invalid: negative port
  username: ""  # Invalid: empty username
`

	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		log.Printf("Failed to create invalid config file: %v", err)
		return
	}

	// Try to load and validate
	var appConfig AppConfig
	if err := config.LoadYAMLConfigWithValidation(configFile, &appConfig); err != nil {
		fmt.Printf("Validation failed (as expected): %v\n", err)
	} else {
		fmt.Println("Validation unexpectedly passed")
	}

	// Create valid config
	validConfigFile := filepath.Join(configDir, "valid.yaml")
	validYamlContent := `app:
  name: "ValidApp"
  version: "1.0.0"
  environment: "test"

database:
  host: "test-db.local"
  port: 5432
  username: "test_user"
  password: "test_password"
`

	if err := os.WriteFile(validConfigFile, []byte(validYamlContent), 0644); err != nil {
		log.Printf("Failed to create valid config file: %v", err)
		return
	}

	if err := config.LoadYAMLConfigWithValidation(validConfigFile, &appConfig); err != nil {
		fmt.Printf("Validation failed unexpectedly: %v\n", err)
	} else {
		fmt.Printf("Validation passed for app: %s\n", appConfig.App.Name)
	}
}

// sectionsExample demonstrates loading specific configuration sections
func sectionsExample(configDir string) {
	// Use the existing full config file
	configFile := filepath.Join(configDir, "app.yaml")

	// Load only database section
	var dbConfig DatabaseConfig
	dbSection := config.NewConfigSection("database")
	if err := dbSection.LoadSection(configFile, &dbConfig); err != nil {
		log.Printf("Failed to load database section: %v", err)
		return
	}

	fmt.Printf("Database section: %s:%d\n", dbConfig.Host, dbConfig.Port)

	// Load only server section
	var serverConfig ServerConfig
	serverSection := config.NewConfigSection("server")
	if err := serverSection.LoadSection(configFile, &serverConfig); err != nil {
		log.Printf("Failed to load server section: %v", err)
		return
	}

	fmt.Printf("Server section: %s:%d (TLS: %v)\n",
		serverConfig.Address, serverConfig.Port, serverConfig.TLS.Enabled)

	// Load nested section (server.tls)
	var tlsConfig TLSConfig
	tlsSection := config.NewConfigSection("server.tls")
	if err := tlsSection.LoadSection(configFile, &tlsConfig); err != nil {
		log.Printf("Failed to load TLS section: %v", err)
		return
	}

	fmt.Printf("TLS section: Enabled=%v\n", tlsConfig.Enabled)
}

// watchingExample demonstrates file watching
func watchingExample(configDir string) {
	// Create a config file for watching
	watchFile := filepath.Join(configDir, "watch.yaml")
	initialContent := `app:
  name: "WatchApp"
  version: "1.0.0"
  counter: 1
`

	if err := os.WriteFile(watchFile, []byte(initialContent), 0644); err != nil {
		log.Printf("Failed to create watch config file: %v", err)
		return
	}

	// Setup configuration manager
	manager := config.NewConfigManager()
	defer manager.Close()

	var appConfig struct {
		App struct {
			Name    string `yaml:"name"`
			Version string `yaml:"version"`
			Counter int    `yaml:"counter"`
		} `yaml:"app"`
	}

	// Watch configuration file
	changeCount := 0
	if err := manager.WatchConfig(watchFile, &appConfig, func() {
		changeCount++
		fmt.Printf("Config changed (count %d): %s v%s, counter=%d\n",
			changeCount, appConfig.App.Name, appConfig.App.Version, appConfig.App.Counter)
	}); err != nil {
		log.Printf("Failed to watch config file: %v", err)
		return
	}

	fmt.Printf("Initial config: %s v%s, counter=%d\n",
		appConfig.App.Name, appConfig.App.Version, appConfig.App.Counter)

	// Simulate config changes
	for i := 2; i <= 3; i++ {
		time.Sleep(500 * time.Millisecond)

		newContent := fmt.Sprintf(`app:
  name: "WatchApp"
  version: "1.0.0"
  counter: %d
`, i)

		if err := os.WriteFile(watchFile, []byte(newContent), 0644); err != nil {
			log.Printf("Failed to update watch config: %v", err)
			continue
		}

		// Give time for file watcher to detect changes
		time.Sleep(200 * time.Millisecond)
	}

	// Wait a bit more for final changes to be processed
	time.Sleep(300 * time.Millisecond)

	fmt.Printf("File watching completed with %d changes detected\n", changeCount)
}

// managerExample demonstrates advanced configuration manager usage
func managerExample(configDir string) {
	manager := config.NewConfigManager()
	defer manager.Close()

	// Load multiple configurations
	configs := map[string]string{
		"database": `
host: "manager-db.local"
port: 5432
username: "manager_user"`,
		"cache": `
type: "redis"
servers:
  - "cache1.local:6379"
  - "cache2.local:6379"
options:
  max_retries: 3`,
		"logging": `
level: "debug"
format: "text"
outputs:
  - "stdout"
  - "file:logs/app.log"`,
	}

	// Create and load each config
	loadedConfigs := make(map[string]interface{})
	for name, content := range configs {
		configFile := filepath.Join(configDir, fmt.Sprintf("manager_%s.yaml", name))

		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			log.Printf("Failed to create %s config: %v", name, err)
			continue
		}

		var config interface{}
		switch name {
		case "database":
			var dbConfig DatabaseConfig
			config = &dbConfig
		case "cache":
			var cacheConfig map[string]interface{}
			config = &cacheConfig
		case "logging":
			var logConfig map[string]interface{}
			config = &logConfig
		default:
			var genericConfig map[string]interface{}
			config = &genericConfig
		}

		if err := manager.LoadConfig(configFile, config); err != nil {
			log.Printf("Failed to load %s config: %v", name, err)
			continue
		}

		loadedConfigs[name] = config
		fmt.Printf("Loaded %s config successfully\n", name)
	}

	// Test cached config retrieval
	fmt.Printf("\nCached configs:\n")
	for name := range configs {
		configFile := filepath.Join(configDir, fmt.Sprintf("manager_%s.yaml", name))
		if cachedConfig, exists := manager.GetCachedConfig(configFile); exists {
			fmt.Printf("- %s: cached ✓\n", name)
			_ = cachedConfig // Use the cached config
		} else {
			fmt.Printf("- %s: not cached ✗\n", name)
		}
	}

	// Save a new configuration
	newConfig := DatabaseConfig{
		Host:         "saved-db.local",
		Port:         5432,
		Username:     "saved_user",
		Password:     "saved_password",
		Database:     "saved_db",
		MaxOpenConns: 20,
		MaxIdleConns: 10,
		Timeout:      "30s",
	}

	saveFile := filepath.Join(configDir, "saved_config.yaml")
	if err := manager.SaveConfig(saveFile, &newConfig); err != nil {
		log.Printf("Failed to save config: %v", err)
	} else {
		fmt.Printf("\nSaved config to %s\n", saveFile)

		// Verify saved config
		var loadedConfig DatabaseConfig
		if err := manager.LoadConfig(saveFile, &loadedConfig); err == nil {
			fmt.Printf("Verified saved config: %s:%d\n", loadedConfig.Host, loadedConfig.Port)
		}
	}
}
