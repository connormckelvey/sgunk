package util

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

type PageFrontMatter struct {
	Title    string            `yaml:"title" mapstructure:"title"`
	Meta     []*PageMetaValue  `yaml:"meta" mapstructure:"meta"`
	Links    []*PageLinksValue `yaml:"links" mapstructure:"links"`
	Template string            `yaml:"template" mapstructure:"template"`
}

type PageMetaValue struct {
	Title    *string `yaml:"title" mapstructure:"title"`
	Property *string `yaml:"property" mapstructure:"property"`
	Content  *string `yaml:"content" mapstructure:"content"`
	Name     *string `yaml:"name" mapstructure:"name"`
}

type PageLinksValue struct {
	Rel  *string `yaml:"rel" mapstructure:"rel"`
	Href *string `yaml:"href" mapstructure:"href"`
	Type *string `yaml:"type" mapstructure:"type"`
	Page *string `yaml:"page" mapstructure:"page"`
	As   *string `yaml:"as" mapstructure:"as"`
}

func TestMarshalMap(t *testing.T) {
	v, err := MarshalMap(&PageFrontMatter{Title: "foo", Template: "asd"})
	assert.NoError(t, err)

	spew.Dump(v)
}
