package sgunk

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type MiddlewareConfig struct {
	Path      string `yaml:"path"`
	Extension string `yaml:"extension"`
}

type DirConfig interface {
	GetDir() string
}

type SiteConfig struct {
	Dir string `yaml:"dir"`
}

func (c *SiteConfig) GetDir() string {
	return c.Dir
}

type ThemeConfig struct {
	Dir string `yaml:"dir"`
}

func (c *ThemeConfig) GetDir() string {
	return c.Dir
}

type BuildConfig struct {
	Dir string `yaml:"dir"`
}

func (c *BuildConfig) GetDir() string {
	return c.Dir
}

type ExtensionConfig struct {
	Name   string
	Config map[string]any
}

func (ex *ExtensionConfig) UnmarshalYAML(value *yaml.Node) error {
	var ext struct {
		Name string `yaml:"extension"`
	}
	if err := value.Decode(&ext); err != nil {
		return err
	}
	var config map[string]any
	if err := value.Decode(&config); err != nil {
		return err
	}
	delete(config, "extension")

	ex.Name = ext.Name
	ex.Config = config
	return nil
}

type ProjectConfig struct {
	Name  string            `yaml:"name"`
	Site  SiteConfig        `yaml:"site"`
	Theme ThemeConfig       `yaml:"theme"`
	Build BuildConfig       `yaml:"build"`
	Uses  []ExtensionConfig `yaml:"uses"`
}

var configFiles = map[string]func([]byte, any) error{
	"project.json": json.Unmarshal,
	"project.yml":  yaml.Unmarshal,
	"project.yaml": yaml.Unmarshal,
}

func LoadConfigFile(projectFS afero.Fs) (*ProjectConfig, error) {
	for name, unmarshal := range configFiles {
		b, err := afero.ReadFile(projectFS, name)
		if err != nil && os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var c ProjectConfig
		if err := unmarshal(b, &c); err != nil {
			return nil, err
		}
		return &c, nil
	}

	return nil, errors.New("config file not found")
}
