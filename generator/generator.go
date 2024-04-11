package generator

type Generator struct {
	options []GeneratorOption
	project *Project
}

type GeneratorOption interface {
	Apply(g *Generator) error
}

type GeneratorOptionFunc func(g *Generator) error

func (apply GeneratorOptionFunc) Apply(g *Generator) error {
	return apply(g)
}

func New(opts ...GeneratorOption) *Generator {
	return &Generator{
		options: opts,
	}
}

func (g *Generator) applyOptions() error {
	for _, opt := range g.options {
		if err := opt.Apply(g); err != nil {
			return nil
		}
	}
	return nil
}

func WithProject(opts ...ProjectOption) GeneratorOptionFunc {
	return func(g *Generator) error {
		g.project = NewProject(opts...)
		return nil
	}
}
