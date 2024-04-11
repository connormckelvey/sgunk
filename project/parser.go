package project

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/afero"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type WalkState int

const (
	WalkStop WalkState = iota
	WalkSkipChildren
	WalkConinue
)

func WalkDir(dir string, walkFunc func(path string, entry fs.DirEntry) (WalkState, error)) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		state, err := walkFunc(path, entry)
		if err != nil {
			return err
		}
		switch state {
		case WalkStop, WalkSkipChildren:
			return nil
		default:
		}

		if !entry.IsDir() {
			continue
		}
		if err := WalkDir(path, walkFunc); err != nil {
			return err
		}
	}
	return nil
}

type Node interface {
	Path() string
	IsDir() bool
	Children() []Node
	AppendChild(Node)
}

type BaseNode struct {
	path     string
	isDir    bool
	children []Node
}

func NewBaseNode(path string, isDir bool) BaseNode {
	return BaseNode{path: path, isDir: isDir}
}

func (s *BaseNode) IsDir() bool {
	return s.isDir
}

func (s *BaseNode) Path() string {
	return s.path
}

func (s *BaseNode) Children() []Node {
	return s.children
}

func (s *BaseNode) AppendChild(n Node) {
	s.children = append(s.children, n)
}

// Root node
type Site struct {
	BaseNode
}

type EntryParser interface {
	Test(path string, entry fs.FileInfo) (bool, error)
	// Open parses the current line and returns a result of parsing.
	//
	// Open must not parse beyond the current line.
	// If Open has been able to parse the current line, Open must advance a reader
	// position by consumed byte length.
	//
	// If Open has not been able to parse the current line, Open should returns
	// (nil, NoChildren). If Open has been able to parse the current line, Open
	// should returns a new Block node and returns HasChildren or NoChildren.
	// Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State)
	// Open(parent Node, path string) (Node, State)

	Parse(path string, entry fs.FileInfo) (Node, error)
}

// State represents parser's state.
// State is designed to use as a bit flag.
type State int

const (
	// None is a default value of the [State].
	None State = 1 << iota

	// Continue indicates parser can continue parsing.
	Continue

	// Close indicates parser cannot parse anymore.
	Close

	// HasChildren indicates parser may have child blocks.
	HasChildren

	// NoChildren indicates parser does not have child blocks.
	NoChildren
)

func Parse(fsys afero.Fs, dir string, root Node, parsers []EntryParser) error {
	entries, err := afero.ReadDir(fsys, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		// find parser
		var parser EntryParser
		for _, p := range parsers {
			ok, err := p.Test(path, entry)
			if err != nil {
				return err
			}
			if ok {
				parser = p
				break
			}

		}
		if parser == nil {
			log.Printf("no parser for '%s', skipping...", path)
			continue
		}

		n, err := parser.Parse(path, entry)
		if err != nil {
			return err
		}
		if n == nil {
			continue
		}
		root.AppendChild(n)

		if entry.IsDir() {
			if err := Parse(fsys, path, n, parsers); err != nil {
				return err
			}
		}
	}
	return nil
	// return WalkDir(dir, func(path string, entry fs.DirEntry) (WalkState, error) {

	// })
}

type DefaultPageParser struct {
}

func (pp *DefaultPageParser) Test(path string, entry fs.FileInfo) (bool, error) {
	return true, nil
}

type DefaultPage struct {
	BaseNode
	parts PageNameParts
}

type DefaultDir struct {
	BaseNode
}

func (pp *DefaultPageParser) Parse(path string, entry fs.FileInfo) (Node, error) {
	name := filepath.Base(path)

	if entry.IsDir() {
		return &DefaultDir{
			BaseNode: NewBaseNode(path, true),
		}, nil
	}
	parts, ok := GetEntryNameParts(name)
	if !ok {
		return nil, errors.New("shouldnt have passed Test")
	}

	return &DefaultPage{
		BaseNode: NewBaseNode(path, false),
		parts:    parts,
	}, nil

}

type BlogEntryParser struct {
	root string
}

func (pp *BlogEntryParser) Test(path string, entry fs.FileInfo) (bool, error) {
	hasPrefix := strings.HasPrefix(path, pp.root)
	if entry.IsDir() {
		return hasPrefix, nil
	}
	name := filepath.Base(path)
	parts, ok := GetEntryNameParts(name)
	if !ok {
		return false, nil
	}
	return hasPrefix && parts.Kind == "post", nil
}

type BlogNode struct {
	BaseNode
	root string
}

type BlogCollection struct {
	BaseNode
}

type BlogPost struct {
	BaseNode
	parts     PageNameParts
	createdAt time.Time
}

func (pp *BlogEntryParser) Parse(path string, entry fs.FileInfo) (Node, error) {
	if path == pp.root && entry.IsDir() {
		return &BlogNode{
			BaseNode: NewBaseNode(path, true),
			root:     pp.root,
		}, nil
	}
	if entry.IsDir() {
		return &BlogCollection{
			BaseNode: NewBaseNode(path, true),
		}, nil
	}
	name := filepath.Base(path)
	parts, _ := GetEntryNameParts(name)

	post := &BlogPost{
		BaseNode: NewBaseNode(path, false),
		parts:    parts,
	}
	if len(parts.Extra) > 0 {
		ms, err := strconv.ParseInt(parts.Extra[0], 10, 64)
		if err != nil {
			return nil, err
		}
		post.createdAt = time.UnixMilli(ms)
	}

	return post, nil

}

