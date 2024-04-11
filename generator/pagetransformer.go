package generator

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/connormckelvey/tmplrun"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type PageTransformer interface {
	Transform(project *Project, page *Page, src io.Reader, dst io.Writer) error
}

type PageTransformerFunc func(project *Project, page *Page, src io.Reader, dst io.Writer) error

func (transform PageTransformerFunc) Transform(project *Project, page *Page, src io.Reader, dst io.Writer) error {
	return transform(project, page, src, dst)
}

type PageFrontMatter struct {
	Title string   `yaml:"title" mapstructure:"title"`
	Tags  []string `yaml:"tags" mapstructure:"tags"`
}

func NewFrontMatterTransformer() PageTransformerFunc {
	return func(project *Project, page *Page, src io.Reader, dst io.Writer) error {
		var matter PageFrontMatter
		rest, err := frontmatter.Parse(src, &matter)
		if err != nil {
			return err
		}
		page.Metadata["title"] = matter.Title
		page.Metadata["tags"] = matter.Tags
		if _, err := dst.Write(rest); err != nil {
			return err
		}
		return nil
	}
}

func NewMarkdownTransformer() PageTransformerFunc {
	return func(project *Project, page *Page, src io.Reader, dst io.Writer) error {
		if page.Kind != PageKindMarkdown {
			_, err := io.Copy(dst, src)
			return err
		}
		md := goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
			),
			goldmark.WithRendererOptions(html.WithUnsafe()),
		)
		s, err := io.ReadAll(src)
		if err != nil {
			return err
		}
		return md.Convert(s, dst)
	}
}

func NewBlogPostTransformer() PageTransformerFunc {
	return func(project *Project, page *Page, src io.Reader, dst io.Writer) error {
		base := filepath.Base(page.SourcePath)
		parts := strings.Split(base, ".")
		if len(parts) < 3 {
			return nil
		}

		ms, err := strconv.ParseInt(parts[len(parts)-3], 10, 64)
		if err != nil {
			return err
		}

		page.Metadata["slug"] = parts[len(parts)-2]
		page.Metadata["publishTime"] = time.UnixMilli(ms).Format(time.RFC3339)

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
		return nil
	}
}

func NewTMPLRunTransformer() PageTransformerFunc {
	return func(project *Project, page *Page, src io.Reader, dst io.Writer) error {
		wd, _ := os.Getwd()
		tmpl := tmplrun.New(os.DirFS(wd))
		tmpPath := page.SourcePath + ".tmp"
		tmpFile, err := os.Create(tmpPath)
		if err != nil {
			return err
		}
		defer os.Remove(tmpPath)
		defer tmpFile.Close()

		if _, err := io.Copy(tmpFile, src); err != nil {
			return err
		}
		if err := tmpFile.Close(); err != nil {
			return err
		}

		err = tmpl.Render(dst, &tmplrun.RenderInput{
			Entrypoint: tmpPath,
			Props: map[string]any{
				"metadata": page.Metadata,
			},
		})
		if err != nil {
			return err
		}
		return nil
	}
}
