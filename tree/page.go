package tree

import "strings"

type PageFrontMatter struct {
	Title    string            `yaml:"title"`
	Meta     []*PageMetaValue  `yaml:"meta"`
	Links    []*PageLinksValue `yaml:"links"`
	Template string            `yaml:"template"`
}

type PageMetaValue struct {
	Title    *string `yaml:"title"`
	Property *string `yaml:"property"`
	Content  *string `yaml:"content"`
	Name     *string `yaml:"name"`
}

type PageLinksValue struct {
	Rel  *string `yaml:"rel"`
	Href *string `yaml:"href"`
	Type *string `yaml:"type"`
	Page *string `yaml:"page"`
	As   *string `yaml:"as"`
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
