package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	File string `toml:"file"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}

	// Environment variable takes lowest precedence (config and flags override)
	if env := os.Getenv("BK_FILE"); env != "" {
		cfg.File = env
	}

	// Try to load config file
	if path == "" {
		configDir, err := os.UserConfigDir()
		if err == nil {
			path = filepath.Join(configDir, "bk", "config.toml")
		}
	}

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, cfg); err != nil {
				return nil, err
			}
		}
	}

	// Expand environment variables ($HOME, ${HOME}, etc.)
	if cfg.File != "" {
		cfg.File = os.ExpandEnv(cfg.File)
	}

	// Expand ~ prefix (after env expansion, in case someone writes ~/ literally)
	if cfg.File != "" && cfg.File[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.File = filepath.Join(home, cfg.File[2:]) // skip "~/"
		}
	}

	return cfg, nil
}