type Renderer interface {
	Test(Node) (bool, error)
	Props(node Node, context *RenderContext) (map[string]any, error)
	Open(node Node, context *RenderContext) error
	Close(node Node, context *RenderContext) error
}

func Render(siteFS afero.Fs, buildFS afero.Fs, root Node, renderers []Renderer, context *RenderContext) error {

	for _, renderer := range renderers {
		ok, err := renderer.Test(root)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		if err := renderer.Open(root, context); err != nil {
			return err
		}
		for _, child := range root.Children() {
			if err := Render(siteFS, buildFS, child, renderers, context); err != nil {
				return err
			}
		}

		if currentFile := context.CurrentFile(); !root.IsDir() && currentFile != nil {
			source, err := context.Source(root)
			if err != nil {
				return err
			}

			var fm struct {
				Page PageFrontMatter `yaml:"page"`
			}
			content, err := frontmatter.Parse(bytes.NewReader(source), &fm)
			if err != nil {
				return err
			}

			var metaProps []map[string]any
			for _, meta := range fm.Page.Meta {
				m, err := MarshalMap(meta)
				if err != nil {
					return err
				}
				metaProps = append(metaProps, m)
			}

			var linkProps []map[string]any
			for _, link := range fm.Page.Links {
				l, err := MarshalMap(link)
				if err != nil {
					return err
				}
				linkProps = append(linkProps, l)
			}
			props := map[string]any{
				"page": map[string]any{
					"meta":     metaProps,
					"links":    linkProps,
					"template": fm.Page.Template,
					"title":    fm.Page.Title,
				},
			}
			p, err := renderer.Props(root, context)
			if err != nil {
				return err
			}
			maps.Copy(props, p)

			var templated bytes.Buffer
			ev := NewEvaluator(afero.NewIOFS(siteFS))
			if err := ev.Render(bytes.NewReader(content), root.Path(), props, &templated); err != nil {
				return err
			}

			var compiledMarkdown bytes.Buffer
			md := goldmark.New(
				goldmark.WithExtensions(extension.GFM),
				goldmark.WithRendererOptions(html.WithUnsafe()),
			)
			if err := md.Convert(templated.Bytes(), &compiledMarkdown); err != nil {
				return err
			}
			spew.Dump(fm)
			if fm.Page.Template == "" {
				_, err := io.Copy(currentFile, &compiledMarkdown)
				return err
			}

			b, err := WrapTheme(context.themeFS, fm.Page.Template, compiledMarkdown.Bytes(), props)
			if err != nil {
				return err
			}

			if _, err := currentFile.Write(b); err != nil {
				return err
			}

			// if _, err := currentFile.Write(content); err != nil {
			// 	return err
			// }
		}

		if err := renderer.Close(root, context); err != nil {
			return err
		}
		break
	}
	return nil
}

type DefaultRenderer struct {
}

func (r *DefaultRenderer) Props(node Node, context *RenderContext) (map[string]any, error) {
	if node.IsDir() {
		return nil, nil
	}
	source, err := context.Source(node)
	if err != nil {
		return nil, err
	}

	var fm PageFrontMatter
	if _, err := frontmatter.Parse(bytes.NewReader(source), &fm); err != nil {
		return nil, err
	}
	return map[string]any{
		"title":    fm.Title,
		"meta":     fm.Meta,
		"links":    fm.Links,
		"template": fm.Template,
	}, nil
}

func (r *DefaultRenderer) Test(node Node) (bool, error) {
	switch node.(type) {
	case *DefaultDir, *DefaultPage, *Site:
		return true, nil
	}
	return false, nil
}

func (r *DefaultRenderer) openDefaultDir(node *DefaultDir, context *RenderContext) error {
	log.Println("render blog", node.Path())
	if err := context.MkdirAll(node.Path(), 0755); err != nil {
		return err
	}
	context.PushDir(node.Path())
	return nil
}

func (r *DefaultRenderer) closeDefaultDir(node *DefaultDir, context *RenderContext) error {
	context.PopDir(node.Path())
	return nil
}

func (r *DefaultRenderer) openDefaultPage(node *DefaultPage, context *RenderContext) error {
	log.Println("render post", node.path)

	_, err := context.CreateFile(node.parts.Slug + ".html")
	if err != nil {
		return err
	}
	return nil
}

func (r *DefaultRenderer) closeDefaultPage(node *DefaultPage, context *RenderContext) error {
	popped := context.PopFile()
	return popped.Close()
}

func (r *DefaultRenderer) Open(node Node, context *RenderContext) error {
	switch n := node.(type) {
	case *DefaultDir:
		log.Println("open dir", n.Path())
		return r.openDefaultDir(n, context)
	case *DefaultPage:
		log.Println("open page", n.Path())
		return r.openDefaultPage(n, context)
	}
	return nil
}

