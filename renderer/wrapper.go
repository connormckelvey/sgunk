package renderer

import (
	"bytes"
	"log"
	"maps"

	"github.com/adrg/frontmatter"
	"github.com/spf13/afero"
)

func WrapTheme(themeFs afero.Fs, themeFile string, content []byte, props map[string]any) ([]byte, error) {
	log.Println("wrapping", string(content))
	file, err := themeFs.Open(themeFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var fm struct {
		Template string `yaml:"template"`
	}
	theme, err := frontmatter.Parse(file, &fm)
	if err != nil {
		return nil, err
	}

	ev := NewEvaluator(afero.NewIOFS(themeFs))

	var w bytes.Buffer
	newProps := maps.Clone(props)
	newProps["$outlet"] = string(content)
	if err := ev.Render(bytes.NewReader(theme), "", newProps, &w); err != nil {
		return nil, err
	}

	if fm.Template == "" {
		return w.Bytes(), nil
	}

	return WrapTheme(themeFs, fm.Template, w.Bytes(), props)
}
