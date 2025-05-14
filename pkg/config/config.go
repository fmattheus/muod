// Package config provides configuration management for MUOD
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigFileName is the name of the config file
	DefaultConfigFileName = "muod.yaml"
	// DefaultConfigDirName is the directory name under XDG_CONFIG_HOME
	DefaultConfigDirName = "muod"
)

// Debug flag to control debug output
var Debug bool

// debugPrint prints debug messages if debug mode is enabled
func debugPrint(format string, args ...interface{}) {
	if Debug {
		fmt.Fprintf(os.Stderr, "Debug: "+format+"\n", args...)
	}
}

// Config represents the application configuration
type Config struct {
	// Default timeout for ping requests
	DefaultTimeout time.Duration `yaml:"default_timeout"`
	
	// Whether to show timestamps by default
	ShowTimestamps bool `yaml:"show_timestamps"`
	
	// Default number of ping rounds (-1 for infinite)
	DefaultCount int `yaml:"default_count"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultTimeout: 5 * time.Second,
		ShowTimestamps: true,
		DefaultCount:   -1,
	}
}

// getConfigPath returns the path to the config file following XDG Base Directory Specification
func getConfigPath(customPath string) (string, error) {
	if customPath != "" {
		debugPrint("Using custom config path: %s", customPath)
		return customPath, nil
	}

	debugPrint("No custom config path provided, checking XDG_CONFIG_HOME")
	// Check XDG_CONFIG_HOME first
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		debugPrint("XDG_CONFIG_HOME not set, using ~/.config")
		// Default to ~/.config if XDG_CONFIG_HOME is not set
		home, err := os.UserHomeDir()
		if err != nil {
			debugPrint("Failed to get user home directory: %v", err)
			return "", fmt.Errorf("failed to get user home directory: %v", err)
		}
		configHome = filepath.Join(home, ".config")
	}
	debugPrint("Using config home: %s", configHome)

	// Create the config directory if it doesn't exist
	configDir := filepath.Join(configHome, DefaultConfigDirName)
	debugPrint("Using config directory: %s", configDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		debugPrint("Failed to create config directory: %v", err)
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, DefaultConfigFileName)
	debugPrint("Final config path: %s", configPath)
	return configPath, nil
}

// LoadConfig loads configuration from the specified file
// If no file is specified, it looks for config file in XDG standard directories
func LoadConfig(configPath string) (*Config, error) {
	debugPrint("Loading config, custom path provided: %v", configPath != "")
	
	path, err := getConfigPath(configPath)
	if err != nil {
		debugPrint("Failed to get config path: %v", err)
		return nil, err
	}

	cfg := DefaultConfig()
	debugPrint("Created default config: timeout=%v, timestamps=%v, count=%d", 
		cfg.DefaultTimeout, cfg.ShowTimestamps, cfg.DefaultCount)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			debugPrint("Config file does not exist at %s, using defaults", path)
			return cfg, nil // Return default config if file doesn't exist
		}
		debugPrint("Failed to read config file: %v", err)
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	debugPrint("Successfully read config file: %s", path)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		debugPrint("Failed to parse config file: %v", err)
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	debugPrint("Successfully loaded config: timeout=%v, timestamps=%v, count=%d",
		cfg.DefaultTimeout, cfg.ShowTimestamps, cfg.DefaultCount)
	return cfg, nil
}

// SaveConfig saves the configuration to the specified file
func SaveConfig(cfg *Config, configPath string) error {
	path, err := getConfigPath(configPath)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
} 