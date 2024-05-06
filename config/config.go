package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

var (
	mu     sync.RWMutex
	config *Configuration
)

var _writeLock sync.Mutex

type Keybinds struct {
	Quicksave string `yaml:"quicksave" default:"-"`
}

type Configuration struct {
	Debug bool `yaml:"debug" default:"false"`
	Auto  bool `yaml:"auto" default:"false"`

	Keybinds Keybinds `yaml:"keybinds"`
}

func getConfigPath() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(path, "config.yaml")
}

func Init() error {
	file, err := os.ReadFile(getConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			config = &Configuration{}
			defaults.Set(config)
			WriteToDisk()
			return nil
		} else {
			return err
		}
	}

	if err := yaml.Unmarshal(file, &config); err != nil {
		return err
	}

	return nil
}

func Set(c *Configuration) {
	mu.Lock()
	config = c
	mu.Unlock()
}

func Get() *Configuration {
	mu.RLock()
	// Create a copy of the struct so that all modifications made beyond this
	// point are immutable.
	c := *config
	mu.RUnlock()
	return &c
}

func Update(callback func(c *Configuration)) {
	mu.Lock()
	callback(config)
	mu.Unlock()
}

func WriteToDisk() error {
	_writeLock.Lock()
	defer _writeLock.Unlock()

	ccopy := *config
	b, err := yaml.Marshal(&ccopy)
	if err != nil {
		return err
	}
	if err := os.WriteFile(getConfigPath(), b, 0o600); err != nil {
		return err
	}
	return nil
}
