package ssg

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type MiddlewareConfig struct {
	Path      string `yaml:"path"`
	Extension string `yaml:"extension"`
}

type SiteConfig struct {
	Dir string `yaml:"dir"`
}

type ThemeConfig struct {
	Dir string `yaml:"dir"`
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

type BuildConfig struct {
	Dir string `yaml:"dir"`
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
