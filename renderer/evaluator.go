package renderer

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/connormckelvey/tmplrun/ast"
	"github.com/connormckelvey/tmplrun/evaluator"
	"github.com/connormckelvey/tmplrun/evaluator/driver"
	"github.com/connormckelvey/tmplrun/lexer"
	"github.com/connormckelvey/tmplrun/parser"
)

type Evaluator struct {
	fs fs.FS
}

func NewEvaluator(fsys fs.FS) *Evaluator {
	return &Evaluator{fsys}
}

// Render renders the template specified by the input and writes the result to the given writer.
func (ev *Evaluator) Render(source io.Reader, currentFile string, props map[string]any, w io.Writer) error {
	doc, err := ev.parse(source)
	if err != nil {
		return err
	}
	err = ev.render(w, currentFile, doc, props)
	if err != nil {
		return err
	}
	return nil
}

func (tr *Evaluator) parse(r io.Reader) (*ast.Document, error) {
	lex := lexer.New(r)
	par := parser.New(lex)
	return par.Parse()
}

func (tr *Evaluator) render(w io.Writer, currentFile string, doc *ast.Document, props map[string]any) error {
	hooks := &hooks{
		tr:          tr,
		currentFile: currentFile,
	}
	ev := evaluator.New(driver.NewGoja(), hooks)
	res, err := ev.Render(doc, evaluator.NewEnvironment(tr.fs, props, hooks))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(res))
	if err != nil {
		return err
	}

	return nil
}

type hooks struct {
	tr          *Evaluator
	currentFile string
}

func (th *hooks) resolve(name string) string {
	currentDir := filepath.Dir(th.currentFile)
	return filepath.Join(currentDir, name)
}

// Include resolves and includes the template specified by name.
func (th *hooks) Include(name string) (string, error) {
	rel := th.resolve(name)
	b, err := fs.ReadFile(th.tr.fs, rel)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Render resolves, parses, and renders the template specified by name with the given properties.
func (th *hooks) Render(name string, props map[string]any) (string, error) {
	rel := th.resolve(name)
	src, err := fs.ReadFile(th.tr.fs, rel)
	if err != nil {
		return "", err
	}

	doc, err := th.tr.parse(bytes.NewReader(src))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = th.tr.render(&buf, rel, doc, props)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
