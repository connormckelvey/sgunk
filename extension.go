package sgunk

type Extension interface {
	Name() string
	Register(p *Project, c map[string]any) error
}
