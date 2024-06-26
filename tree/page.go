package tree

import "strings"

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

type PageNameParts struct {
	Raw   string
	Ext   string
	Kind  string
	Slug  string
	Extra []string
}

func GetEntryNameParts(name string) (PageNameParts, bool) {
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return PageNameParts{}, false
	}

	p := PageNameParts{
		Raw:  name,
		Ext:  parts[len(parts)-1],
		Kind: parts[0],
		Slug: parts[0],
	}

	if len(parts) > 2 {
		p.Slug = parts[len(parts)-2]
	}

	if len(parts) > 3 {
		p.Extra = parts[1 : len(parts)-2]
	}

	return p, true
}
