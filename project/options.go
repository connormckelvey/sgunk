package project

type ProjectOption interface {
	Apply(*Project) error
}

type ProjectOptionFunc func(*Project) error

func (apply ProjectOptionFunc) Apply(p *Project) error {
	return apply(p)
}

func WithDir(dir string) ProjectOptionFunc {
	return func(p *Project) error {
		p.dir = dir
		return nil
	}
}

func WithExtensions(loaders ...ExtensionLoader) ProjectOptionFunc {
	return func(p *Project) error {
		for _, loader := range loaders {
			p.extensionLoaders[loader.Name()] = loader
		}
		return nil
	}
}
