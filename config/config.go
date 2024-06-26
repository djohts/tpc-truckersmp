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

type Features struct {
	AttachTrailer bool `yaml:"attach_trailer" default:"true"`

	Refuel         bool `yaml:"refuel" default:"false"`
	RefuelRelative int  `yaml:"refuel_relative" default:"1"`

	Teleport bool `yaml:"teleport" default:"true"`
}

type Configuration struct {
	Debug bool `yaml:"debug" default:"false"`
	Auto  bool `yaml:"auto" default:"false"`

	Keybinds Keybinds `yaml:"keybinds"`
	Features Features `yaml:"features"`
}

func getConfigPath() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(path, "config.yaml")
}

func Init() error {
	config = &Configuration{}
	defaults.Set(config)
	if file, err := os.ReadFile(getConfigPath()); err != nil {
		if os.IsNotExist(err) {
		} else {
			return err
		}
	} else {
		if err := yaml.Unmarshal(file, &config); err != nil {
			return err
		}
	}
	WriteToDisk()

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