func (r *DefaultRenderer) Close(node Node, context *RenderContext) error {
	switch n := node.(type) {
	case *DefaultDir:
		log.Println("close dir", n.Path())
		return r.closeDefaultDir(n, context)
	case *DefaultPage:
		log.Println("close page", n.Path())
		return r.closeDefaultPage(n, context)
	}
	return nil
}

type BlogRenderer struct {
}

func (r *BlogRenderer) Test(node Node) (bool, error) {
	switch node.(type) {
	case *BlogNode, *BlogCollection, *BlogPost:
		return true, nil
	}
	return false, nil
}

func (r *BlogRenderer) Props(node Node, context *RenderContext) (map[string]any, error) {
	if node.IsDir() {
		return nil, nil
	}
	source, err := context.Source(node)
	if err != nil {
		return nil, err
	}

	var fm struct {
		Post BlogPostFrontMatter `yaml:"post"`
	}
	if _, err := frontmatter.Parse(bytes.NewReader(source), &fm); err != nil {
		return nil, err
	}
	return map[string]any{
		"post": map[string]any{
			"createdAt": node.(*BlogPost).createdAt,
			"title":     fm.Post.Title,
			"tags":      fm.Post.Tags,
		},
	}, nil
}

func (f *BlogRenderer) openBlogNode(node *BlogNode, context *RenderContext) error {
	log.Println("render blog", node.root)
	if err := context.MkdirAll(node.root, 0755); err != nil {
		return err
	}
	context.PushDir(node.root)
	return nil
}

func (*BlogRenderer) closeBlogNode(node *BlogNode, context *RenderContext) error {
	context.PopDir(node.root)
	return nil
}

func (f *BlogRenderer) renderBlogCollection(node *BlogCollection, context *RenderContext) error {
	log.Println("render collection", node.path)

	return nil
}

func (f *BlogRenderer) openBlogPost(node *BlogPost, context *RenderContext) error {
	log.Println("render post", node.path)

	datePath := time.Now().Format("2006/01/02")
	if err := context.MkdirAll(datePath, 0755); err != nil {
		return err
	}
	postPath := filepath.Join(datePath, node.parts.Slug+".html")
	_, err := context.CreateFile(postPath)
	if err != nil {
		return err
	}

	return nil
}

func (f *BlogRenderer) closeBlogPost(node *BlogPost, context *RenderContext) error {
	popped := context.PopFile()
	return popped.Close()
}

type PropsBuilder func(props map[string]any)

type RenderContext struct {
	dirstack  []string
	themeFS   afero.Fs
	buildFS   afero.Fs
	siteFS    afero.Fs
	openFiles []afero.File
	props     []func(last map[string]any) map[string]any
}

func (rc *RenderContext) Source(node Node) ([]byte, error) {
	return afero.ReadFile(rc.siteFS, node.Path())
}

func (rc *RenderContext) WorkDir() string {
	return filepath.Join(rc.dirstack...)
}

func (rc *RenderContext) MkdirAll(path string, perm fs.FileMode) error {
	return rc.buildFS.MkdirAll(
		filepath.Join(rc.WorkDir(), path),
		perm,
	)
}

func (rc *RenderContext) CurrentFile() afero.File {
	if len(rc.openFiles) == 0 {
		return nil
	}
	return rc.openFiles[len(rc.openFiles)-1]
}

func (rc *RenderContext) CreateFile(path string) (io.Writer, error) {
	file, err := rc.buildFS.Create(
		filepath.Join(rc.WorkDir(), path),
	)
	if err != nil {
		return nil, err
	}
	rc.openFiles = append(rc.openFiles, file)
	return file, nil
}

func (rc *RenderContext) PopFile() afero.File {
	popped := rc.openFiles[len(rc.openFiles)-1]
	rc.openFiles = rc.openFiles[0 : len(rc.openFiles)-1]
	return popped
}

func (rc *RenderContext) PushDir(dir string) {
	rc.dirstack = append(rc.dirstack, dir)
}

func (rc *RenderContext) PopDir(dir string) string {
	popped := rc.dirstack[len(rc.dirstack)-1]
	rc.dirstack = rc.dirstack[0 : len(rc.dirstack)-1]
	return popped
}

func (r *BlogRenderer) Open(node Node, context *RenderContext) error {
	// TODO handle entering == false for changing anything after children have been rendered
	switch n := node.(type) {
	case *BlogNode:
		return r.openBlogNode(n, context)
	case *BlogCollection:
		return nil
	case *BlogPost:
		return r.openBlogPost(n, context)
	}
	return nil
}

func (r *BlogRenderer) Close(node Node, context *RenderContext) error {
	// TODO handle entering == false for changing anything after children have been rendered
	switch n := node.(type) {
	case *BlogNode:
		return r.closeBlogNode(n, context)
	case *BlogCollection:
		return nil
	case *BlogPost:
		return r.closeBlogPost(n, context)
	}
	return nil
}
